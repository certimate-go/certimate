// A mock HTTP client for uniCloud.
package unicloud

import (
	"context"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/certimate-go/certimate/internal/app"
	xcrypto "github.com/certimate-go/certimate/pkg/utils/crypto"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

type Client struct {
	username string
	password string

	identityToken   string
	identityTokenAt time.Time
	identityTokenMu sync.Mutex

	consoleToken   string
	consoleTokenMu sync.Mutex

	rcForIdentity *resty.Client
	rcForConsole  *resty.Client
}

const (
	uniIdentityEndpoint     = "https://account.dcloud.net.cn/client"
	uniIdentityClientSecret = "ba461799-fde8-429f-8cc4-4b6d306e2339"
	uniIdentityAppId        = "__UNI__uniid_server"
	uniIdentitySpaceId      = "uni-id-server"
	uniConsoleEndpoint      = "https://unicloud.dcloud.net.cn/client"
	uniConsoleClientSecret  = "4c1f7fbf-c732-42b0-ab10-4634a8bbe834"
	uniConsoleAppId         = "__UNI__unicloud_console"
	uniConsoleSpaceId       = "dc-6nfabcn6ada8d3dd"
	uniPlatform             = "web"
)

const (
	challengeAlgRSAOAEP256 = "RSA-OAEP-256+A256GCM"
)

func NewClient(optFns ...OptionsFunc) (*Client, error) {
	opts := &Options{}
	for _, fn := range optFns {
		fn(opts)
	}

	if opts.Username == "" {
		return nil, fmt.Errorf("sdkerr: unset username")
	}
	if opts.Password == "" {
		return nil, fmt.Errorf("sdkerr: unset password")
	}

	client := &Client{
		username: opts.Username,
		password: opts.Password,
	}
	client.rcForIdentity = resty.New().
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		SetHeader("User-Agent", app.AppUserAgent)
	client.rcForConsole = resty.New().
		SetBaseURL("https://unicloud-api.dcloud.net.cn/unicloud/api").
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		SetHeader("User-Agent", app.AppUserAgent).
		SetPreRequestHook(func(_ *resty.Client, req *http.Request) error {
			if client.consoleToken != "" {
				req.Header.Set("Token", client.consoleToken)
			}

			return nil
		})

	return client, nil
}

func (c *Client) SetTimeout(timeout time.Duration) *Client {
	c.rcForIdentity.SetTimeout(timeout)
	return c
}

func (c *Client) sendIdentityRequest(ctx context.Context, endpoint, clientSecret, appId, spaceId, functionTarget, functionMethod, functionAction string, functionParams, functionData any) (*resty.Response, error) {
	if endpoint == "" {
		return nil, fmt.Errorf("sdkerr: bad request: endpoint cannot be empty")
	}

	payloadInfo, clientInfo, err := buildServerlessPayloadInfo(appId, spaceId, functionTarget, functionMethod, functionAction, functionParams, functionData, c.identityToken)
	if err != nil {
		return nil, fmt.Errorf("sdkerr: bad request: failed to build request: %w", err)
	}

	sign := generateSignature(payloadInfo, clientSecret)

	req := c.rcForIdentity.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Origin", "https://account.dcloud.net.cn").
		SetHeader("Referer", "https://account.dcloud.net.cn").
		SetHeader("X-Serverless-Sign", sign).
		SetBody(payloadInfo).
		SetContext(ctx)
	if clientInfo != nil {
		clientInfoBytes, _ := json.Marshal(clientInfo)
		req.SetHeader("X-Client-Info", string(clientInfoBytes))
	}
	if c.identityToken != "" {
		req.SetHeader("X-Client-Token", c.identityToken)
	}

	resp, err := req.Post(endpoint)
	if err != nil {
		return resp, fmt.Errorf("sdkerr: failed to send request: %w", err)
	} else if resp.IsError() {
		return resp, fmt.Errorf("sdkerr: unexpected status code: %d (resp: %s)", resp.StatusCode(), resp.String())
	}

	return resp, nil
}

func (c *Client) sendIdentityRequestWithResult(ctx context.Context, endpoint, clientSecret, appId, spaceId, functionTarget, functionMethod, functionAction string, functionParams, functionData any, result sdkResponse) error {
	resp, err := c.sendIdentityRequest(ctx, endpoint, clientSecret, appId, spaceId, functionTarget, functionMethod, functionAction, functionParams, functionData)
	if err != nil {
		if resp != nil {
			json.Unmarshal(resp.Body(), &result)
		}
		return err
	}

	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return fmt.Errorf("sdkerr: failed to unmarshal response: %w", err)
	} else if rSuccess := result.GetSuccess(); !rSuccess {
		return fmt.Errorf("sdkerr: api error: code='%s', message='%s'", result.GetErrorCode(), result.GetErrorMessage())
	}

	return nil
}

