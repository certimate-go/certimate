package engine

import (
	"crypto/x509"
	"fmt"
	"log/slog"
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/certimate-go/certimate/internal/app"
	"github.com/certimate-go/certimate/internal/repository"
	xcertx509 "github.com/certimate-go/certimate/pkg/utils/cert/x509"
	xhttp "github.com/certimate-go/certimate/pkg/utils/http"
	xtls "github.com/certimate-go/certimate/pkg/utils/tls"
)

/**
 * Variables:
 *   - "certificate.commanName": string
 *   - "certificate.subjectAltNames": string
 *   - "certificate.notBefore": datetime
 *   - "certificate.notAfter": datetime
 *   - "certificate.hoursLeft": number
 *   - "certificate.daysLeft": number
 *   - "certificate.validity": boolean
 */
type bizMonitorNodeExecutor struct {
	nodeExecutor

	certificateRepo certificateRepository
}

func (ne *bizMonitorNodeExecutor) Execute(execCtx *NodeExecutionContext) (*NodeExecutionResult, error) {
	execRes := newNodeExecutionResult(execCtx.Node)

	nodeCfg := execCtx.Node.Data.Config.AsBizMonitor()
	ne.logger.Info("ready to monitor certificate ...", slog.Any("config", nodeCfg))

	targetAddr := net.JoinHostPort(nodeCfg.Host, strconv.Itoa(int(nodeCfg.Port)))
	if nodeCfg.Port == 0 {
		targetAddr = net.JoinHostPort(nodeCfg.Host, "443")
	}

	targetDomain := nodeCfg.Domain
	if targetDomain == "" {
		targetDomain = nodeCfg.Host
	}

	ne.logger.Info(fmt.Sprintf("retrieving certificate at %s (domain: %s)", targetAddr, targetDomain))

	const MAX_ATTEMPTS = 3
	const RETRY_INTERVAL = 2 * time.Second
	var err error
	var certs []*x509.Certificate
	for attempt := 0; attempt < MAX_ATTEMPTS; attempt++ {
		if attempt > 0 {
			ne.logger.Info(fmt.Sprintf("retry %d time(s) ...", attempt))

			select {
			case <-execCtx.ctx.Done():
				return execRes, execCtx.ctx.Err()
			case <-time.After(RETRY_INTERVAL):
			}
		}

		certs, err = ne.tryRetrievePeerCertificates(execCtx, targetAddr, targetDomain, nodeCfg.RequestPath)
		if err == nil {
			break
		}
	}

	if err != nil {
		ne.logger.Warn("could not retrieve certificate")
		return execRes, err
	} else {
		if len(certs) == 0 {
			ne.logger.Warn("no ssl certificates retrieved in http response")

			ne.setVariablesOfResult(execCtx, execRes, nil)
		} else {
			cert := certs[0] // 只取证书链中的第一个证书，即服务器证书
			ne.logger.Info(fmt.Sprintf("ssl certificate retrieved (serial='%s', subject='%s', issuer='%s', not_before='%s', not_after='%s', sans='%s')",
				cert.SerialNumber, cert.Subject.String(), cert.Issuer.String(),
				cert.NotBefore.Format(time.RFC3339), cert.NotAfter.Format(time.RFC3339),
				strings.Join(xcertx509.GetSubjectAltNames(cert), ";")),
			)
			ne.setVariablesOfResult(execCtx, execRes, cert)

			now := time.Now()
			isCertPeriodValid := now.Before(cert.NotAfter) && now.After(cert.NotBefore)
			isCertHostMatched := cert.VerifyHostname(targetDomain) == nil
			daysLeft := int32(math.Floor(time.Until(cert.NotAfter).Hours() / 24))
			validated := isCertPeriodValid && isCertHostMatched

			if validated {
				ne.logger.Info(fmt.Sprintf("the certificate is valid, and will expire in %d day(s)", daysLeft))
			} else {
				if !isCertHostMatched {
					ne.logger.Warn("the certificate is invalid, because it is not matched the host")
				} else if !isCertPeriodValid {
					ne.logger.Warn("the certificate is invalid, because it is either expired or not yet valid")
				} else {
					ne.logger.Warn("the certificate is invalid")
				}

				// 除了验证证书有效期，还要确保证书与域名匹配
				execRes.AddVariable(stateVarKeyCertificateValidity, false, stateValTypeBoolean)
				execRes.AddVariableWithScope(execCtx.Node.Id, stateVarKeyCertificateValidity, false, stateValTypeBoolean)
			}
		}
	}

	ne.logger.Info("monitoring completed")
	return execRes, nil
}

