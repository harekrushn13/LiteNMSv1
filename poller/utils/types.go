package utils

type Events struct {
	ObjectId uint32

	CounterId uint16

	Timestamp uint32

	Value interface{}
}

type Device struct {
	ObjectID uint32 `json:"object_id"`

	IP string `json:"ip"`

	CredentialID uint16 `json:"credential_id"`

	DiscoveryID uint16 `json:"discovery_id"`

	Username string `json:"username"`

	Password string `json:"password"`

	Port uint16 `json:"port"`
}
