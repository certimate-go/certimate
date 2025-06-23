package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		tracer := NewTracer("(v0.3)1750687200")
		tracer.Printf("go ...")

		// update collection `workflow_run`
		{
			collection, err := app.FindCollectionByNameOrId("qjp8lygssgwyqyz")
			if err != nil {
				return err
			}

			// update collection data
			if err := json.Unmarshal([]byte(`{
				"indexes": [
					"CREATE INDEX `+"`"+`idx_7ZpfjTFsD2`+"`"+` ON `+"`"+`workflow_run`+"`"+` (`+"`"+`workflowId`+"`"+`)"
				]
			}`), &collection); err != nil {
				return err
			}

			if err := app.Save(collection); err != nil {
				return err
			}

			tracer.Printf("collection '%s' updated", collection.Name)
		}

		// update collection `workflow_output`
		{
			collection, err := app.FindCollectionByNameOrId("bqnxb95f2cooowp")
			if err != nil {
				return err
			}

			// update collection data
			if err := json.Unmarshal([]byte(`{
				"indexes": [
					"CREATE INDEX `+"`"+`idx_BYoQPsz4my`+"`"+` ON `+"`"+`workflow_output`+"`"+` (`+"`"+`workflowId`+"`"+`)",
					"CREATE INDEX `+"`"+`idx_O9zxLETuxJ`+"`"+` ON `+"`"+`workflow_output`+"`"+` (`+"`"+`runId`+"`"+`)",
					"CREATE INDEX `+"`"+`idx_luac8Ul34G`+"`"+` ON `+"`"+`workflow_output`+"`"+` (`+"`"+`nodeId`+"`"+`)"
				]
			}`), &collection); err != nil {
				return err
			}

			if err := app.Save(collection); err != nil {
				return err
			}

			tracer.Printf("collection '%s' updated", collection.Name)
		}

		// update collection `workflow_logs`
		{
			collection, err := app.FindCollectionByNameOrId("pbc_1682296116")
			if err != nil {
				return err
			}

			// update collection data
			if err := json.Unmarshal([]byte(`{
				"indexes": [
					"CREATE INDEX `+"`"+`idx_IOlpy6XuJ2`+"`"+` ON `+"`"+`workflow_logs`+"`"+` (`+"`"+`workflowId`+"`"+`)",
					"CREATE INDEX `+"`"+`idx_qVlTb2yl7v`+"`"+` ON `+"`"+`workflow_logs`+"`"+` (`+"`"+`runId`+"`"+`)",
					"CREATE INDEX `+"`"+`idx_UL4tdCXNlA`+"`"+` ON `+"`"+`workflow_logs`+"`"+` (`+"`"+`nodeId`+"`"+`)"
				]
			}`), &collection); err != nil {
				return err
			}

			if err := app.Save(collection); err != nil {
				return err
			}

			tracer.Printf("collection '%s' updated", collection.Name)
		}

		tracer.Printf("done")
		return nil
	}, func(app core.App) error {
		return nil
	})
}
