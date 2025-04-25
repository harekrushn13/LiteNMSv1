package models

import "time"

type Credential struct {
	CredentialID uint16 `db:"credential_id" json:"credential_id"`

	Username string `db:"username" json:"username"`

	Password string `db:"password" json:"password"`

	Port uint16 `db:"port" json:"port"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`

	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
