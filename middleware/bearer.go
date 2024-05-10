package middleware

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"qrcode/database"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type KeycloakKey struct {
	Kid     string   `json:"kid"`
	Kty     string   `json:"kty"`
	Alg     string   `json:"alg"`
	Use     string   `json:"use"`
	N       string   `json:"n"`
	E       string   `json:"e"`
	X5c     []string `json:"x5c"`
	X5t     string   `json:"x5t"`
	X5tS256 string   `json:"x5t#S256"`
}

type KeycloakKeys struct {
	Keys []KeycloakKey `json:"keys"`
}

func BearerTokenMiddleware() gin.HandlerFunc {
	return func(request *gin.Context) {
		// Get the token from the Authorization header
		tokenString := request.GetHeader("Authorization")
		if tokenString == "" {
			request.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Bearer token is required",
			})
			return
		}

		// Remove "Bearer " from tokenString
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")

		// Parse the token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			keycloakEndpoint := os.Getenv("KEYCLOAK_BASE_URL") + "/protocol/openid-connect/certs"
			response, err := http.Get(keycloakEndpoint)
			if err != nil {
				return nil, fmt.Errorf("cannot get keycloak certs: %v", err)
			}
			defer response.Body.Close()

			var keys KeycloakKeys
			body, err := io.ReadAll(response.Body)
			if err != nil {
				return nil, fmt.Errorf("cannot read response body: %v", err)
			}

			json.Unmarshal(body, &keys)

			for _, key := range keys.Keys {
				if key.Kid == token.Header["kid"] {
					certPEM := "-----BEGIN CERTIFICATE-----\n" + key.X5c[0] + "\n-----END CERTIFICATE-----"
					block, _ := pem.Decode([]byte(certPEM))
					if block == nil {
						return nil, fmt.Errorf("failed to parse certificate PEM")
					}
					cert, err := x509.ParseCertificate(block.Bytes)
					if err != nil {
						return nil, fmt.Errorf("failed to parse certificate: %v", err)
					}
					rsaPublicKey, ok := cert.PublicKey.(*rsa.PublicKey)
					if !ok {
						return nil, fmt.Errorf("key is not of type *rsa.PublicKey")
					}
					return rsaPublicKey, nil
				}
			}

			return []byte(""), fmt.Errorf("cannot find key with kid: %v", token.Header["kid"])
		})

		if err != nil {
			request.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Cannot parse bearer token",
				"err":   err.Error(),
			})
			return
		}

		// Check if the token is valid
		if !token.Valid {
			request.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Bearer token not valid",
			})
			return
		}

		user, userErr := database.InsertUserIfNotExists(token.Claims.(jwt.MapClaims)["preferred_username"].(string))
		if userErr != nil {
			request.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Cannot insert user",
				"err":   userErr.Error(),
			})
			return
		}

		request.Set("user", user)

		log.Println("next")
		// Continue
		request.Next()
	}
}
