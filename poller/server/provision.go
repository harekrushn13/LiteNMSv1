package server

import (
	"github.com/gin-gonic/gin"
	"net/http"
	. "poller/polling"
	"sync"
)

func StartProvisionListener(deviceChan chan<- []Device, wg *sync.WaitGroup) {

	wg.Add(1)

	defer wg.Done()

	router := gin.Default()

	router.POST("/lnms/provision", func(c *gin.Context) {

		var devices []Device

		if err := c.ShouldBindJSON(&devices); err != nil {

			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

			return
		}

		for _, device := range devices {

			if device.IP == "" || device.Username == "" || device.Password == "" || device.Port == 0 {

				c.JSON(http.StatusBadRequest, gin.H{"error": "missing required fields"})

				return
			}
		}

		deviceChan <- devices

		c.JSON(http.StatusOK, gin.H{

			"message": "Devices received successfully",

			"count": len(devices),
		})
	})

	if err := router.Run(":8081"); err != nil {

		panic(err)
	}
}
