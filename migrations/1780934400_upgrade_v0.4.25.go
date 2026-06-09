package migrations

import (
	"errors"
	"strings"

	"github.com/go-acme/lego/v5/acme"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"

	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
	xcertx509 "github.com/certimate-go/certimate/pkg/utils/cert/x509"
)

func init() {
	m.Register(func(app core.App) error {
		tracer := NewTracer("v0.4.25")
		tracer.Printf("go ...")

		// update collection `acme_accounts`
		//   - rename field `acmeAccount` to `resourceObj`
		{
			collection, err := app.FindCollectionByNameOrId("012d7abbod1hwvr")
			if err != nil {
				return err
			}

			if err := collection.Fields.AddMarshaledJSONAt(4, []byte(`{
				"hidden": false,
				"id": "1aoia909",
				"maxSize": 2000000,
				"name": "resourceObj",
				"presentable": false,
				"required": false,
				"system": false,
				"type": "json"
			}`)); err != nil {
				return err
			}

			if err := app.Save(collection); err != nil {
				return err
			}

			tracer.Printf("collection '%s' updated", collection.Name)

			records, err := app.FindAllRecords(collection)
			if err != nil {
				return err
			}

			for _, record := range records {
				changed := false

				var resourceObj *acme.Account
				if err := record.UnmarshalJSONField("resourceObj", &resourceObj); err != nil {
					return err
				} else {
					if len(resourceObj.Contact) == 0 {
						resourceObj.Contact = []string{"mailto:" + record.GetString("email")}

						record.Set("resourceObj", resourceObj)
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

		// update collection `access`
		//   - modify field `config` schema
		{
			collection, err := app.FindCollectionByNameOrId("4yzbv8urny5ja1e")
			if err != nil {
				return err
			}

			records, err := app.FindAllRecords(collection)
			if err != nil {
				return err
			}

			for _, record := range records {
				changed := false

				provider := record.GetString("provider")
				config := make(map[string]any)
				if err := record.UnmarshalJSONField("config", &config); err != nil {
					return err
				}

				switch provider {
				case "cloudflare":
					{
						if _, ok := config["dnsApiToken"]; ok {
							config["apiToken"] = config["dnsApiToken"]
							delete(config, "dnsApiToken")
							record.Set("config", config)
							changed = true
						}
						if _, ok := config["zoneApiToken"]; ok {
							config["apiTokenForZone"] = config["zoneApiToken"]
							delete(config, "zoneApiToken")
							record.Set("config", config)
							changed = true
						}
					}
				case "byteplus":
					{
						if _, ok := config["accessKey"]; ok {
							config["accessKeyId"] = config["accessKey"]
							delete(config, "accessKey")
							record.Set("config", config)
							changed = true
						}
						if _, ok := config["secretKey"]; ok {
							config["secretAccessKey"] = config["secretKey"]
							delete(config, "secretKey")
							record.Set("config", config)
							changed = true
						}
					}
				case "volcengine":
					{
						if _, ok := config["accessKeySecret"]; ok {
							config["secretAccessKey"] = config["accessKeySecret"]
							delete(config, "accessKeySecret")
							record.Set("config", config)
							changed = true
						}
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

		// update collection `certificate`
		//   - add field `subjectName`
		//   - add field `issuerName`
		//   - add field `validationPolicy`
		//   - add field `ca`
		{
			collection, err := app.FindCollectionByNameOrId("4szxr9x43tpj6np")
			if err != nil {
				return err
			}

			if err := collection.Fields.AddMarshaledJSONAt(2, []byte(`{
				"autogeneratePattern": "",
				"help": "",
				"hidden": false,
				"id": "plmambpz",
				"max": 100000,
				"min": 0,
				"name": "certificate",
				"pattern": "",
				"presentable": false,
				"primaryKey": false,
				"required": true,
				"system": false,
				"type": "text"
			}`)); err != nil {
				return err
			}

			if err := collection.Fields.AddMarshaledJSONAt(3, []byte(`{
				"autogeneratePattern": "",
				"help": "",
				"hidden": false,
				"id": "49qvwxcg",
				"max": 100000,
				"min": 0,
				"name": "privateKey",
				"pattern": "",
				"presentable": false,
				"primaryKey": false,
				"required": true,
				"system": false,
				"type": "text"
			}`)); err != nil {
				return err
			}

			if err := collection.Fields.AddMarshaledJSONAt(4, []byte(`{
				"autogeneratePattern": "",
				"help": "",
				"hidden": false,
				"id": "text2069360702",
				"max": 0,
				"min": 0,
				"name": "serialNumber",
				"pattern": "",
				"presentable": false,
				"primaryKey": false,
				"required": false,
				"system": false,
				"type": "text"
			}`)); err != nil {
				return err
			}

			if err := collection.Fields.AddMarshaledJSONAt(5, []byte(`{
				"autogeneratePattern": "",
				"help": "",
				"hidden": false,
				"id": "text2876278798",
				"max": 0,
				"min": 0,
				"name": "subjectName",
				"pattern": "",
				"presentable": false,
				"primaryKey": false,
				"required": false,
				"system": false,
				"type": "text"
			}`)); err != nil {
				return err
			}

			if err := collection.Fields.AddMarshaledJSONAt(7, []byte(`{
				"autogeneratePattern": "",
				"help": "",
				"hidden": false,
				"id": "text2678583873",
				"max": 0,
				"min": 0,
				"name": "issuerName",
				"pattern": "",
				"presentable": false,
				"primaryKey": false,
				"required": false,
				"system": false,
				"type": "text"
			}`)); err != nil {
				return err
			}

			if err := collection.Fields.AddMarshaledJSONAt(11, []byte(`{
				"autogeneratePattern": "",
				"help": "",
				"hidden": false,
				"id": "text2516249007",
				"max": 0,
				"min": 0,
				"name": "validationPolicy",
				"pattern": "",
				"presentable": false,
				"primaryKey": false,
				"required": false,
				"system": false,
				"type": "text"
			}`)); err != nil {
				return err
			}

			if err := collection.Fields.AddMarshaledJSONAt(15, []byte(`{
				"autogeneratePattern": "",
				"help": "",
				"hidden": false,
				"id": "text4045880084",
				"max": 0,
				"min": 0,
				"name": "ca",
				"pattern": "",
				"presentable": false,
				"primaryKey": false,
				"required": false,
				"system": false,
				"type": "text"
			}`)); err != nil {
				return err
			}

			if err := app.Save(collection); err != nil {
				return err
			}

			tracer.Printf("collection '%s' updated", collection.Name)

			records, err := app.FindAllRecords(collection)
			if err != nil {
				return err
			}

			for _, record := range records {
				changed := false

				if record.GetString("certificate") != "" {
					certX509, err := xcert.ParseCertificateFromPEM(record.GetString("certificate"))
					if err != nil {
						continue
					}

					record.Set("subjectName", certX509.Subject.CommonName)
					record.Set("issuerName", certX509.Issuer.CommonName)

					switch xcertx509.GetValidationType(certX509) {
					case xcertx509.ExtendedValidation:
						record.Set("validationPolicy", "EV")
					case xcertx509.DomainValidated:
						record.Set("validationPolicy", "DV")
					case xcertx509.OrganizationalValidated:
						record.Set("validationPolicy", "OV")
					case xcertx509.IndividualValidated:
						record.Set("validationPolicy", "IV")
					default:
						record.Set("validationPolicy", "")
					}

					changed = true
				}

				if record.GetString("acmeAcctUrl") != "" {
					acmeAcctUrl := record.GetString("acmeAcctUrl")
					acmeAcctRecord, err := app.FindFirstRecordByFilter("acme_accounts", "acmeAcctUrl={:acmeAcctUrl}", dbx.Params{"acmeAcctUrl": acmeAcctUrl})
					if err != nil {
						mapping := map[string]string{
							"https://acme-v02.api.letsencrypt.org/":         "letsencrypt",
							"https://acme-staging-v02.api.letsencrypt.org/": "letsencryptstaging",
							"https://acme-api.actalis.com/":                 "actalisssl",
							"https://acme.digicert.com/":                    "digicert",
							"https://emea.acme.atlas.globalsign.com/":       "globalsignatlas",
							"https://dv.acme-v02.api.pki.goog/":             "googletrustservices",
							"https://acme.litessl.com/acme/v2/":             "litessl",
							"https://acme.ssl.com/":                         "sslcom",
							"https://acme.sectigo.com/":                     "sectigo",
							"https://acme.zerossl.com/":                     "zerossl",
						}
						for url, ca := range mapping {
							if strings.HasPrefix(acmeAcctUrl, url) {
								record.Set("ca", ca)
								changed = true
								break
							}
						}
					} else {
						record.Set("ca", acmeAcctRecord.GetString("ca"))
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
		return errors.ErrUnsupported
	})
}
