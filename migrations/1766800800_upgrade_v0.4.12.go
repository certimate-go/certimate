package migrations

import (
	"strings"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"

	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
	xcertx509 "github.com/certimate-go/certimate/pkg/utils/cert/x509"
)

func init() {
	m.Register(func(app core.App) error {
		tracer := NewTracer("v0.4.12")
		tracer.Printf("go ...")

		// update collection `certificate`
		//   - update field `subjectAltNames`
		{
			collection, err := app.FindCollectionByNameOrId("4szxr9x43tpj6np")
			if err != nil {
				return err
			}

			records, err := app.FindAllRecords(collection)
			if err != nil {
				return err
			}

			for _, record := range records {
				changed := false

				if certX509, err := xcert.ParseCertificateFromPEM(record.GetString("certificate")); err == nil {
					certSANs := xcertx509.GetSubjectAltNames(certX509)
					if strings.Join(certSANs, ";") != record.GetString("subjectAltNames") {
						record.Set("subjectAltNames", strings.Join(certSANs, ";"))
						changed = true
					}
				}

				if changed {
					if err := app.Save(record); err != nil {
						return err
					}

					tracer.Printf("record #%s in collection '%s' updated", record.Id, collection.Name)
				}
			}
		}

		tracer.Printf("done")
		return nil
	}, func(app core.App) error {
		return nil
	})
}
