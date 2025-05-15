package controllers

import (
	. "backend/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

type ProvisionController struct {
	service *ProvisionService
}

func NewProvisionController(service *ProvisionService) *ProvisionController {

	return &ProvisionController{service: service}
}

func (controller *ProvisionController) ProvisionDevice(context *gin.Context) {

	var request ProvisionRequest

	if err := context.ShouldBindJSON(&request); err != nil {

		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	devices, err := controller.service.HandleProvisioning(request)

	if err != nil {

		context.JSON(err.Code, gin.H{"error": err.Message, "details": err.Details})

		return
	}

	context.JSON(http.StatusCreated, gin.H{

		"message": "Devices queued for provisioning",

		"count": len(devices),

		"devices": devices,
	})
}

func (controller *ProvisionController) GetProvisionedDevices(context *gin.Context) {

	discoveryID := context.Param("id")

	var discovery interface{}

	err := controller.service.DB.Get(&discovery, `SELECT discovery_id FROM discovery_profile WHERE discovery_id = $1`, discoveryID)

	if err != nil {

		context.JSON(http.StatusNotFound, gin.H{"error": "discovery not found"})

		return
	}

	var devices []struct {
		ObjectID uint32 `db:"object_id" json:"object_id"`

		IP string `db:"ip" json:"ip"`

		CredentialID uint16 `db:"credential_id" json:"credential_id"`

		IsProvisioned bool `db:"is_provisioned" json:"is_provisioned"`
	}

	err = controller.service.DB.Select(&devices, `SELECT object_id, ip, credential_id, is_provisioned FROM provision WHERE discovery_id = $1`, discoveryID)

	if err != nil {

		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	context.JSON(http.StatusOK, devices)
}
