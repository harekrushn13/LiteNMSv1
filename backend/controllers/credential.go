package controllers

import (
	. "backend/models"
	. "backend/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

type CredentialController struct {
	Service *CredentialService
}

func NewCredentialController(service *CredentialService) *CredentialController {

	return &CredentialController{Service: service}
}

func (controller *CredentialController) CreateCredential(context *gin.Context) {

	var credential Credential

	if err := context.ShouldBindJSON(&credential); err != nil {

		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	credentialID, err := controller.Service.CreateCredential(&credential)

	if err != nil {

		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	context.JSON(http.StatusCreated, gin.H{

		"credential_id": credentialID,

		"message": "Credential created successfully",
	})
}

func (controller *CredentialController) GetCredentials(context *gin.Context) {

	credentials, err := controller.Service.GetCredentials()

	if err != nil {

		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	context.JSON(http.StatusOK, credentials)
}

func (controller *CredentialController) UpdateCredential(context *gin.Context) {

	credentialID := context.Param("id")

	var credential Credential

	if err := context.ShouldBindJSON(&credential); err != nil {

		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	err := controller.Service.UpdateCredential(credentialID, &credential)

	if err != nil {

		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Credential updated successfully"})
}
