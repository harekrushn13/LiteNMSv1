package controllers

import (
	. "backend/models"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"golang.org/x/crypto/ssh"
	"net"
	"net/http"
	"sync"
	"time"
)

type DiscoveryController struct {
	DB *sqlx.DB
}

func NewDiscoveryController(db *sqlx.DB) *DiscoveryController {

	return &DiscoveryController{DB: db}
}

func (controller *DiscoveryController) CreateDiscovery(context *gin.Context) {

	var request struct {
		CredentialIDs []uint16 `json:"credential_ids" binding:"required,min=1"`

		IP string `json:"ip"`

		IPRange string `json:"ip_range"`
	}

	if err := context.ShouldBindJSON(&request); err != nil {

		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	if request.IP == "" && request.IPRange == "" {

		context.JSON(http.StatusBadRequest, gin.H{"error": "either ip or ip_range must be provided"})

		return
	}

	if request.IP != "" && request.IPRange != "" {

		context.JSON(http.StatusBadRequest, gin.H{"error": "provide either ip or ip_range, not both"})

		return
	}

	if request.IP != "" {

		if net.ParseIP(request.IP) == nil {

			context.JSON(http.StatusBadRequest, gin.H{"error": "invalid ip format"})

			return
		}
	}

	if request.IPRange != "" {

		_, _, err := net.ParseCIDR(request.IPRange)

		if err != nil {

			context.JSON(http.StatusBadRequest, gin.H{"error": "invalid ip_range format"})

			return
		}
	}

	var count int

	query := "SELECT COUNT(*) FROM credential_profile WHERE credential_id = ANY($1)"

	err := controller.DB.Get(&count, query, pq.Array(request.CredentialIDs))

	if err != nil {

		context.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify credentials"})

		return
	}

	if count != len(request.CredentialIDs) {

		context.JSON(http.StatusBadRequest, gin.H{"error": "one or more credential_ids not found"})

		return
	}

	discovery := Discovery{

		CredentialIDs: CredentialIDArray(request.CredentialIDs),

		IP: request.IP,

		IPRange: request.IPRange,

		DiscoveryStatus: Pending,
	}

	query = `INSERT INTO discovery_profile 
             (credential_id, ip, ip_range, discovery_status) 
             VALUES ($1, $2, $3, $4) 
             RETURNING discovery_id`

	var discoveryID uint16

	err = controller.DB.QueryRow(

		query,

		pq.Array(discovery.CredentialIDs),

		discovery.IP,

		discovery.IPRange,

		discovery.DiscoveryStatus,
	).Scan(&discoveryID)

	if err != nil {

		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	context.JSON(http.StatusCreated, gin.H{

		"discovery_id": discoveryID,

		"message": "Discovery created successfully",
	})

}

func (controller *DiscoveryController) StartDiscovery(c *gin.Context) {

	discoveryID := c.Param("id")

	var discovery Discovery

	err := controller.DB.Get(&discovery, `
        SELECT discovery_id, credential_id, ip, ip_range, discovery_status 
        FROM discovery_profile 
        WHERE discovery_id = $1`, discoveryID)

	if err != nil {

		c.JSON(http.StatusNotFound, gin.H{"error": "discovery not found"})

		return
	}

	//if discovery.DiscoveryStatus == Success || discovery.DiscoveryStatus == Failed {
	//
	//	c.JSON(http.StatusBadRequest, gin.H{"error": "discovery already completed"})
	//
	//	return
	//}

	var credentials []Credential

	err = controller.DB.Select(&credentials, `
        SELECT credential_id, username, password, port 
        FROM credential_profile 
        WHERE credential_id = ANY($1)`, pq.Array(discovery.CredentialIDs))

	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get credentials"})

		return
	}

	var allIPs []string

	if discovery.IP != "" {

		allIPs = []string{discovery.IP}

	} else {

		_, ipNet, err := net.ParseCIDR(discovery.IPRange)

		if err != nil {

			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ip range in discovery"})

			return
		}

		allIPs, err = getAllIps(ipNet)

		if err != nil {

			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate IP list"})

			return
		}
	}

	discoveredDevices := controller.runDiscovery(allIPs, credentials)

	_, err = controller.DB.Exec(`
        UPDATE discovery_profile 
        SET discovery_status = $1, updated_at = NOW() 
        WHERE discovery_id = $2`, Success, discoveryID)

	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update discovery status"})

		return
	}

	c.JSON(http.StatusOK, gin.H{

		"discovery_id": discoveryID,

		"discovered_devices": discoveredDevices,

		"message": "Discovery completed successfully",
	})

}

func (controller *DiscoveryController) runDiscovery(allIPs []string, credentials []Credential) []map[string]interface{} {

	type result struct {
		IP string

		CredentialID uint16

		Success bool

		Error string
	}

	results := make(chan result, len(allIPs)*len(credentials))

	var discoveredDevices []map[string]interface{}

	var wg sync.WaitGroup

	for _, cred := range credentials {

		config := &ssh.ClientConfig{

			User: cred.Username,

			Auth: []ssh.AuthMethod{

				ssh.Password(cred.Password),
			},

			HostKeyCallback: ssh.InsecureIgnoreHostKey(),

			Timeout: 5 * time.Second,
		}

		for _, ip := range allIPs {

			wg.Add(1)

			go func(ip string, cred Credential) {

				defer wg.Done()

				conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", ip, cred.Port), config)

				if err != nil {

					return
				}

				defer conn.Close()

				session, err := conn.NewSession()

				if err != nil {

					return
				}

				defer session.Close()

				_, err = session.Output("uname")

				if err != nil {

					results <- result{

						IP: ip,

						CredentialID: cred.CredentialID,

						Success: false,

						Error: err.Error(),
					}

					return
				}

				results <- result{

					IP: ip,

					CredentialID: cred.CredentialID,

					Success: true,
				}

			}(ip, cred)
		}
	}

	wg.Wait()

	close(results)

	deviceMap := make(map[string]bool)

	for i := 0; i < len(allIPs)*len(credentials); i++ {

		res := <-results

		if res.Success && !deviceMap[res.IP] {

			deviceMap[res.IP] = true

			discoveredDevices = append(discoveredDevices, map[string]interface{}{

				"ip": res.IP,

				"credential_id": res.CredentialID,
			})
		}
	}

	return discoveredDevices
}

func getAllIps(ipNet *net.IPNet) ([]string, error) {

	var ips []string

	ip := ipNet.IP

	for ip := ip.Mask(ipNet.Mask); ipNet.Contains(ip); increment(ip) {

		ips = append(ips, ip.String())
	}

	lenIPs := len(ips)

	switch {

	case lenIPs < 2:

		return ips, nil

	default:

		return ips[1 : len(ips)-1], nil
	}
}

func increment(ip net.IP) {

	for j := len(ip) - 1; j >= 0; j-- {

		ip[j]++

		if ip[j] > 0 {

			break
		}
	}
}
