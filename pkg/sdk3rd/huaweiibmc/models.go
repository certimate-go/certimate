package huaweiibmc

type Entity struct {
	ID          string `json:"Id,omitempty"`
	ODataID     string `json:"@odata.id,omitempty"`
	ODataType   string `json:"@odata.type,omitempty"`
	Name        string `json:"Name,omitempty"`
	Description string `json:"Description"`
}