func (c *Client) sendConsoleRequest(ctx context.Context, method string, path string, params any) (*resty.Response, error) {
	req := c.rcForConsole.R().
		SetContext(ctx)
	if strings.EqualFold(method, http.MethodGet) {
		qs := make(map[string]string)
		if params != nil {
			temp := make(map[string]any)
			jsonb, _ := json.Marshal(params)
			json.Unmarshal(jsonb, &temp)
			for k, v := range temp {
				if v != nil {
					qs[k] = fmt.Sprintf("%v", v)
				}
			}
		}

		req = req.SetQueryParams(qs)
	} else {
		req = req.SetHeader("Content-Type", "application/json").SetBody(params)
	}

	resp, err := req.Execute(method, path)
	if err != nil {
		return resp, fmt.Errorf("sdkerr: failed to send request: %w", err)
	} else if resp.IsError() {
		return resp, fmt.Errorf("sdkerr: unexpected status code: %d (resp: %s)", resp.StatusCode(), resp.String())
	}

	return resp, nil
}

func (c *Client) sendConsoleRequestWithResult(ctx context.Context, method string, path string, params any, result sdkResponse) error {
	resp, err := c.sendConsoleRequest(ctx, method, path, params)
	if err != nil {
		if resp != nil {
			json.Unmarshal(resp.Body(), &result)
		}
		return err
	}

	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return fmt.Errorf("sdkerr: failed to unmarshal response: %w", err)
	} else if rReturnCode := result.GetReturnCode(); rReturnCode != 0 {
		return fmt.Errorf("sdkerr: api error: ret='%d', desc='%s'", rReturnCode, result.GetReturnDesc())
	}

	return nil
}

func (c *Client) ensureIdentityToken(ctx context.Context) error {
	c.identityTokenMu.Lock()
	defer c.identityTokenMu.Unlock()
	if c.identityToken != "" && c.identityTokenAt.After(time.Now()) {
		return nil
	}

	challangeId := ""
	challengeVer := 0
	challengeAlg := ""
	challengeKey := ""
	challengeKeyId := ""

	// 生成密码挑战
	{
		createPasswordChallengeParams := map[string]any{
			"action":           "login",
			"resetAppId":       uniConsoleAppId,
			"resetUniPlatform": uniPlatform,
		}

		createPasswordChallengeResp := &struct {
			sdkResponseBase
			Data *struct {
				ErrCode     json.RawMessage `json:"errCode,omitempty"`
				ErrMsg      string          `json:"errMsg,omitempty"`
				Version     int             `json:"version"`
				Algorithm   string          `json:"algorithm"`
				KeyId       string          `json:"keyId"`
				PublicKey   string          `json:"publicKey"`
				ChallengeId string          `json:"challengeId"`
				ExpiresAt   int64           `json:"expiresAt"`
			} `json:"data,omitempty"`
		}{}

		if err := c.sendIdentityRequestWithResult(ctx,
			uniIdentityEndpoint, uniIdentityClientSecret, uniIdentityAppId, uniIdentitySpaceId,
			"uni-id-co", "createPasswordChallenge", "", createPasswordChallengeParams, nil,
			createPasswordChallengeResp); err != nil {
			return err
		} else {
			if createPasswordChallengeResp.Data != nil && createPasswordChallengeResp.Data.ErrCode != nil {
				respErrCode := strings.Trim(string(createPasswordChallengeResp.Data.ErrCode), "\"")
				if respErrCode != "" && respErrCode != "0" {
					return fmt.Errorf("sdkerr: auth error: errCode='%s', errMsg='%s'", respErrCode, createPasswordChallengeResp.Data.ErrMsg)
				}
			}

			if createPasswordChallengeResp.Data == nil || createPasswordChallengeResp.Data.ChallengeId == "" {
				return fmt.Errorf("sdkerr: auth error: received empty challenge")
			}

			challangeId = createPasswordChallengeResp.Data.ChallengeId
			challengeVer = createPasswordChallengeResp.Data.Version
			challengeAlg = createPasswordChallengeResp.Data.Algorithm
			challengeKey = createPasswordChallengeResp.Data.PublicKey
			challengeKeyId = createPasswordChallengeResp.Data.KeyId
		}
	}

	// 登录
	{
		envelope := make(map[string]any)

		switch challengeAlg {
		case challengeAlgRSAOAEP256:
			envelopeInfo, err := buildPasswordEnvelopeWithRSAOAEP256AES256GCM(c.username, c.password, challangeId, challengeVer, challengeKey, challengeKeyId)
			if err != nil {
				return fmt.Errorf("sdkerr: auth error: failed to build password envelope: %w", err)
			}
			envelope = envelopeInfo
		default:
			return fmt.Errorf("sdkerr: auth error: unsupported challenge algorithm: %s", challengeAlg)
		}

		loginParams := map[string]any{
			"captcha":          "",
			"resetAppId":       uniConsoleAppId,
			"resetUniPlatform": uniPlatform,
			"isReturnToken":    false,
			"passwordEnvelope": envelope,
		}

		loginResp := &struct {
			sdkResponseBase
			Data *struct {
				ErrCode  json.RawMessage `json:"errCode,omitempty"`
				ErrMsg   string          `json:"errMsg,omitempty"`
				UID      string          `json:"uid"`
				NewToken *struct {
					Token        string `json:"token"`
					TokenExpired int64  `json:"tokenExpired"`
				} `json:"newToken,omitempty"`
			} `json:"data,omitempty"`
		}{}

		if err := c.sendIdentityRequestWithResult(ctx,
			uniIdentityEndpoint, uniIdentityClientSecret, uniIdentityAppId, uniIdentitySpaceId,
			"uni-id-co", "login", "", loginParams, nil,
			loginResp); err != nil {
			return err
		} else {
			if loginResp.Data != nil && loginResp.Data.ErrCode != nil {
				respErrCode := strings.Trim(string(loginResp.Data.ErrCode), "\"")
				if respErrCode != "" && respErrCode != "0" {
					return fmt.Errorf("sdkerr: auth error: errCode='%s', errMsg='%s'", respErrCode, loginResp.Data.ErrMsg)
				}
			}

			if loginResp.Data == nil || loginResp.Data.NewToken == nil || loginResp.Data.NewToken.Token == "" {
				return fmt.Errorf("sdkerr: auth error: received empty token")
			}

			tokenAt := time.UnixMilli(loginResp.Data.NewToken.TokenExpired)
			if tokenAt.IsZero() {
				return fmt.Errorf("sdkerr: auth error: received invalid token expiration")
			}

			c.identityToken = loginResp.Data.NewToken.Token
			c.identityTokenAt = tokenAt
		}
	}

	return nil
}

