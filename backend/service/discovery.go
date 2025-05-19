package service

import (
	. "backend/models"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"golang.org/x/crypto/ssh"
	"net"
	"sync"
	"time"
)

type DiscoveryService struct {
	DB *sqlx.DB
}

func NewDiscoveryService(db *sqlx.DB) *DiscoveryService {

	return &DiscoveryService{DB: db}
}

type DiscoveryRequest struct {
	CredentialIDs []uint16 `json:"credential_ids" binding:"required,min=1"`

	IP string `json:"ip"`

	IPRange string `json:"ip_range"`
}

func (service *DiscoveryService) CreateDiscovery(req DiscoveryRequest) (uint16, error) {

	if err := validateDiscoveryInput(req); err != nil {

		return 0, err
	}

	var count int

	query := "SELECT COUNT(*) FROM credential_profile WHERE credential_id = ANY($1)"

	err := service.DB.Get(&count, query, pq.Array(req.CredentialIDs))

	if err != nil || count != len(req.CredentialIDs) {

		return 0, fmt.Errorf("some credential IDs are invalid")
	}

	discovery := Discovery{

		CredentialIDs: CredentialIDArray(req.CredentialIDs),

		IP: req.IP,

		IPRange: req.IPRange,

		DiscoveryStatus: Pending,
	}

	query = `INSERT INTO discovery_profile (credential_id, ip, ip_range, discovery_status) VALUES ($1, $2, $3, $4) RETURNING discovery_id`

	var discoveryID uint16

	err = service.DB.QueryRow(query, pq.Array(discovery.CredentialIDs), discovery.IP, discovery.IPRange, discovery.DiscoveryStatus).Scan(&discoveryID)

	if err != nil {

		return 0, err
	}

	return discoveryID, nil
}

func (service *DiscoveryService) StartDiscovery(id string) (map[string]interface{}, error) {

	var discovery Discovery

	err := service.DB.Get(&discovery, `SELECT discovery_id, credential_id, ip, ip_range, discovery_status FROM discovery_profile WHERE discovery_id = $1`, id)

	if err != nil {

		return nil, fmt.Errorf("discovery not found")
	}

	var credentials []Credential

	err = service.DB.Select(&credentials, `SELECT credential_id, username, password, port FROM credential_profile WHERE credential_id = ANY($1)`, pq.Array(discovery.CredentialIDs))

	if err != nil {

		return nil, err
	}

	allIPs, err := generateIPs(discovery)

	if err != nil {

		return nil, err
	}

	discovered := runDiscovery(allIPs, credentials)

	_, err = service.DB.Exec(`UPDATE discovery_profile SET discovery_status = $1, updated_at = NOW() WHERE discovery_id = $2`, Success, id)

	if err != nil {

		return nil, err
	}

	return map[string]interface{}{

		"discovery_id": id,

		"discovered_devices": discovered,

		"discovery_count": len(discovered),

		"message": "Discovery completed successfully",
	}, nil
}

func (service *DiscoveryService) UpdateDiscovery(discoveryID string, req DiscoveryRequest) error {

	if err := validateDiscoveryInput(req); err != nil {

		return err
	}

	var count int

	query := "SELECT COUNT(*) FROM credential_profile WHERE credential_id = ANY($1)"

	err := service.DB.Get(&count, query, pq.Array(req.CredentialIDs))

	if err != nil || count != len(req.CredentialIDs) {

		return fmt.Errorf("some credential IDs are invalid")
	}

	query = `UPDATE discovery_profile 
	         SET credential_id = $1, ip = $2, ip_range = $3, updated_at = NOW() 
	         WHERE discovery_id = $4`

	_, err = service.DB.Exec(query, pq.Array(req.CredentialIDs), req.IP, req.IPRange, discoveryID)

	if err != nil {

		return fmt.Errorf("failed to update discovery: %w", err)
	}

	return nil
}

func runDiscovery(allIPs []string, credentials []Credential) []map[string]interface{} {

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

			Auth: []ssh.AuthMethod{ssh.Password(cred.Password)},

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

					results <- result{IP: ip, CredentialID: cred.CredentialID, Success: false, Error: err.Error()}

					return
				}

				results <- result{IP: ip, CredentialID: cred.CredentialID, Success: true}

			}(ip, cred)

		}
	}

	wg.Wait()

	close(results)

	deviceMap := make(map[string]bool)

	for res := range results {

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

func validateDiscoveryInput(req DiscoveryRequest) error {

	if req.IP == "" && req.IPRange == "" {

		return fmt.Errorf("either ip or ip_range must be provided")
	}

	if req.IP != "" && req.IPRange != "" {

		return fmt.Errorf("provide either ip or ip_range, not both")
	}

	if req.IP != "" && net.ParseIP(req.IP) == nil {

		return fmt.Errorf("invalid ip format")

	}

	if req.IPRange != "" {

		if _, _, err := net.ParseCIDR(req.IPRange); err != nil {

			return fmt.Errorf("invalid ip_range format")
		}
	}

	if len(req.CredentialIDs) != len(checkUniqueIDs(req.CredentialIDs)) {

		return fmt.Errorf("credential_ids must contain unique values")
	}

	return nil
}

func checkUniqueIDs(ids []uint16) []uint16 {

	seen := make(map[uint16]bool)

	var result []uint16

	for _, id := range ids {

		if !seen[id] {

			seen[id] = true

			result = append(result, id)
		}
	}

	return result
}

func generateIPs(discovery Discovery) ([]string, error) {

	if discovery.IP != "" {

		return []string{discovery.IP}, nil
	}

	_, ipNet, err := net.ParseCIDR(discovery.IPRange)

	if err != nil {

		return nil, err
	}

	return getAllIps(ipNet)
}

func getAllIps(ipNet *net.IPNet) ([]string, error) {

	var ips []string

	for ip := ipNet.IP.Mask(ipNet.Mask); ipNet.Contains(ip); increment(ip) {

		ips = append(ips, ip.String())
	}

	if len(ips) <= 2 {

		return ips, nil
	}

	return ips[1 : len(ips)-1], nil
}

func increment(ip net.IP) {

	for j := len(ip) - 1; j >= 0; j-- {

		ip[j]++

		if ip[j] > 0 {

			break
		}
	}
}
