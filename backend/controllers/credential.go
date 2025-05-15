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

	var credentials []struct {
		CredentialID uint16 `db:"credential_id" json:"credential_id"`

		Username string `db:"username" json:"username"`

		Password string `db:"password" json:"password"`

		Port uint16 `db:"port" json:"port"`
	}

	err := controller.Service.DB.Select(&credentials, `SELECT credential_id, username, password, port FROM credential_profile`)

	if err != nil {

		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	context.JSON(http.StatusOK, credentials)
}