func (c *Client) ensureConsoleToken(ctx context.Context) error {
	if err := c.ensureIdentityToken(ctx); err != nil {
		return err
	}

	c.consoleTokenMu.Lock()
	defer c.consoleTokenMu.Unlock()
	if c.consoleToken != "" {
		return nil
	}

	type getUserTokenResponse struct {
		sdkResponseBase
		Data *struct {
			Code int32 `json:"code"`
			Data *struct {
				Result      int32  `json:"ret"`
				Description string `json:"desc"`
				Data        *struct {
					Email string `json:"email"`
					Token string `json:"token"`
				} `json:"data,omitempty"`
			} `json:"data,omitempty"`
		} `json:"data,omitempty"`
	}

	resp := &getUserTokenResponse{}
	if err := c.sendIdentityRequestWithResult(ctx,
		uniConsoleEndpoint, uniConsoleClientSecret, uniConsoleAppId, uniConsoleSpaceId,
		"uni-cloud-kernel", "", "user/getUserToken", nil, map[string]any{"isLogin": true},
		resp); err != nil {
		return err
	} else {
		if resp.Data == nil || resp.Data.Data == nil || resp.Data.Data.Data == nil || resp.Data.Data.Data.Token == "" {
			return fmt.Errorf("sdkerr: auth error: received empty token")
		}

		c.consoleToken = resp.Data.Data.Data.Token
	}

	return nil
}

func buildServerlessClientInfo(appId string) (map[string]any, error) {
	deviceId := strings.ToLower(app.AppName)
	return map[string]any{
		"APPID":              appId,
		"DEVICEID":           deviceId,
		"LOCALE":             "zh-Hans",
		"OS":                 strings.ToLower(runtime.GOOS),
		"PLATFORM":           uniPlatform,
		"appId":              appId,
		"appLanguage":        "zh-Hans",
		"appName":            "uniCloud",
		"appVersion":         "1.0.0",
		"appVersionCode":     "100",
		"deviceId":           deviceId,
		"deviceModel":        "PC",
		"deviceType":         "pc",
		"locale":             "zh-Hans",
		"osName":             runtime.GOOS,
		"osVersion":          runtime.GOARCH,
		"scene":              1001,
		"uniPlatform":        uniPlatform,
		"uniCompilerVersion": "5.23",
		"uniRuntimeVersion":  "5.23",
	}, nil
}

