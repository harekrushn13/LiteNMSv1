package routes

import (
	. "backend/controllers"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func SetupRoutes(db *sqlx.DB) *gin.Engine {

	router := gin.Default()

	credentialCtrl := NewCredentialController(db)

	discoveryCtrl := NewDiscoveryController(db)

	pollerAPI := "http://localhost:8081"

	provisionCtrl := NewProvisionController(db, pollerAPI)

	v1 := router.Group("/lnms")

	{

		credentials := v1.Group("/credentials")

		{
			credentials.POST("/", credentialCtrl.CreateCredential)
		}

		discoveries := v1.Group("/discoveries")

		{
			discoveries.POST("/", discoveryCtrl.CreateDiscovery)

			discoveries.GET("/:id", discoveryCtrl.StartDiscovery)
		}

		provisions := v1.Group("/provisions")

		{
			provisions.POST("/", provisionCtrl.ProvisionDevice)
		}

	}

	return router
}
