package cert

import (
	"archive/zip"
	"bytes"
	"fmt"

	xcertpfx "github.com/certimate-go/certimate/pkg/utils/cert/pfx"
)

// 表示证书压缩包（zip）生成选项。
type CertificateArchiveOptions struct {
	// 证书文件格式（PEM / PFX / JKS）。
	FileFormat string
	// PFX 导出密码。
	// 证书文件格式为 PFX 时使用，缺省默认值 "certimate"。
	PfxPassword string
	// PFX 编码器。
	PfxEncoder string
	// JKS 别名。
	JksAlias string
	// JKS 密钥密码。
	JksKeypass string
	// JKS 存储密码。
	JksStorepass string
}

// 将证书打包为 zip 压缩包，文件命名规则与页面「下载证书」保持一致。
//
// 入参：
//   - certPEM：证书 PEM 内容。
//   - privkeyPEM：私钥 PEM 内容。
//   - canonicalName：证书主域名（不含通配符前缀，如 *.example.com 传 example.com）。
//   - opts：压缩包生成选项。
//
// 出参：
//   - data：zip 格式的压缩包数据。
//   - err：错误。
func BuildCertificateArchive(certPEM string, privkeyPEM string, canonicalName string, opts CertificateArchiveOptions) (_data []byte, _err error) {
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	var zipBytes []byte
	switch opts.FileFormat {
	case "", "PEM":
		{
			serverCertPEM, issuerCertPEM, err := ExtractCertificatesFromPEM(certPEM)
			if err != nil {
				return nil, fmt.Errorf("failed to extract certs: %w", err)
			}

			keyWriter, err := zipWriter.Create(fmt.Sprintf("%s.key", canonicalName))
			if err != nil {
				return nil, err
			} else {
				_, err = keyWriter.Write([]byte(privkeyPEM))
				if err != nil {
					return nil, err
				}
			}

			certWriter, err := zipWriter.Create(fmt.Sprintf("%s.crt", canonicalName))
			if err != nil {
				return nil, err
			} else {
				_, err = certWriter.Write([]byte(certPEM))
				if err != nil {
					return nil, err
				}
			}

			serverCertWriter, err := zipWriter.Create(fmt.Sprintf("%s (server).pem", canonicalName))
			if err != nil {
				return nil, err
			} else {
				_, err = serverCertWriter.Write([]byte(serverCertPEM))
				if err != nil {
					return nil, err
				}
			}

			intermediaCertWriter, err := zipWriter.Create(fmt.Sprintf("%s (intermedia).pem", canonicalName))
			if err != nil {
				return nil, err
			} else {
				_, err = intermediaCertWriter.Write([]byte(issuerCertPEM))
				if err != nil {
					return nil, err
				}
			}

			err = zipWriter.Close()
			if err != nil {
				return nil, err
			}

			zipBytes = buf.Bytes()
		}

	case "PFX":
		{
			pfxPassword := "certimate"
			if opts.PfxPassword != "" {
				pfxPassword = opts.PfxPassword
			}

			pfxEncoder, err := xcertpfx.ResolvePfxEncoder(opts.PfxEncoder)
			if err != nil {
				return nil, err
			}

			certPFX, err := TransformCertificateFromPEMToPFX(certPEM, privkeyPEM, pfxPassword, pfxEncoder)
			if err != nil {
				return nil, err
			}

			certWriter, err := zipWriter.Create(fmt.Sprintf("%s.pfx", canonicalName))
			if err != nil {
				return nil, err
			} else {
				_, err = certWriter.Write(certPFX)
				if err != nil {
					return nil, err
				}
			}

			readmeWriter, err := zipWriter.Create("README.txt")
			if err != nil {
				return nil, err
			} else {
				readme := fmt.Sprintf("[PFX Password]\n%s\n", pfxPassword)
				_, err = readmeWriter.Write([]byte(readme))
				if err != nil {
					return nil, err
				}
			}

			err = zipWriter.Close()
			if err != nil {
				return nil, err
			}

			zipBytes = buf.Bytes()
		}

	case "JKS":
		{
			jksAlias := "certimate"
			if opts.JksAlias != "" {
				jksAlias = opts.JksAlias
			}

			jksKeypass := "certimate"
			if opts.JksKeypass != "" {
				jksKeypass = opts.JksKeypass
			}

			jksStorepass := "certimate"
			if opts.JksStorepass != "" {
				jksStorepass = opts.JksStorepass
			}

			certJKS, err := TransformCertificateFromPEMToJKS(certPEM, privkeyPEM, jksAlias, jksKeypass, jksStorepass)
			if err != nil {
				return nil, err
			}

			certWriter, err := zipWriter.Create(fmt.Sprintf("%s.jks", canonicalName))
			if err != nil {
				return nil, err
			} else {
				_, err = certWriter.Write(certJKS)
				if err != nil {
					return nil, err
				}
			}

			readmeWriter, err := zipWriter.Create("README.txt")
			if err != nil {
				return nil, err
			} else {
				readme := fmt.Sprintf("[JKS Alias]\n%s\n\n[JKS Key Password]\n%s\n\n[JKS Store Password]\n%s\n", jksAlias, jksKeypass, jksStorepass)
				_, err = readmeWriter.Write([]byte(readme))
				if err != nil {
					return nil, err
				}
			}

			err = zipWriter.Close()
			if err != nil {
				return nil, err
			}

			zipBytes = buf.Bytes()
		}

	default:
		return nil, fmt.Errorf("unsupported certificate format: '%s'", opts.FileFormat)
	}

	return zipBytes, nil
}
