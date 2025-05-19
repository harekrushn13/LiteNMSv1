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

	devices, err := controller.service.GetProvisionedDevices(discoveryID)

	if err != nil {

		context.JSON(err.Code, gin.H{"error": err.Message, "details": err.Details})

		return
	}

	context.JSON(http.StatusOK, devices)
}
