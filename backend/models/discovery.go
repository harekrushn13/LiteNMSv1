package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"github.com/lib/pq"
	"time"
)

type DiscoveryStatus string

const (
	Pending DiscoveryStatus = "Pending"

	Success DiscoveryStatus = "Success"

	Failed DiscoveryStatus = "Failed"
)

type Discovery struct {
	DiscoveryID uint16 `db:"discovery_id" json:"discovery_id"`

	CredentialIDs CredentialIDArray `db:"credential_id" json:"credential_id"`

	IP string `db:"ip" json:"ip"`

	IPRange string `db:"ip_range" json:"ip_range"`

	DiscoveryStatus DiscoveryStatus `db:"discovery_status" json:"discovery_status"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`

	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type CredentialIDArray []uint16

func (a *CredentialIDArray) Scan(value interface{}) error {

	var intArray pq.Int64Array

	if err := intArray.Scan(value); err == nil {

		*a = make(CredentialIDArray, len(intArray))

		for i, v := range intArray {

			(*a)[i] = uint16(v)
		}

		return nil
	}

	bytes, ok := value.([]byte)

	if !ok {

		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, a)
}

func (a CredentialIDArray) Value() (driver.Value, error) {

	intArray := make(pq.Int64Array, len(a))

	for i, v := range a {

		intArray[i] = int64(v)
	}

	return intArray.Value()
}
