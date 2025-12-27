package migrations

import (
	"os"
	"strings"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		if mr, _ := app.FindFirstRecordByFilter("_migrations", "file='1757476801_m0.4.0_initialize.go'"); mr != nil {
			return nil
		}

		// snapshot
		{
			jsonData := `[
				{
					"fields": [
						{
							"autogeneratePattern": "[a-z0-9]{15}",
							"hidden": false,
							"id": "text3208210256",
							"max": 15,
							"min": 15,
							"name": "id",
							"pattern": "^[a-z0-9]+$",
							"presentable": false,
							"primaryKey": true,
							"required": true,
							"system": true,
							"type": "text"
						},
						{
							"autogeneratePattern": "",
							"hidden": false,
							"id": "geeur58v",
							"max": 0,
							"min": 0,
							"name": "name",
							"pattern": "",
							"presentable": false,
							"primaryKey": false,
							"required": false,
							"system": false,
							"type": "text"
						},
						{
							"autogeneratePattern": "",
							"hidden": false,
							"id": "text2024822322",
							"max": 0,
							"min": 0,
							"name": "provider",
							"pattern": "",
							"presentable": false,
							"primaryKey": false,
							"required": false,
							"system": false,
							"type": "text"
						},
						{
							"hidden": false,
							"id": "iql7jpwx",
							"maxSize": 2000000,
							"name": "config",
							"presentable": false,
							"required": false,
							"system": false,
							"type": "json"
						},
						{
							"autogeneratePattern": "",
							"hidden": false,
							"id": "text2859962647",
							"max": 0,
							"min": 0,
							"name": "reserve",
							"pattern": "",
							"presentable": false,
							"primaryKey": false,
							"required": false,
							"system": false,
							"type": "text"
						},
						{
							"hidden": false,
							"id": "lr33hiwg",
							"max": "",
							"min": "",
							"name": "deleted",
							"presentable": false,
							"required": false,
							"system": false,
							"type": "date"
						},
						{
							"hidden": false,
							"id": "autodate2990389176",
							"name": "created",
							"onCreate": true,
							"onUpdate": false,
							"presentable": false,
							"system": false,
							"type": "autodate"
						},
						{
							"hidden": false,
							"id": "autodate3332085495",
							"name": "updated",
							"onCreate": true,
							"onUpdate": true,
							"presentable": false,
							"system": false,
							"type": "autodate"
						}
					],
					"id": "4yzbv8urny5ja1e",
					"indexes": [
						"CREATE INDEX ` + "`" + `idx_wkoST0j` + "`" + ` ON ` + "`" + `access` + "`" + ` (` + "`" + `name` + "`" + `)",
						"CREATE INDEX ` + "`" + `idx_frh0JT1Aqx` + "`" + ` ON ` + "`" + `access` + "`" + ` (` + "`" + `provider` + "`" + `)"
					],
					"name": "access",
					"system": false,
					"type": "base"
				},
				{
					"fields": [
						{
							"autogeneratePattern": "[a-z0-9]{15}",
							"hidden": false,
							"id": "text3208210256",
							"max": 15,
							"min": 15,
							"name": "id",
							"pattern": "^[a-z0-9]+$",
							"presentable": false,
							"primaryKey": true,
							"required": true,
							"system": true,
							"type": "text"
						},
						{
							"autogeneratePattern": "",
							"hidden": false,
							"id": "1tcmdsdf",
							"max": 0,
							"min": 0,
							"name": "name",
							"pattern": "",
							"presentable": false,
							"primaryKey": false,
							"required": false,
							"system": false,
							"type": "text"
						},
						{
							"hidden": false,
							"id": "f9wyhypi",
							"maxSize": 2000000,
							"name": "content",
							"presentable": false,
							"required": false,
							"system": false,
							"type": "json"
						},
						{
							"hidden": false,
							"id": "autodate2990389176",
							"name": "created",
							"onCreate": true,
							"onUpdate": false,
							"presentable": false,
							"system": false,
							"type": "autodate"
						},
						{
							"hidden": false,
							"id": "autodate3332085495",
							"name": "updated",
							"onCreate": true,
							"onUpdate": true,
							"presentable": false,
							"system": false,
							"type": "autodate"
						}
					],
					"id": "dy6ccjb60spfy6p",
					"indexes": [
						"CREATE UNIQUE INDEX ` + "`" + `idx_RO7X9Vw` + "`" + ` ON ` + "`" + `settings` + "`" + ` (` + "`" + `name` + "`" + `)"
					],
					"name": "settings",
					"system": false,
					"type": "base"
				},
				{
					"fields": [
						{
							"autogeneratePattern": "[a-z0-9]{15}",
							"hidden": false,
							"id": "text3208210256",
							"max": 15,
							"min": 15,
							"name": "id",
							"pattern": "^[a-z0-9]+$",
							"presentable": false,
							"primaryKey": true,
							"required": true,
							"system": true,
							"type": "text"
						},
						{
							"autogeneratePattern": "",
							"hidden": false,
							"id": "fmjfn0yw",
							"max": 0,
							"min": 0,
							"name": "ca",
							"pattern": "",
							"presentable": false,
							"primaryKey": false,
							"required": false,
							"system": false,
							"type": "text"
						},
						{
							"exceptDomains": null,
							"hidden": false,
							"id": "qqwijqzt",
							"name": "email",
							"onlyDomains": null,
							"presentable": false,
							"required": false,
							"system": false,
							"type": "email"
						},
						{
							"autogeneratePattern": "",
							"hidden": false,
							"id": "genxqtii",
							"max": 0,
							"min": 0,
							"name": "privateKey",
							"pattern": "",
							"presentable": false,
							"primaryKey": false,
							"required": false,
							"system": false,
							"type": "text"
						},
						{
							"hidden": false,
							"id": "1aoia909",
							"maxSize": 2000000,
							"name": "acmeAccount",
							"presentable": false,
							"required": false,
							"system": false,
							"type": "json"
						},
						{
							"exceptDomains": null,
							"hidden": false,
							"id": "url2424532088",
							"name": "acmeAcctUrl",
							"onlyDomains": null,
							"presentable": false,
							"required": false,
							"system": false,
							"type": "url"
						},
						{
							"exceptDomains": null,
							"hidden": false,
							"id": "url3632694140",
							"name": "acmeDirUrl",
							"onlyDomains": null,
							"presentable": false,
							"required": false,
							"system": false,
							"type": "url"
						},
						{
							"hidden": false,
							"id": "autodate2990389176",
							"name": "created",
							"onCreate": true,
							"onUpdate": false,
							"presentable": false,
							"system": false,
							"type": "autodate"
						},
						{
							"hidden": false,
							"id": "autodate3332085495",
							"name": "updated",
							"onCreate": true,
							"onUpdate": true,
							"presentable": false,
							"system": false,
							"type": "autodate"
						}
					],
					"id": "012d7abbod1hwvr",
					"indexes": [
						"CREATE INDEX ` + "`" + `idx_dQiYzimY7m` + "`" + ` ON ` + "`" + `acme_accounts` + "`" + ` (` + "`" + `ca` + "`" + `)",
						"CREATE INDEX ` + "`" + `idx_TjyqY6LAGa` + "`" + ` ON ` + "`" + `acme_accounts` + "`" + ` (\n  ` + "`" + `ca` + "`" + `,\n  ` + "`" + `acmeDirUrl` + "`" + `\n)",
						"CREATE UNIQUE INDEX ` + "`" + `idx_G4brUDgxzc` + "`" + ` ON ` + "`" + `acme_accounts` + "`" + ` (\n  ` + "`" + `ca` + "`" + `,\n  ` + "`" + `acmeDirUrl` + "`" + `,\n  ` + "`" + `acmeAcctUrl` + "`" + `\n)"
					],
					"name": "acme_accounts",
					"system": false,
					"type": "base"
				},
				{
					"fields": [
						{
							"autogeneratePattern": "[a-z0-9]{15}",
							"hidden": false,
							"id": "text3208210256",
							"max": 15,
							"min": 15,
							"name": "id",
							"pattern": "^[a-z0-9]+$",
							"presentable": false,
							"primaryKey": true,
							"required": true,
							"system": true,
							"type": "text"
						},
						{
							"autogeneratePattern": "",
							"hidden": false,
							"id": "8yydhv1h",
							"max": 0,
							"min": 0,
							"name": "name",
							"pattern": "",
							"presentable": false,
							"primaryKey": false,
							"required": false,
							"system": false,
							"type": "text"
						},
						{
							"autogeneratePattern": "",
							"hidden": false,
							"id": "1buzebwz",
							"max": 0,
							"min": 0,
							"name": "description",
							"pattern": "",
							"presentable": false,
							"primaryKey": false,
							"required": false,
							"system": false,
							"type": "text"
						},
						{
							"hidden": false,
							"id": "vqoajwjq",
							"maxSelect": 1,
							"name": "trigger",
							"presentable": false,
							"required": false,
							"system": false,
							"type": "select",
							"values": [
								"manual",
								"scheduled"
							]
						},
						{
							"autogeneratePattern": "",
							"hidden": false,
							"id": "8ho247wh",
							"max": 0,
							"min": 0,
							"name": "triggerCron",
							"pattern": "",
							"presentable": false,
							"primaryKey": false,
							"required": false,
							"system": false,
							"type": "text"
						},
						{
							"hidden": false,
							"id": "nq7kfdzi",
							"name": "enabled",
							"presentable": false,
							"required": false,
							"system": false,
							"type": "bool"
						},
						{
							"hidden": false,
							"id": "g9ohkk5o",
							"maxSize": 5000000,
							"name": "graphDraft",
							"presentable": false,
							"required": false,
							"system": false,
							"type": "json"
						},
						{
							"hidden": false,
							"id": "awlphkfe",
							"maxSize": 5000000,
							"name": "graphContent",
							"presentable": false,
							"required": false,
							"system": false,
							"type": "json"
						},
						{
							"hidden": false,
							"id": "2rpfz9t3",
							"name": "hasDraft",
							"presentable": false,
							"required": false,
							"system": false,
							"type": "bool"
						},
						{
							"hidden": false,
							"id": "bool3832150317",
							"name": "hasContent",
							"presentable": false,
							"required": false,
							"system": false,
							"type": "bool"
						},
						{
							"cascadeDelete": false,
							"collectionId": "qjp8lygssgwyqyz",
							"hidden": false,
							"id": "a23wkj9x",
							"maxSelect": 1,
							"minSelect": 0,
							"name": "lastRunRef",
							"presentable": false,
							"required": false,
							"system": false,
							"type": "relation"
						},
						{
							"hidden": false,
							"id": "zivdxh23",
							"maxSelect": 1,
							"name": "lastRunStatus",
							"presentable": false,
							"required": false,
							"system": false,
							"type": "select",
							"values": [
								"pending",
								"processing",
								"succeeded",
								"failed",
								"canceled"
							]
						},
						{
							"hidden": false,
							"id": "u9bosu36",
							"max": "",
							"min": "",
							"name": "lastRunTime",
							"presentable": false,
							"required": false,
							"system": false,
							"type": "date"
						},
						{
							"hidden": false,
							"id": "autodate2990389176",
							"name": "created",
							"onCreate": true,
							"onUpdate": false,
							"presentable": false,
							"system": false,
							"type": "autodate"
						},
						{
							"hidden": false,
							"id": "autodate3332085495",
							"name": "updated",
							"onCreate": true,
							"onUpdate": true,
							"presentable": false,
							"system": false,
							"type": "autodate"
						}
					],
					"id": "tovyif5ax6j62ur",
					"indexes": [],
					"name": "workflow",
					"system": false,
					"type": "base"
				},
				{
					"fields": [
						{
							"autogeneratePattern": "[a-z0-9]{15}",
							"hidden": false,
							"id": "text3208210256",
							"max": 15,
							"min": 15,
							"name": "id",
							"pattern": "^[a-z0-9]+$",
							"presentable": false,
							"primaryKey": true,
							"required": true,
							"system": true,
							"type": "text"
						},
						{
							"cascadeDelete": true,
							"collectionId": "tovyif5ax6j62ur",
							"hidden": false,
							"id": "jka88auc",
							"maxSelect": 1,
							"minSelect": 0,
							"name": "workflowRef",
							"presentable": false,
							"required": false,
							"system": false,
							"type": "relation"
						},
						{
							"cascadeDelete": true,
							"collectionId": "qjp8lygssgwyqyz",
							"hidden": false,
							"id": "relation821863227",
							"maxSelect": 1,
							"minSelect": 0,
							"name": "runRef",
							"presentable": false,
							"required": false,
							"system": false,
							"type": "relation"
						},
						{
							"autogeneratePattern": "",
							"hidden": false,
							"id": "z9fgvqkz",
							"max": 0,
							"min": 0,
							"name": "nodeId",
							"pattern": "",
							"presentable": false,
							"primaryKey": false,
							"required": false,
							"system": false,
							"type": "text"
						},
						{
							"hidden": false,
							"id": "json2239752261",
							"maxSize": 5000000,
							"name": "nodeConfig",
							"presentable": false,
							"required": false,
							"system": false,
							"type": "json"
						},
						{
							"hidden": false,
							"id": "he4cceqb",
							"maxSize": 5000000,
							"name": "outputs",
							"presentable": false,
							"required": false,
							"system": false,
							"type": "json"
						},
						{
							"hidden": false,
							"id": "2yfxbxuf",
							"name": "succeeded",
							"presentable": false,
							"required": false,
							"system": false,
							"type": "bool"
						},
						{
							"hidden": false,
							"id": "autodate2990389176",
							"name": "created",
							"onCreate": true,
							"onUpdate": false,
							"presentable": false,
							"system": false,
							"type": "autodate"
						},
						{
							"hidden": false,
							"id": "autodate3332085495",
							"name": "updated",
							"onCreate": true,
							"onUpdate": true,
							"presentable": false,
							"system": false,
							"type": "autodate"
						}
					],
					"id": "bqnxb95f2cooowp",
					"indexes": [
						"CREATE INDEX ` + "`" + `idx_BYoQPsz4my` + "`" + ` ON ` + "`" + `workflow_output` + "`" + ` (` + "`" + `workflowRef` + "`" + `)",
						"CREATE INDEX ` + "`" + `idx_O9zxLETuxJ` + "`" + ` ON ` + "`" + `workflow_output` + "`" + ` (` + "`" + `runRef` + "`" + `)",
						"CREATE INDEX ` + "`" + `idx_luac8Ul34G` + "`" + ` ON ` + "`" + `workflow_output` + "`" + ` (` + "`" + `nodeId` + "`" + `)"
					],
					"name": "workflow_output",
					"system": false,
					"type": "base"
				},
				{
					"fields": [
						{
							"autogeneratePattern": "[a-z0-9]{15}",
							"hidden": false,
							"id": "text3208210256",
							"max": 15,
							"min": 15,
							"name": "id",
							"pattern": "^[a-z0-9]+$",
							"presentable": false,
							"primaryKey": true,
							"required": true,
							"system": true,
							"type": "text"
						},
						{
							"hidden": false,
							"id": "by9hetqi",
							"maxSelect": 1,
							"name": "source",
							"presentable": false,
							"required": false,
							"system": false,
							"type": "select",
							"values": [
								"request",
								"upload"
							]
						},
						{
							"autogeneratePattern": "",
							"hidden": false,
							"id": "fugxf58p",
							"max": 0,
							"min": 0,
							"name": "subjectAltNames",
							"pattern": "",
							"presentable": false,
							"primaryKey": false,
							"required": false,
							"system": false,
							"type": "text"
						},
						{
							"autogeneratePattern": "",
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
						},
						{
							"autogeneratePattern": "",
							"hidden": false,
							"id": "plmambpz",
							"max": 100000,
							"min": 0,
							"name": "certificate",
							"pattern": "",
							"presentable": false,
							"primaryKey": false,
							"required": false,
							"system": false,
							"type": "text"
						},
						{
							"autogeneratePattern": "",
							"hidden": false,
							"id": "49qvwxcg",
							"max": 100000,
							"min": 0,
							"name": "privateKey",
							"pattern": "",
							"presentable": false,
							"primaryKey": false,
							"required": false,
							"system": false,
							"type": "text"
						},
						{
							"autogeneratePattern": "",
							"hidden": false,
							"id": "text2910474005",
							"max": 0,
							"min": 0,
							"name": "issuerOrg",
							"pattern": "",
							"presentable": false,
							"primaryKey": false,
							"required": false,
							"system": false,
							"type": "text"
						},
						{
							"autogeneratePattern": "",
							"hidden": false,
							"id": "agt7n5bb",
							"max": 100000,
							"min": 0,
							"name": "issuerCertificate",
							"pattern": "",
							"presentable": false,
							"primaryKey": false,
							"required": false,
							"system": false,
							"type": "text"
						},
						{
							"autogeneratePattern": "",
							"hidden": false,
							"id": "text4164403445",
							"max": 0,
							"min": 0,
							"name": "keyAlgorithm",
							"pattern": "",
							"presentable": false,
							"primaryKey": false,
							"required": false,
							"system": false,
							"type": "text"
						},
						{
							"hidden": false,
							"id": "v40aqzpd",
							"max": "",
							"min": "",
							"name": "validityNotBefore",
							"presentable": false,
							"required": false,
							"system": false,
							"type": "date"
						},
						{
							"hidden": false,
							"id": "zgpdby2k",
							"max": "",
							"min": "",
							"name": "validityNotAfter",
							"presentable": false,
							"required": false,
							"system": false,
							"type": "date"
						},
						{
							"autogeneratePattern": "",
							"hidden": false,
							"id": "text2045248758",
							"max": 0,
							"min": 0,
							"name": "acmeAcctUrl",
							"pattern": "",
							"presentable": false,
							"primaryKey": false,
							"required": false,
							"system": false,
							"type": "text"
						},
						{
							"exceptDomains": null,
							"hidden": false,
							"id": "ayyjy5ve",
							"name": "acmeCertUrl",
							"onlyDomains": null,
							"presentable": false,
							"required": false,
							"system": false,
							"type": "url"
						},
						{
							"exceptDomains": null,
							"hidden": false,
							"id": "3x5heo8e",
							"name": "acmeCertStableUrl",
							"onlyDomains": null,
							"presentable": false,
							"required": false,
							"system": false,
							"type": "url"
						},
						{
							"hidden": false,
							"id": "bool810050391",
							"name": "acmeRenewed",
							"presentable": false,
							"required": false,
							"system": false,
							"type": "bool"
						},
						{
							"cascadeDelete": false,
							"collectionId": "tovyif5ax6j62ur",
							"hidden": false,
							"id": "uvqfamb1",
							"maxSelect": 1,
							"minSelect": 0,
							"name": "workflowRef",
							"presentable": false,
							"required": false,
							"system": false,
							"type": "relation"
						},
						{
							"cascadeDelete": false,
							"collectionId": "qjp8lygssgwyqyz",
							"hidden": false,
							"id": "relation3917999135",
							"maxSelect": 1,
							"minSelect": 0,
							"name": "workflowRunRef",
							"presentable": false,
							"required": false,
							"system": false,
							"type": "relation"
						},
						{
							"autogeneratePattern": "",
							"hidden": false,
							"id": "uqldzldw",
							"max": 0,
							"min": 0,
							"name": "workflowNodeId",
							"pattern": "",
							"presentable": false,
							"primaryKey": false,
							"required": false,
							"system": false,
							"type": "text"
						},
						{
							"hidden": false,
							"id": "klyf4nlq",
							"max": "",
							"min": "",
							"name": "deleted",
							"presentable": false,
							"required": false,
							"system": false,
							"type": "date"
						},
						{
							"hidden": false,
							"id": "autodate2990389176",
							"name": "created",
							"onCreate": true,
							"onUpdate": false,
							"presentable": false,
							"system": false,
							"type": "autodate"
						},
						{
							"hidden": false,
							"id": "autodate3332085495",
							"name": "updated",
							"onCreate": true,
							"onUpdate": true,
							"presentable": false,
							"system": false,
							"type": "autodate"
						}
					],
					"id": "4szxr9x43tpj6np",
					"indexes": [
						"CREATE INDEX ` + "`" + `idx_Jx8TXzDCmw` + "`" + ` ON ` + "`" + `certificate` + "`" + ` (` + "`" + `workflowRef` + "`" + `)",
						"CREATE INDEX ` + "`" + `idx_2cRXqNDyyp` + "`" + ` ON ` + "`" + `certificate` + "`" + ` (` + "`" + `workflowRunRef` + "`" + `)",
						"CREATE INDEX ` + "`" + `idx_kcKpgAZapk` + "`" + ` ON ` + "`" + `certificate` + "`" + ` (` + "`" + `workflowNodeId` + "`" + `)"
					],
					"name": "certificate",
					"system": false,
					"type": "base"
				},
				{
					"fields": [
						{
							"autogeneratePattern": "[a-z0-9]{15}",
							"hidden": false,
							"id": "text3208210256",
							"max": 15,
							"min": 15,
							"name": "id",
							"pattern": "^[a-z0-9]+$",
							"presentable": false,
							"primaryKey": true,
							"required": true,
							"system": true,
							"type": "text"
						},
						{
							"cascadeDelete": true,
							"collectionId": "tovyif5ax6j62ur",
							"hidden": false,
							"id": "m8xfsyyy",
							"maxSelect": 1,
							"minSelect": 0,
							"name": "workflowRef",
							"presentable": false,
							"required": false,
							"system": false,
							"type": "relation"
						},
						{
							"hidden": false,
							"id": "qldmh0tw",
							"maxSelect": 1,
							"name": "status",
							"presentable": false,
							"required": false,
							"system": false,
							"type": "select",
							"values": [
								"pending",
								"processing",
								"succeeded",
								"failed",
								"canceled"
							]
						},
						{
							"hidden": false,
							"id": "jlroa3fk",
							"maxSelect": 1,
							"name": "trigger",
							"presentable": false,
							"required": false,
							"system": false,
							"type": "select",
							"values": [
								"manual",
								"scheduled"
							]
						},
						{
							"hidden": false,
							"id": "k9xvtf89",
							"max": "",
							"min": "",
							"name": "startedAt",
							"presentable": false,
							"required": false,
							"system": false,
							"type": "date"
						},
						{
							"hidden": false,
							"id": "3ikum7mk",
							"max": "",
							"min": "",
							"name": "endedAt",
							"presentable": false,
							"required": false,
							"system": false,
							"type": "date"
						},
						{
							"hidden": false,
							"id": "json772177811",
							"maxSize": 5000000,
							"name": "graph",
							"presentable": false,
							"required": false,
							"system": false,
							"type": "json"
						},
						{
							"autogeneratePattern": "",
							"hidden": false,
							"id": "hvebkuxw",
							"max": 20000,
							"min": 0,
							"name": "error",
							"pattern": "",
							"presentable": false,
							"primaryKey": false,
							"required": false,
							"system": false,
							"type": "text"
						},
						{
							"hidden": false,
							"id": "autodate2990389176",
							"name": "created",
							"onCreate": true,
							"onUpdate": false,
							"presentable": false,
							"system": false,
							"type": "autodate"
						},
						{
							"hidden": false,
							"id": "autodate3332085495",
							"name": "updated",
							"onCreate": true,
							"onUpdate": true,
							"presentable": false,
							"system": false,
							"type": "autodate"
						}
					],
					"id": "qjp8lygssgwyqyz",
					"indexes": [
						"CREATE INDEX ` + "`" + `idx_7ZpfjTFsD2` + "`" + ` ON ` + "`" + `workflow_run` + "`" + ` (` + "`" + `workflowRef` + "`" + `)"
					],
					"name": "workflow_run",
					"system": false,
					"type": "base"
				},
				{
					"fields": [
						{
							"autogeneratePattern": "[a-z0-9]{15}",
							"hidden": false,
							"id": "text3208210256",
							"max": 15,
							"min": 15,
							"name": "id",
							"pattern": "^[a-z0-9]+$",
							"presentable": false,
							"primaryKey": true,
							"required": true,
							"system": true,
							"type": "text"
						},
						{
							"cascadeDelete": true,
							"collectionId": "tovyif5ax6j62ur",
							"hidden": false,
							"id": "relation3371272342",
							"maxSelect": 1,
							"minSelect": 0,
							"name": "workflowRef",
							"presentable": false,
							"required": false,
							"system": false,
							"type": "relation"
						},
						{
							"cascadeDelete": true,
							"collectionId": "qjp8lygssgwyqyz",
							"hidden": false,
							"id": "relation821863227",
							"maxSelect": 1,
							"minSelect": 0,
							"name": "runRef",
							"presentable": false,
							"required": false,
							"system": false,
							"type": "relation"
						},
						{
							"autogeneratePattern": "",
							"hidden": false,
							"id": "text157423495",
							"max": 0,
							"min": 0,
							"name": "nodeId",
							"pattern": "",
							"presentable": false,
							"primaryKey": false,
							"required": false,
							"system": false,
							"type": "text"
						},
						{
							"autogeneratePattern": "",
							"hidden": false,
							"id": "text3227511481",
							"max": 0,
							"min": 0,
							"name": "nodeName",
							"pattern": "",
							"presentable": false,
							"primaryKey": false,
							"required": false,
							"system": false,
							"type": "text"
						},
						{
							"hidden": false,
							"id": "number2782324286",
							"max": null,
							"min": null,
							"name": "timestamp",
							"onlyInt": false,
							"presentable": false,
							"required": false,
							"system": false,
							"type": "number"
						},
						{
							"hidden": false,
							"id": "number760395071",
							"max": null,
							"min": null,
							"name": "level",
							"onlyInt": false,
							"presentable": false,
							"required": false,
							"system": false,
							"type": "number"
						},
						{
							"autogeneratePattern": "",
							"hidden": false,
							"id": "text3065852031",
							"max": 20000,
							"min": 0,
							"name": "message",
							"pattern": "",
							"presentable": false,
							"primaryKey": false,
							"required": false,
							"system": false,
							"type": "text"
						},
						{
							"hidden": false,
							"id": "json2918445923",
							"maxSize": 5000000,
							"name": "data",
							"presentable": false,
							"required": false,
							"system": false,
							"type": "json"
						},
						{
							"hidden": false,
							"id": "autodate2990389176",
							"name": "created",
							"onCreate": true,
							"onUpdate": false,
							"presentable": false,
							"system": false,
							"type": "autodate"
						}
					],
					"id": "pbc_1682296116",
					"indexes": [
						"CREATE INDEX ` + "`" + `idx_IOlpy6XuJ2` + "`" + ` ON ` + "`" + `workflow_logs` + "`" + ` (` + "`" + `workflowRef` + "`" + `)",
						"CREATE INDEX ` + "`" + `idx_qVlTb2yl7v` + "`" + ` ON ` + "`" + `workflow_logs` + "`" + ` (` + "`" + `runRef` + "`" + `)",
						"CREATE INDEX ` + "`" + `idx_UL4tdCXNlA` + "`" + ` ON ` + "`" + `workflow_logs` + "`" + ` (` + "`" + `nodeId` + "`" + `)"
					],
					"name": "workflow_logs",
					"system": false,
					"type": "base"
				}
			]`

			if err := app.ImportCollectionsByMarshaledJSON([]byte(jsonData), false); err != nil {
				return err
			}
		}

		// initialize superuser
		{
			collection, err := app.FindCollectionByNameOrId(core.CollectionNameSuperusers)
			if err != nil {
				return err
			}

			records, err := app.FindAllRecords(collection)
			if err != nil {
				return err
			}

			if len(records) == 0 {
				envUsername := strings.TrimSpace(os.Getenv("CERTIMATE_ADMIN_USERNAME"))
				if envUsername == "" {
					envUsername = "admin@certimate.fun"
				}

				envPassword := strings.TrimSpace(os.Getenv("CERTIMATE_ADMIN_PASSWORD"))
				if envPassword == "" {
					envPassword = "1234567890"
				}

				record := core.NewRecord(collection)
				record.Set("email", envUsername)
				record.Set("password", envPassword)
				return app.Save(record)
			}
		}

		// clean old migrations
		{
			migrations := []string{
				"1739462400_collections_snapshot.go",
				"1739462401_superusers_initial.go",
				"1740050400_upgrade.go",
				"1742209200_upgrade.go",
				"1742392800_upgrade.go",
				"1742644800_upgrade.go",
				"1743264000_upgrade.go",
				"1744192800_upgrade.go",
				"1744459000_upgrade.go",
				"1745308800_upgrade.go",
				"1745726400_upgrade.go",
				"1747314000_upgrade.go",
				"1747389600_upgrade.go",
				"1748178000_upgrade.go",
				"1748228400_upgrade.go",
				"1748959200_upgrade.go",
				"1750687200_upgrade.go",
				"1751961600_upgrade.go",
				"1753272000_v0.4.0_migrate.go",
				"1755187200_cm0.4.0_migrate.go",
				"1756296000_cm0.4.0_migrate.go",
				"1757476800_cm0.4.0_initialize.go",
				"1757476800_m0.4.0_migrate.go",
				"1757476801_m0.4.0_initialize.go",
				"1760486400_m0.4.1.go",
				"1762142400_m0.4.3.go",
				"1762516800_m0.4.4.go",
				"1763373600_m0.4.5.go",
				"1763640000_m0.4.6.go",
			}
			for _, name := range migrations {
				app.DB().NewQuery("DELETE FROM _migrations WHERE file='" + name + "'").Execute()
			}
		}

		return nil
	}, func(app core.App) error {
		return nil
	})
}
