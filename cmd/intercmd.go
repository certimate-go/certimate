package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/go-acme/lego/v4/lego"
	legolog "github.com/go-acme/lego/v4/log"
	"github.com/pocketbase/pocketbase/core"
	"github.com/spf13/cobra"

	"github.com/certimate-go/certimate/internal/certacme"
	"github.com/certimate-go/certimate/internal/tools/mproc"
)

func NewInternalCommand(app core.App) *cobra.Command {
	command := &cobra.Command{
		Use:   "intercmd",
		Short: "[INTERNAL] Internal dedicated for Certimate",
	}

	command.AddCommand(internalCertApplyCommand(app))

	return command
}

func internalCertApplyCommand(app core.App) *cobra.Command {
	var flagInput string
	var flagOutput string
	var flagError string
	var flagEncryptionKey string

	command := &cobra.Command{
		Use:          "certapply",
		Example:      "internal certapply --in ./in.file --out ./out.file --enckey aeskey",
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			type InData struct {
				Account *certacme.ACMEAccount              `json:"account,omitempty"`
				Request *certacme.ObtainCertificateRequest `json:"request,omitempty"`
			}

			type OutData struct {
				Response *certacme.ObtainCertificateResponse `json:"response"`
			}

			mreceiver := mproc.NewReceiver(func(ctx context.Context, params *InData) (*OutData, error) {
				if params.Account == nil {
					return nil, errors.New("illegal params")
				}
				if params.Request == nil {
					return nil, errors.New("illegal params")
				}

				// redirect to stdout, remove datetime prefix
				// so that the logger can split logs correctly
				// see: /internal/tools/mproc/sender.go
				legolog.Logger = log.New(os.Stdout, "", 0)

				client, err := certacme.NewACMEClientWithAccount(params.Account, func(c *lego.Config) error {
					c.UserAgent = "certimate"
					c.Certificate.KeyType = params.Request.PrivateKeyType
					c.Certificate.DisableCommonName = params.Request.NoCommonName
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

	command.PersistentFlags().StringVar(&flagInput, "in", "", "")
	command.PersistentFlags().StringVar(&flagOutput, "out", "", "")
	command.PersistentFlags().StringVar(&flagError, "err", "", "")
	command.PersistentFlags().StringVar(&flagEncryptionKey, "enckey", "", "")

	return command
}
