package models

import "time"

type Provision struct {
	ObjectID uint32 `db:"object_id" json:"object_id"`

	IP string `db:"ip" json:"ip"`

	CredentialID uint16 `db:"credential_id" json:"credential_id"`

	DiscoveryID uint16 `db:"discovery_id" json:"discovery_id"`

	IsProvisioned bool `db:"is_provisioned" json:"is_provisioned"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
