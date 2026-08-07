package certificate

type Certificate struct {
	Id           string `json:"certificate-id"`
	Name         string `json:"name"`
	Comment      string `json:"comment"`
	ValidityFrom string `json:"certificate-validity-from"`
	ValidityTo   string `json:"certificate-validity-to"`
	Serial       string `json:"certificate-serial"`
}
