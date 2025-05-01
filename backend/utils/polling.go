package utils

type PollerDevice struct {
	ObjectID uint32 `msgpack:"object_id" json:"object_id"`

	IP string `msgpack:"ip" json:"ip"`

	CredentialID uint16 `msgpack:"credential_id" json:"credential_id"`

	DiscoveryID uint16 `msgpack:"discovery_id" json:"discovery_id"`

	Username string `msgpack:"username" json:"username"`

	Password string `msgpack:"password" json:"password"`

	Port uint16 `msgpack:"port" json:"port"`
}
