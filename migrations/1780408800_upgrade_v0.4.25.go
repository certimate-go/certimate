package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		tracer := NewTracer("v0.4.25")
		tracer.Printf("go ...")

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
				case "byteplus":
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
