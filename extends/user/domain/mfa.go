package domain

type MFAStatus struct {
	TOTPEnabled  bool     `json:"totp_enabled"`
	SMSEnabled   bool     `json:"sms_enabled"`
	EmailEnabled bool     `json:"email_enabled"`
	Enabled      bool     `json:"enabled"`
	Phone        string   `json:"phone"`
	HasSecret    bool     `json:"has_secret"`
	BackupCount  int      `json:"backup_count"`
	Methods      []string `json:"methods"`
}

type MFASecretInfo struct {
	Secret string `json:"secret"`
	URL    string `json:"otpauth_url"`
}
