package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/go-acme/lego/v5/lego"
	"github.com/go-acme/lego/v5/log"
	"github.com/pocketbase/pocketbase/core"
	"github.com/spf13/cobra"

	"github.com/certimate-go/certimate/internal/certacme"
	"github.com/certimate-go/certimate/internal/tools/mproc"
	"github.com/certimate-go/certimate/pkg/logging"
)

func NewInternalCommand(app core.App) *cobra.Command {
	command := &cobra.Command{
		Use:   "intercmd",
		Short: "[RESERVED] PLEASE DO NOT USE!",
	}

	command.AddCommand(internalCertApplyCommand(app))

	return command
}

func internalCertApplyCommand(_ core.App) *cobra.Command {
	var flagInput string
	var flagOutput string
	var flagError string
	var flagEncryptionKey string

	hookStdLog := func(namespace string) *slog.Logger {
		jHdr := slog.NewJSONHandler(os.Stdout, nil)
		nHdr := logging.NewNamedHandler(jHdr, namespace)
		hHdr := logging.NewHookHandler(nil, &logging.HookHandlerOptions{
			WriteFunc: func(ctx context.Context, record logging.Record) error {
				copy := slog.NewRecord(record.Time, record.Level, record.Message, record.PC)

				record.Attrs(func(a slog.Attr) bool {
					if a.Value.Kind() == slog.KindDuration {
						a = log.DurationAttr(a.Key, a.Value.Duration())
					} else if a.Value.Kind() == slog.KindAny && a.Value.Any() != nil {
						if d, ok := a.Value.Any().(time.Duration); ok {
							a = log.DurationAttr(a.Key, d)
						}
					}

					copy.AddAttrs(a)
					return true
				})

				nHdr.Handle(ctx, copy)
				return nil
			},
		})
		return slog.New(hHdr)
	}

	command := &cobra.Command{
		Use:          "certapply",
		Example:      "internal certapply --mprocIn ./in.file --mprocOut ./out.file --mprocSecret aeskey",
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			type InData struct {
				Request *certacme.ObtainCertificateRequest `json:"request,omitempty"`

				LegoAccount         *certacme.ACMEAccount   `json:"legoAccount,omitempty"`
				LegoCertifierConfig *lego.CertificateConfig `json:"legoCertifierConfig,omitempty"`
			}

			type OutData struct {
				Response *certacme.ObtainCertificateResponse `json:"response"`
			}

			mreceiver := mproc.NewReceiver(func(ctx context.Context, params *InData) (*OutData, error) {
				if params.LegoAccount == nil {
					return nil, fmt.Errorf("illegal params")
				}
				if params.Request == nil {
					return nil, fmt.Errorf("illegal params")
				}

				// set lego logger to a wrapped JSON handler,
				// so that the logger can parse logs correctly.
				// see: /internal/tools/mproc/sender.go
				log.SetDefault(hookStdLog("go-acme/lego"))

				client, err := certacme.NewACMEClientWithAccount(params.LegoAccount, func(legoCfg *lego.Config) error {
					if params.LegoCertifierConfig != nil {
						legoCfg.Certificate = *params.LegoCertifierConfig
					}

					return nil
				})
				if err != nil {
					return nil, fmt.Errorf("failed to initialize acme client: %w", err)
				}

				resp, err := client.ObtainCertificate(ctx, params.Request)
				if err != nil {
					return nil, fmt.Errorf("failed to obtain certificate: %w", err)
				}

				return &OutData{
					Response: resp,
				}, nil
			})
			if err := mreceiver.ReceiveWithContext(cmd.Context(), flagInput, flagOutput, flagEncryptionKey); err != nil {
				os.WriteFile(flagError, []byte(err.Error()), 0o644)
			}
		},
	}

	command.PersistentFlags().StringVar(&flagInput, "mprocIn", "", "")
	command.PersistentFlags().StringVar(&flagOutput, "mprocOut", "", "")
	command.PersistentFlags().StringVar(&flagError, "mprocErr", "", "")
	command.PersistentFlags().StringVar(&flagEncryptionKey, "mprocSecret", "", "")

	return command
}
