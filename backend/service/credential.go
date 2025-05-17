package service

import (
	. "backend/models"
	"errors"
	"github.com/jmoiron/sqlx"
)

type CredentialService struct {
	DB *sqlx.DB
}

func NewCredentialService(db *sqlx.DB) *CredentialService {

	return &CredentialService{DB: db}
}

func (service *CredentialService) CreateCredential(credential *Credential) (uint16, error) {

	if err := validateCredential(credential); err != nil {

		return 0, err
	}

	var exists bool

	query := `SELECT EXISTS (
		SELECT 1 FROM credential_profile 
		WHERE username = $1 AND password = $2 AND port = $3
	)`

	err := service.DB.Get(&exists, query, credential.Username, credential.Password, credential.Port)

	if err != nil {

		return 0, err
	}

	if exists {

		return 0, errors.New("credential already exists with the same username, password, and port")
	}

	query = `INSERT INTO credential_profile (username, password, port) VALUES ($1, $2, $3) RETURNING credential_id`

	var credentialID uint16

	err = service.DB.QueryRow(query, credential.Username, credential.Password, credential.Port).Scan(&credentialID)

	if err != nil {

		return 0, err
	}

	return credentialID, nil
}

func (service *CredentialService) GetCredentials() ([]Credential, error) {

	var credentials []Credential

	query := `SELECT credential_id, username, password, port FROM credential_profile`

	err := service.DB.Select(&credentials, query)

	if err != nil {

		return nil, err
	}

	return credentials, nil
}

func (service *CredentialService) UpdateCredential(id string, credential *Credential) error {

	if err := validateCredential(credential); err != nil {

		return err
	}

	var exists bool

	query := `SELECT EXISTS(SELECT 1 FROM credential_profile WHERE credential_id = $1)`

	err := service.DB.Get(&exists, query, id)

	if err != nil {

		return err
	}

	if !exists {

		return errors.New("credential not found")
	}

	query = `UPDATE credential_profile SET username = $1, password = $2, port = $3 WHERE credential_id = $4`

	_, err = service.DB.Exec(query, credential.Username, credential.Password, credential.Port, id)

	return err
}

func validateCredential(credential *Credential) error {

	if credential.Username == "" {

		return errors.New("username is required")
	}

	if credential.Password == "" {

		return errors.New("password is required")

	}

	if credential.Port == 0 {

		return errors.New("port must be greater than 0")
	}

	return nil
}
