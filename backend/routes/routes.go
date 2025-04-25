package routes

import (
	. "backend/controllers"
	. "backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func InitRoutes(db *sqlx.DB, deviceChannel chan []PollerDevice, queryChannel chan QueryMap) *gin.Engine {

	router := gin.Default()

	credentialCtrl := NewCredentialController(db)

	discoveryCtrl := NewDiscoveryController(db)

	provisionCtrl := NewProvisionController(db, deviceChannel)

	queryCtrl := NewQueryController(db, queryChannel)

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

		query := v1.Group("/query")

		{
			query.POST("/", queryCtrl.FetchQuery)
		}

	}

	return router
}
