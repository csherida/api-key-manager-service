package domain

import "time"

type ApiKey struct {
	ApiId            string     `json:"api_id"`
	PrivateKey       string     `json:"private_key"` // Ideally we store the public key and provide the private key
	OrganizationName string     `json:"organization_name"`
	ExpirationDate   *time.Time `json:"expiration_date"`
}
