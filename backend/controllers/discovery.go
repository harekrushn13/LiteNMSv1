package controllers

import (
	. "backend/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

type DiscoveryController struct {
	Service *DiscoveryService
}

func NewDiscoveryController(service *DiscoveryService) *DiscoveryController {

	return &DiscoveryController{Service: service}
}

func (controller *DiscoveryController) CreateDiscovery(context *gin.Context) {

	var request DiscoveryRequest

	if err := context.ShouldBindJSON(&request); err != nil {

		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	discoveryID, err := controller.Service.CreateDiscovery(request)

	if err != nil {

		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	context.JSON(http.StatusCreated, gin.H{

		"discovery_id": discoveryID,

		"message": "Discovery created successfully",
	})

}

func (controller *DiscoveryController) StartDiscovery(context *gin.Context) {

	discoveryID := context.Param("id")

	result, err := controller.Service.StartDiscovery(discoveryID)

	if err != nil {

		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	context.JSON(http.StatusOK, result)

}
