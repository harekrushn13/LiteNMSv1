package controllers

import (
	. "backend/models"
	. "backend/utils"
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"log"
	"net/http"
)

type ProvisionController struct {
	DB *sqlx.DB

	deviceChannel chan []PollerDevice
}

func NewProvisionController(db *sqlx.DB, deviceChannel chan []PollerDevice) *ProvisionController {

	return &ProvisionController{
		DB: db,

		deviceChannel: deviceChannel,
	}
}

type ProvisionDevice struct {
	IP string `json:"ip" binding:"required"`

	CredentialID uint16 `json:"credential_id" binding:"required"`
}

func (pc *ProvisionController) ProvisionDevice(c *gin.Context) {

	var request struct {
		DiscoveryID uint16 `json:"discovery_id" binding:"required"`

		Devices []ProvisionDevice `json:"devices" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	var discoveryStatus DiscoveryStatus

	err := pc.DB.Get(&discoveryStatus, `
		SELECT discovery_status 
		FROM discovery_profile 
		WHERE discovery_id = $1`, request.DiscoveryID)

	if err != nil {

		c.JSON(http.StatusNotFound, gin.H{"error": "discovery not found"})

		return
	}

	var count int

	err = pc.DB.Get(&count, `
		SELECT COUNT(*) 
		FROM credential_profile 
		WHERE credential_id = ANY($1)`,
		pq.Array(getCredentialIDs(request.Devices)))

	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify credentials"})

		return
	}

	tx, err := pc.DB.Beginx()

	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to begin transaction"})

		return
	}

	defer tx.Rollback()

	var provisionedDevices []Provision

	for _, device := range request.Devices {

		var objectID uint32

		err := tx.QueryRowx(`
			SELECT object_id FROM provision 
			WHERE ip = $1 AND discovery_id = $2
		`, device.IP, request.DiscoveryID).Scan(&objectID)

		if err != nil && !errors.Is(err, sql.ErrNoRows) {

			tx.Rollback()

			c.JSON(http.StatusInternalServerError, gin.H{

				"error": "failed to check existing provision",

				"details": err.Error(),
			})

			return
		}

		provision := Provision{
			IP: device.IP,

			CredentialID: device.CredentialID,

			DiscoveryID: request.DiscoveryID,

			IsProvisioned: false,
		}

		if err == nil {

			provision.ObjectID = objectID

			provisionedDevices = append(provisionedDevices, provision)

			continue
		}

		err = tx.QueryRow(`
    					INSERT INTO provision 
    					(ip, credential_id, discovery_id, is_provisioned)
    					VALUES ($1, $2, $3, $4)
    					RETURNING object_id`,
			provision.IP,
			provision.CredentialID,
			provision.DiscoveryID,
			provision.IsProvisioned,
		).Scan(&provision.ObjectID)

		if err != nil {
			tx.Rollback()

			c.JSON(http.StatusInternalServerError, gin.H{

				"error": "failed to create provision record",

				"details": err.Error(),
			})

			return
		}

		provisionedDevices = append(provisionedDevices, provision)
	}

	if err := tx.Commit(); err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit transaction"})

		return
	}

	go func() {

		if err := pc.sendToPoller(provisionedDevices); err != nil {

			log.Printf("Failed to notify poller service: %v", err)

			return

		}

	}()

	c.JSON(http.StatusCreated, gin.H{

		"message": "Devices queued for provisioning",

		"count": len(provisionedDevices),

		"devices": provisionedDevices,
	})
}

func getCredentialIDs(devices []ProvisionDevice) []uint16 {

	ids := make([]uint16, len(devices))

	for i, d := range devices {

		ids[i] = d.CredentialID
	}

	return ids
}

func (pc *ProvisionController) sendToPoller(provisionedDevices []Provision) error {

	var pollerDevices []PollerDevice

	for _, device := range provisionedDevices {

		var cred Credential

		err := pc.DB.Get(&cred, `
            SELECT username, password, port 
            FROM credential_profile 
            WHERE credential_id = $1`, device.CredentialID)

		if err != nil {

			continue
		}

		pollerDevices = append(pollerDevices, PollerDevice{
			ObjectID: device.ObjectID,

			IP: device.IP,

			CredentialID: device.CredentialID,

			DiscoveryID: device.DiscoveryID,

			Username: cred.Username,

			Password: cred.Password,

			Port: cred.Port,
		})

		pc.deviceChannel <- pollerDevices
	}

	return nil
}
