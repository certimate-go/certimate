package autossl

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSigner_Sign(t *testing.T) {
	frozen := time.UnixMilli(1768816669358)
	s := &signer{
		accessKey: "test-access-key",
		secretKey: "test-secret-key",
		now:       func() time.Time { return frozen },
	}

	req, err := http.NewRequest(http.MethodGet, "https://api.auto-ssl.cn/v1/api/cert/getCertList?page=1&size=50", nil)
	require.NoError(t, err)

	err = s.Sign(req)
	require.NoError(t, err)

	assert.Equal(t, "test-access-key", req.Header.Get("X-AUTOSSL-API-KEY"))
	assert.Equal(t, "1768816669358", req.Header.Get("X-AUTOSSL-TIMESTAMP"))
	assert.NotEmpty(t, req.Header.Get("X-AUTOSSL-NONCE"))

	// 期望签名：HMAC-SHA256(SK, "GET\n/v1/api/cert/getCertList\n{timestamp}") 的 Base64 值
	// 注意：签名路径不含查询参数。
	mac := hmac.New(sha256.New, []byte("test-secret-key"))
	mac.Write([]byte("GET\n/v1/api/cert/getCertList\n1768816669358"))
	expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	assert.Equal(t, expected, req.Header.Get("X-AUTOSSL-SIGNATURE"))

	// POST 请求签名应使用 POST 方法
	reqPost, err := http.NewRequest(http.MethodPost, "https://api.auto-ssl.cn/v1/api/cert/downloadCert", nil)
	require.NoError(t, err)

	err = s.Sign(reqPost)
	require.NoError(t, err)

	mac = hmac.New(sha256.New, []byte("test-secret-key"))
	mac.Write([]byte("POST\n/v1/api/cert/downloadCert\n1768816669358"))
	expected = base64.StdEncoding.EncodeToString(mac.Sum(nil))
	assert.Equal(t, expected, reqPost.Header.Get("X-AUTOSSL-SIGNATURE"))
}

func TestParseCertTime(t *testing.T) {
	assert.Equal(t, time.UnixMilli(1768816669358), parseCertTime(float64(1768816669358)))
	assert.Equal(t, time.Unix(1768816669, 0), parseCertTime(float64(1768816669)))

	parsed := parseCertTime("2026-01-18 18:10:21")
	assert.Equal(t, 2026, parsed.Year())
	assert.Equal(t, time.January, parsed.Month())
	assert.Equal(t, 18, parsed.Hour())

	parsed = parseCertTime("2026-01-18T18:10:21Z")
	assert.Equal(t, 2026, parsed.Year())

	parsed = parseCertTime("2026-01-18")
	assert.Equal(t, 2026, parsed.Year())

	assert.True(t, parseCertTime("not-a-time").IsZero())
	assert.True(t, parseCertTime(nil).IsZero())
}

func TestParseCertificateItems(t *testing.T) {
	// 形态一：data 为对象，含 list 数组
	{
		payload := []byte(`{"code":200,"data":{"total":2,"list":[{"certId":"c1","url":"example.com","certEndTime":"2026-06-01 23:59:59","crtBrand":"Let's Encrypt"},{"certId":"c2","url":"*.example.org","certEndTime":1768816669358}]}}`)
		items, err := parseCertificateItems(payload)
		require.NoError(t, err)
		require.Len(t, items, 2)
		assert.Equal(t, "c1", items[0].CertId)
		assert.Equal(t, "example.com", items[0].Url)
		assert.Equal(t, 2026, items[0].CertEndTime.Year())
		assert.Equal(t, "c2", items[1].CertId)
		assert.Equal(t, "*.example.org", items[1].Url)
	}

	// 形态二：data 直接为数组
	{
		payload := []byte(`{"code":200,"data":[{"certId":"c1","url":"example.com"}]}`)
		items, err := parseCertificateItems(payload)
		require.NoError(t, err)
		require.Len(t, items, 1)
		assert.Equal(t, "c1", items[0].CertId)
	}

	// 形态三：顶层直接为数组
	{
		payload := []byte(`[{"certId":"c1","url":"example.com"}]`)
		items, err := parseCertificateItems(payload)
		require.NoError(t, err)
		require.Len(t, items, 1)
	}

	// 无法识别的结构
	{
		payload := []byte(`{"code":200,"data":{"foo":"bar"}}`)
		_, err := parseCertificateItems(payload)
		assert.Error(t, err)
	}
}

func TestParseCertificateKeypair(t *testing.T) {
	// 形态一：data 为对象
	{
		payload := []byte(`{"code":200,"data":{"pem":"-----BEGIN CERTIFICATE-----","key":"-----BEGIN PRIVATE KEY-----"}}`)
		keypair, err := parseCertificateKeypair(payload)
		require.NoError(t, err)
		assert.Equal(t, "-----BEGIN CERTIFICATE-----", keypair.CertPEM)
		assert.Equal(t, "-----BEGIN PRIVATE KEY-----", keypair.PrivkeyPEM)
	}

	// 形态二：顶层直接为对象
	{
		payload := []byte(`{"pem":"-----BEGIN CERTIFICATE-----","key":"-----BEGIN PRIVATE KEY-----"}`)
		keypair, err := parseCertificateKeypair(payload)
		require.NoError(t, err)
		assert.Equal(t, "-----BEGIN CERTIFICATE-----", keypair.CertPEM)
	}

	// 无法识别的结构
	{
		payload := []byte(`{"code":200,"data":{"foo":"bar"}}`)
		_, err := parseCertificateKeypair(payload)
		assert.Error(t, err)
	}
}