func (ne *bizMonitorNodeExecutor) tryRetrievePeerCertificates(execCtx *NodeExecutionContext, addr, domain, requestPath string) ([]*x509.Certificate, error) {
	transport := xhttp.NewDefaultTransport()
	transport.DisableKeepAlives = true
	transport.TLSClientConfig = xtls.NewInsecureConfig()

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout:   30 * time.Second,
		Transport: transport,
	}

	url := fmt.Sprintf("https://%s/%s", addr, strings.TrimLeft(requestPath, "/"))
	req, err := http.NewRequestWithContext(execCtx.ctx, http.MethodHead, url, nil)
	if err != nil {
		err = fmt.Errorf("failed to create http request: %w", err)
		ne.logger.Warn(err.Error())
		return nil, err
	}

	req.Header.Set("Host", domain)
	req.Header.Set("User-Agent", app.AppUserAgent)
	resp, err := client.Do(req)
	if err != nil {
		err = fmt.Errorf("failed to send http request: %w", err)
		ne.logger.Warn(err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	if resp.TLS == nil || len(resp.TLS.PeerCertificates) == 0 {
		return make([]*x509.Certificate, 0), nil
	}
	return resp.TLS.PeerCertificates, nil
}

func (ne *bizMonitorNodeExecutor) setVariablesOfResult(execCtx *NodeExecutionContext, execRes *NodeExecutionResult, certX509 *x509.Certificate) {
	var vCommonName string
	var vSubjectAltNames string
	var vNotBefore time.Time
	var vNotAfter time.Time
	var vHoursLeft int32
	var vDaysLeft int32
	var vValidity bool

	if certX509 != nil {
		vCommonName = certX509.Subject.CommonName
		vSubjectAltNames = strings.Join(xcertx509.GetSubjectAltNames(certX509), ";")
		vNotBefore = certX509.NotBefore
		vNotAfter = certX509.NotAfter
		vHoursLeft = int32(math.Floor(time.Until(certX509.NotAfter).Hours()))
		vDaysLeft = int32(math.Floor(time.Until(certX509.NotAfter).Hours() / 24))
		vValidity = certX509.NotAfter.After(time.Now())
	}

	execRes.AddVariable(stateVarKeyCertificateDomain, vCommonName, stateValTypeString)
	execRes.AddVariable(stateVarKeyCertificateDomains, vSubjectAltNames, stateValTypeString)
	execRes.AddVariable(stateVarKeyCertificateCommonName, vCommonName, stateValTypeString)
	execRes.AddVariable(stateVarKeyCertificateSubjectAltNames, vSubjectAltNames, stateValTypeString)
	execRes.AddVariable(stateVarKeyCertificateNotBefore, vNotBefore, stateValTypeDateTime)
	execRes.AddVariable(stateVarKeyCertificateNotAfter, vNotAfter, stateValTypeDateTime)
	execRes.AddVariable(stateVarKeyCertificateHoursLeft, vHoursLeft, stateValTypeNumber)
	execRes.AddVariable(stateVarKeyCertificateDaysLeft, vDaysLeft, stateValTypeNumber)
	execRes.AddVariable(stateVarKeyCertificateValidity, vValidity, stateValTypeBoolean)
	execRes.AddVariableWithScope(execCtx.Node.Id, stateVarKeyCertificateDomain, vCommonName, stateValTypeString)
	execRes.AddVariableWithScope(execCtx.Node.Id, stateVarKeyCertificateDomains, vSubjectAltNames, stateValTypeString)
	execRes.AddVariableWithScope(execCtx.Node.Id, stateVarKeyCertificateCommonName, vCommonName, stateValTypeString)
	execRes.AddVariableWithScope(execCtx.Node.Id, stateVarKeyCertificateSubjectAltNames, vSubjectAltNames, stateValTypeString)
	execRes.AddVariableWithScope(execCtx.Node.Id, stateVarKeyCertificateNotBefore, vNotBefore, stateValTypeDateTime)
	execRes.AddVariableWithScope(execCtx.Node.Id, stateVarKeyCertificateNotAfter, vNotAfter, stateValTypeDateTime)
	execRes.AddVariableWithScope(execCtx.Node.Id, stateVarKeyCertificateHoursLeft, vHoursLeft, stateValTypeNumber)
	execRes.AddVariableWithScope(execCtx.Node.Id, stateVarKeyCertificateDaysLeft, vDaysLeft, stateValTypeNumber)
	execRes.AddVariableWithScope(execCtx.Node.Id, stateVarKeyCertificateValidity, vValidity, stateValTypeBoolean)
}

func newBizMonitorNodeExecutor() NodeExecutor {
	return &bizMonitorNodeExecutor{
		nodeExecutor:    nodeExecutor{logger: slog.Default()},
		certificateRepo: repository.NewCertificateRepository(),
	}
}
