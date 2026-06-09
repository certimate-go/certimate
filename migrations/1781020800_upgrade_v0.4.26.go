package migrations

import (
	"strings"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		tracer := NewTracer("v0.4.26")
		tracer.Printf("go ...")

		// update collection `certificate`
		//   - add ARI metadata fields
		{
			collection, err := app.FindCollectionByNameOrId("4szxr9x43tpj6np")
			if err != nil {
				return err
			}

			fields := []struct {
				index int
				name  string
				json  string
			}{
				{22, "ariWindowStart", `{
					"hidden": false,
					"id": "date2264264971",
					"max": "",
					"min": "",
					"name": "ariWindowStart",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "date"
				}`},
				{23, "ariWindowEnd", `{
					"hidden": false,
					"id": "date2359593433",
					"max": "",
					"min": "",
					"name": "ariWindowEnd",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "date"
				}`},
				{24, "ariNextRefreshAt", `{
					"hidden": false,
					"id": "date553418413",
					"max": "",
					"min": "",
					"name": "ariNextRefreshAt",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "date"
				}`},
				{25, "ariSupported", `{
					"hidden": false,
					"id": "bool1675854919",
					"name": "ariSupported",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "bool"
				}`},
			}

			changed := false
			for _, field := range fields {
				if collection.Fields.GetByName(field.name) != nil {
					continue
				}

				if err := collection.Fields.AddMarshaledJSONAt(field.index, []byte(field.json)); err != nil {
					return err
				}
				changed = true
			}

			if changed {
				if err := app.Save(collection); err != nil {
					return err
				}

				tracer.Printf("collection '%s' updated", collection.Name)
			}

			records, err := app.FindAllRecords(collection)
			if err != nil {
				return err
			}

			for _, record := range records {
				if record.GetBool("ariSupported") {
					continue
				}

				if isKnownARISupportingCA(record.GetString("acmeAcctUrl")) {
					record.Set("ariSupported", true)
					if err := app.Save(record); err != nil {
						return err
					}

					tracer.Printf("record #%s in collection '%s' updated", record.Id, collection.Name)
				}
			}
		}

		tracer.Printf("done")
		return nil
	}, nil)
}

func isKnownARISupportingCA(acmeAcctUrl string) bool {
	acmeAcctUrl = strings.ToLower(acmeAcctUrl)
	return strings.Contains(acmeAcctUrl, "letsencrypt.org/") ||
		strings.Contains(acmeAcctUrl, "api.pki.goog/")
}
