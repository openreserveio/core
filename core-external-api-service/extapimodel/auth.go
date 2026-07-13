package extapimodel

import "time"

type CoreAuthClaims struct {
	Audience          []string  `json:"aud"`
	ExpiresAt         time.Time `json:"exp"`
	IssuedAt          time.Time `json:"iat"`
	Issuer            string    `json:"iss"`
	Name              string    `json:"name"`
	NotBefore         time.Time `json:"nbf"`
	PreferredUsername string    `json:"preferred_username"`
	Scope             string    `json:"scp"`
	Subject           string    `json:"sub"`
	Uti               string    `json:"uti"`
	TokenVersion      string    `json:"ver"`
}
