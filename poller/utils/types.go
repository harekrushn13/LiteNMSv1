package utils

type Events struct {
	ObjectId uint32 `msgpack:"objectId" json:"objectId"`

	CounterId uint16 `msgpack:"counterId" json:"counterId"`

	Timestamp uint32 `msgpack:"timestamp" json:"timestamp"`

	Value interface{} `msgpack:"value" json:"value"`
}

type Device struct {
	ObjectID uint32 `msgpack:"object_id" json:"object_id"`

	IP string `msgpack:"ip" json:"ip"`

	IsProvisioned bool `db:"is_provisioned" json:"is_provisioned"`

	Username string `msgpack:"username" json:"username"`

	Password string `msgpack:"password" json:"password"`

	Port uint16 `msgpack:"port" json:"port"`
}
