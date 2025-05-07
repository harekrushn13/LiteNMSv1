package service

import (
	. "backend/models"
	"github.com/jmoiron/sqlx"
)

type CredentialService struct {
	DB *sqlx.DB
}

func NewCredentialService(db *sqlx.DB) *CredentialService {

	return &CredentialService{DB: db}
}

func (service *CredentialService) CreateCredential(credential *Credential) (uint16, error) {

	query := `INSERT INTO credential_profile (username, password, port) VALUES ($1, $2, $3) RETURNING credential_id`

	var credentialID uint16

	err := service.DB.QueryRow(query, credential.Username, credential.Password, credential.Port).Scan(&credentialID)

	if err != nil {

		return 0, err
	}

	return credentialID, nil
}
