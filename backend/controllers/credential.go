package controllers

import (
	. "backend/models"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"net/http"
)

type CredentialController struct {
	DB *sqlx.DB
}

func NewCredentialController(db *sqlx.DB) *CredentialController {

	return &CredentialController{DB: db}
}

func (controller *CredentialController) CreateCredential(context *gin.Context) {

	var credential Credential

	if err := context.ShouldBindJSON(&credential); err != nil {

		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	query := `INSERT INTO credential_profile (username, password, port) VALUES ($1, $2, $3) RETURNING credential_id`

	var credentialID uint16

	err := controller.DB.QueryRow(query, credential.Username, credential.Password, credential.Port).Scan(&credentialID)

	if err != nil {

		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	context.JSON(http.StatusCreated, gin.H{

		"credential_id": credentialID,

		"message": "Credential created successfully",
	})
}
