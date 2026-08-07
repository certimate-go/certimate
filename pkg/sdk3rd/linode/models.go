package linode

type Domain struct {
	ID          *int      `json:"id,omitempty"`
	Domain      *string   `json:"domain,omitempty"`
	Type        *string   `json:"type,omitempty"`
	Group       *string   `json:"group,omitempty"`
	Status      *string   `json:"status,omitempty"`
	Description *string   `json:"description,omitempty"`
	SOAEmail    *string   `json:"soa_email,omitempty"`
	RetrySec    *int      `json:"retry_sec,omitempty"`
	MasterIPs   []*string `json:"master_ips,omitempty"`
	AXfrIPs     []*string `json:"axfr_ips,omitempty"`
	Tags        []*string `json:"tags,omitempty"`
	ExpireSec   *int      `json:"expire_sec,omitempty"`
	RefreshSec  *int      `json:"refresh_sec,omitempty"`
	TTLSec      *int      `json:"ttl_sec,omitempty"`
}

type DomainRecord struct {
	ID       *int    `json:"id,omitempty"`
	Type     *string `json:"type,omitempty"`
	Name     *string `json:"name,omitempty"`
	Target   *string `json:"target,omitempty"`
	Priority *int    `json:"priority,omitempty"`
	Weight   *int    `json:"weight,omitempty"`
	Port     *int    `json:"port,omitempty"`
	Service  *string `json:"service,omitempty"`
	Protocol *string `json:"protocol,omitempty"`
	TTLSec   *int    `json:"ttl_sec,omitempty"`
	Tag      *string `json:"tag,omitempty"`
}
