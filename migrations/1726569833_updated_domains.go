package migrations

import (
	"encoding/json"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/models/schema"
)

func init() {
	m.Register(func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("z3p974ainxjqlvs")
		if err != nil {
			return err
		}

		// update
		edit_targetType := &schema.SchemaField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "srybpixz",
			"name": "targetType",
			"type": "select",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSelect": 1,
				"values": [
					"aliyun-oss",
					"aliyun-cdn",
					"aliyun-dcdn",
					"ssh",
					"webhook",
					"tencent-cdn",
					"qiniu-cdn"
				]
			}
		}`), edit_targetType); err != nil {
			return err
		}
		collection.Schema.AddField(edit_targetType)

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("z3p974ainxjqlvs")
		if err != nil {
			return err
		}

		// update
		edit_targetType := &schema.SchemaField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "srybpixz",
			"name": "targetType",
			"type": "select",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSelect": 1,
				"values": [
					"aliyun-oss",
					"aliyun-cdn",
					"ssh",
					"webhook",
					"tencent-cdn",
					"qiniu-cdn"
				]
			}
		}`), edit_targetType); err != nil {
			return err
		}
		collection.Schema.AddField(edit_targetType)

		return dao.SaveCollection(collection)
	})
}