func buildServerlessPayloadInfo(appId, spaceId, functionTarget, functionMethod, functionAction string, functionParams, functionData any, uniIdToken string) (map[string]any, map[string]any, error) {
	clientInfo, err := buildServerlessClientInfo(appId)
	if err != nil {
		return nil, nil, err
	}

	functionArgsParams := make([]any, 0)
	if functionParams != nil {
		functionArgsParams = append(functionArgsParams, functionParams)
	}

	functionArgs := map[string]any{
		"clientInfo": clientInfo,
	}
	if functionMethod != "" {
		functionArgs["method"] = functionMethod
		functionArgs["params"] = make([]any, 0)
	}
	if functionAction != "" {
		type _obj struct{}
		functionArgs["action"] = functionAction
		functionArgs["data"] = &_obj{}
	}
	if functionParams != nil {
		functionArgs["params"] = []any{functionParams}
	}
	if functionData != nil {
		functionArgs["data"] = functionData
	}
	if uniIdToken != "" {
		functionArgs["uniIdToken"] = uniIdToken
	}

	jsonb, err := json.Marshal(map[string]any{
		"functionTarget": functionTarget,
		"functionArgs":   functionArgs,
	})
	if err != nil {
		return nil, nil, err
	}

	payloadInfo := map[string]any{
		"method":    "serverless.function.runtime.invoke",
		"params":    string(jsonb),
		"spaceId":   spaceId,
		"timestamp": time.Now().UnixMilli(),
	}
	return payloadInfo, clientInfo, nil
}

func buildPasswordEnvelopeWithRSAOAEP256AES256GCM(username, password string, challengeId string, challengeVer int, challengeKey string, challengeKeyId string) (map[string]any, error) {
	plain := map[string]any{
		"action":      "login",
		"challengeId": challengeId,
		"issuedAt":    time.Now().UnixMilli(),
		"username":    username,
		"password":    password,
	}
	if regexp.MustCompile(`^1\d{10}$`).MatchString(username) {
		plain["mobile"] = username
		delete(plain, "username")
	} else if regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`).MatchString(username) {
		plain["email"] = username
		delete(plain, "username")
	}

	plainBytes, err := json.Marshal(plain)
	if err != nil {
		return nil, err
	}

	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return nil, err
	}

	ivBytes := make([]byte, 12)
	if _, err := rand.Read(ivBytes); err != nil {
		return nil, err
	}

	aad := map[string]any{
		"version":     challengeVer,
		"algorithm":   challengeAlgRSAOAEP256,
		"keyId":       challengeKeyId,
		"challengeId": challengeId,
	}
	// 务必是固定字段顺序，否则会登录报错。因此这里暂时写死 JSON 字符串拼接。
	// aadBytes, err := json.Marshal(aad)
	// if err != nil {
	// 	return nil, err
	// }
	aadBytes := fmt.Appendf(nil, `{"version":%d,"algorithm":"%s","keyId":"%s","challengeId":"%s"}`, challengeVer, challengeAlgRSAOAEP256, challengeKeyId, challengeId)

	cryptor := xcrypto.NewAESCryptor(keyBytes)
	cipherBytes, err := cryptor.GCMEncrypt(plainBytes, ivBytes, aadBytes)
	if err != nil {
		return nil, err
	}

	const tagSize int = 16
	tagBytes := cipherBytes[len(cipherBytes)-tagSize:]
	cipherBytes = cipherBytes[:len(cipherBytes)-tagSize]

	rsaPubKeyPemBlock, _ := pem.Decode([]byte(challengeKey))
	if rsaPubKeyPemBlock == nil || (rsaPubKeyPemBlock.Type != "PUBLIC KEY") {
		return nil, fmt.Errorf("invalid public key")
	}

	rsaPubKey, err := x509.ParsePKIXPublicKey(rsaPubKeyPemBlock.Bytes)
	if err != nil {
		return nil, err
	}

	encKeyBytes, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, rsaPubKey.(*rsa.PublicKey), keyBytes, nil)
	if err != nil {
		return nil, err
	}

	envelope := map[string]any{
		"encryptedKey": base64.RawStdEncoding.EncodeToString(encKeyBytes),
		"ciphertext":   base64.RawStdEncoding.EncodeToString(cipherBytes),
		"iv":           base64.RawStdEncoding.EncodeToString(ivBytes),
		"tag":          base64.RawStdEncoding.EncodeToString(tagBytes),
	}
	xmaps.CopyTo(aad, envelope)
	return envelope, nil
}

func generateSignature(params map[string]any, secret string) string {
	keys := xmaps.Keys(params)
	sort.Strings(keys)

	canonicalStr := ""
	for i, k := range keys {
		if i > 0 {
			canonicalStr += "&"
		}
		canonicalStr += k + "=" + fmt.Sprintf("%v", params[k])
	}

	mac := hmac.New(md5.New, []byte(secret))
	mac.Write([]byte(canonicalStr))
	sign := mac.Sum(nil)
	signHex := hex.EncodeToString(sign)

	return signHex
}
