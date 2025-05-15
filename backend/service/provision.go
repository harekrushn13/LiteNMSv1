package service

import (
	. "backend/models"
	. "backend/utils"
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"net/http"
)

type ProvisionService struct {
	DB *sqlx.DB

	DeviceChannel chan []PollerDevice
}

func NewProvisionService(db *sqlx.DB, ch chan []PollerDevice) *ProvisionService {

	return &ProvisionService{

		DB: db,

		DeviceChannel: ch,
	}
}

type ProvisionDevice struct {
	IP string `json:"ip" binding:"required"`

	CredentialID uint16 `json:"credential_id" binding:"required"`
}

type ProvisionRequest struct {
	DiscoveryID uint16 `json:"discovery_id" binding:"required"`

	Devices []ProvisionDevice `json:"devices" binding:"required,min=1"`
}

type APIError struct {
	Code int

	Message string

	Details string
}

func (service *ProvisionService) HandleProvisioning(req ProvisionRequest) ([]Provision, *APIError) {

	var discoveryStatus DiscoveryStatus

	if err := service.DB.Get(&discoveryStatus, `SELECT discovery_status FROM discovery_profile WHERE discovery_id = $1`, req.DiscoveryID); err != nil {

		return nil, &APIError{Code: http.StatusNotFound, Message: "discovery not found"}
	}

	var count int

	err := service.DB.Get(&count, `SELECT COUNT(*) FROM credential_profile WHERE credential_id = ANY($1)`, pq.Array(getCredentialIDs(req.Devices)))

	if err != nil {

		return nil, &APIError{Code: http.StatusInternalServerError, Message: "failed to verify credentials"}
	}

	tx, err := service.DB.Beginx()

	if err != nil {

		return nil, &APIError{Code: http.StatusInternalServerError, Message: "failed to begin transaction"}
	}

	defer tx.Rollback()

	var provisionedDevices []Provision

	for _, device := range req.Devices {

		var provision Provision

		err := tx.QueryRowx(`
			SELECT object_id, is_provisioned FROM provision 
			WHERE ip = $1 AND discovery_id = $2`,
			device.IP, req.DiscoveryID).Scan(&provision.ObjectID, &provision.IsProvisioned)

		if err != nil && !errors.Is(err, sql.ErrNoRows) {

			return nil, &APIError{

				Code: http.StatusInternalServerError,

				Message: "failed to check existing provision",

				Details: err.Error(),
			}
		}

		provision.IP = device.IP

		provision.CredentialID = device.CredentialID

		provision.DiscoveryID = req.DiscoveryID

		if err == nil {

			provision.IsProvisioned = !provision.IsProvisioned

			_, err = tx.Exec(`
				UPDATE provision 
				SET is_provisioned = $1 
				WHERE object_id = $2`,
				provision.IsProvisioned, provision.ObjectID)

			if err != nil {

				return nil, &APIError{

					Code: http.StatusInternalServerError,

					Message: "failed to update provision record",

					Details: err.Error(),
				}
			}

			provisionedDevices = append(provisionedDevices, provision)

			continue
		}

		if err := tx.QueryRow(`
			INSERT INTO provision (ip, credential_id, discovery_id, is_provisioned)
			VALUES ($1, $2, $3, $4)
			RETURNING object_id`,
			provision.IP, provision.CredentialID, provision.DiscoveryID, provision.IsProvisioned,
		).Scan(&provision.ObjectID); err != nil {

			return nil, &APIError{

				Code: http.StatusInternalServerError,

				Message: "failed to create provision record",

				Details: err.Error(),
			}
		}

		provisionedDevices = append(provisionedDevices, provision)

	}

	if err := tx.Commit(); err != nil {

		return nil, &APIError{Code: http.StatusInternalServerError, Message: "failed to commit transaction"}
	}

	go service.sendToPoller(provisionedDevices)

	return provisionedDevices, nil
}

func getCredentialIDs(devices []ProvisionDevice) []uint16 {

	ids := make([]uint16, len(devices))

	for i, d := range devices {

		ids[i] = d.CredentialID
	}
	return ids
}

func (service *ProvisionService) sendToPoller(devices []Provision) {

	var pollerDevices []PollerDevice

	for _, device := range devices {

		var cred Credential

		if err := service.DB.Get(&cred, `SELECT username, password, port FROM credential_profile WHERE credential_id = $1`, device.CredentialID); err != nil {

			continue
		}

		pollerDevices = append(pollerDevices, PollerDevice{
			ObjectID: device.ObjectID,

			IP: device.IP,

			IsProvisioned: device.IsProvisioned,

			Username: cred.Username,

			Password: cred.Password,

			Port: cred.Port,
		})
	}

	if len(pollerDevices) > 0 {

		service.DeviceChannel <- pollerDevices
	}
}
