package utils

type PollerDevice struct {
	ObjectID uint32 `json:"object_id"`

	IP string `json:"ip"`

	CredentialID uint16 `json:"credential_id"`

	DiscoveryID uint16 `json:"discovery_id"`

	Username string `json:"username"`

	Password string `json:"password"`

	Port uint16 `json:"port"`
}
