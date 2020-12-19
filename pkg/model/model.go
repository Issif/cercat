package model

// Result represents a catched certificate
type Result struct {
	Domain          string   `json:"domain"`
	IDN             string   `json:"IDN,omitempty"`
	UnicodeIDN      string   `json:"UnicodeIDN,omitempty"`
	SAN             []string `json:"SAN"`
	Issuer          string   `json:"issuer"`
	Addresses       []string `json:"Addresses"`
	Attack          string   `json:"Attack"`
	ProtectedDomain string   `json:"ProtectedDomain"`
}

// Certificate represents a certificate from CertStream
type Certificate struct {
	MessageType string `json:"message_type"`
	Data        Data   `json:"data"`
}

// Data represents data field for a certificate from CertStream
type Data struct {
	UpdateType string            `json:"update_type"`
	LeafCert   LeafCert          `json:"leaf_cert"`
	Chain      []LeafCert        `json:"chain"`
	CertIndex  float32           `json:"cert_index"`
	Seen       float32           `json:"seen"`
	Source     map[string]string `json:"source"`
}

// LeafCert represents leaf_cert field from CertStream
type LeafCert struct {
	Subject      map[string]string      `json:"subject"`
	Extensions   map[string]interface{} `json:"extensions"`
	NotBefore    float32                `json:"not_before"`
	NotAfter     float32                `json:"not_after"`
	SerialNumber string                 `json:"serial_number"`
	FingerPrint  string                 `json:"fingerprint"`
	AsDer        string                 `json:"as_der"`
	AllDomains   []string               `json:"all_domains"`
}
