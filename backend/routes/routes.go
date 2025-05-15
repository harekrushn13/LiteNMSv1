package routes

import (
	. "backend/controllers"
	. "backend/service"
	. "backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func InitRoutes(db *sqlx.DB, deviceChannel chan []PollerDevice, queryChannel chan QueryMap) *gin.Engine {

	router := gin.Default()

	credentialService := NewCredentialService(db)

	credentialCtrl := NewCredentialController(credentialService)

	discoveryService := NewDiscoveryService(db)

	discoveryCtrl := NewDiscoveryController(discoveryService)

	provisionService := NewProvisionService(db, deviceChannel)

	provisionCtrl := NewProvisionController(provisionService)

	queryCtrl := NewQueryController(db, queryChannel)

	v1 := router.Group("/lnms")

	{

		credentials := v1.Group("/credentials")

		{
			credentials.POST("/", credentialCtrl.CreateCredential)

			credentials.GET("/", credentialCtrl.GetCredentials)
		}

		discoveries := v1.Group("/discoveries")

		{
			discoveries.POST("/", discoveryCtrl.CreateDiscovery)

			discoveries.GET("/:id", discoveryCtrl.StartDiscovery)

			discoveries.GET("/", discoveryCtrl.GetDiscoveries)
		}

		provisions := v1.Group("/provisions")

		{
			provisions.POST("/", provisionCtrl.ProvisionDevice)

			provisions.GET("/:id", provisionCtrl.GetProvisionedDevices)
		}

		query := v1.Group("/query")

		{
			query.POST("/", queryCtrl.FetchQuery)
		}

	}

	return router
}
