package apisix

type Certificate struct {
	ID            *string           `json:"id,omitempty"`
	Status        *int32            `json:"status,omitempty"`
	Certificate   *string           `json:"cert,omitempty"`
	PrivateKey    *string           `json:"key,omitempty"`
	SNIs          []*string         `json:"snis,omitempty"`
	Type          *string           `json:"type,omitempty"`
	ValidityStart *int64            `json:"validity_start,omitempty"`
	ValidityEnd   *int64            `json:"validity_end,omitempty"`
	Labels        map[string]string `json:"labels,omitempty"`
}
