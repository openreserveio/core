package service

import (
	"context"
	"fmt"
	"os"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"github.com/openreserveio/core/core-external-api-service/extapimodel"
	"github.com/openreserveio/core/core-external-api-service/generated/glmodel"
	"github.com/openreserveio/core/core-external-api-service/generated/model"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type CoreExternalApiService struct {
	GLClient         glmodel.GeneralLedgerServiceClient
	CoreLedgerClient model.CoreLedgerServiceClient
	Gin              *gin.Engine
	HttpListenHost   string
	HttpListenPort   string
}

func NewCoreExternalApiService(httpListenHost string, httpListenPort string, coreLedgerUrl string, glUrl string) (*CoreExternalApiService, error) {

	apiService := CoreExternalApiService{
		HttpListenHost: httpListenHost,
		HttpListenPort: httpListenPort,
	}

	// Setup Core Ledger Client
	clConn, err := grpc.NewClient(coreLedgerUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Unable to create Core Ledger Client:  %v", err)
		return nil, err
	}
	apiService.CoreLedgerClient = model.NewCoreLedgerServiceClient(clConn)

	// Setup General Ledger Client
	glConn, err := grpc.NewClient(glUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Unable to create General Ledger Client:  %v", err)
		return nil, err
	}
	apiService.GLClient = glmodel.NewGeneralLedgerServiceClient(glConn)

	// Setup Gin
	oidcProvider, err := oidc.NewProvider(context.Background(), "https://burke.delta.openreserve.io:28443/realms/openreserveio/.well-known/openid-configuration")
	if err != nil {
		log.Fatalf("Unable to create OIDC Provider:  %v", err)
		os.Exit(1)
	}
	oauthConfig := oauth2.Config{
		ClientID:     "core-external-api-service",
		ClientSecret: "siJD1nzf58QN96en1FdRragaS9NaiTSWTgMKVHrhxhYeSxOQlUH9eWDaEQEtD7QTYEbrAIH0CuY6IrTMHrOaPV",
		RedirectURL:  redirectURL,

		// Discovery returns the OAuth2 endpoints.
		Endpoint: oidcProvider.Endpoint(),

		// "openid" is a required scope for OpenID Connect flows.
		Scopes: []string{oidc.ScopeOpenID},
	}

	// Create an ID Token verifier.
	idTokenVerifier := oidcProvider.Verifier(&oidc.Config{ClientID: "core-external-api-service"})

	engine := gin.Default()

	engine.GET("/", func(c *gin.Context) {
		c.String(200, "Hello World!")
	})

	protectedGroup := engine.Group("/protected", oidcHandler)
	protectedGroup.GET("/boo", func(c *gin.Context) {
		c.String(200, "Auth Thing Here")
	})
	apiService.Gin = engine

	return &apiService, nil
}

func (apiService *CoreExternalApiService) Start() {
	err := apiService.Gin.Run(fmt.Sprintf("%s:%s", apiService.HttpListenHost, apiService.HttpListenPort))
	log.Infof("Ending Core External API Service:  %v", err)
}

func (apiService *CoreExternalApiService) validateAuthClaims(claims *extapimodel.CoreAuthClaims) error {

	log.Infof("Validate Auth Claims: %v", claims)
	return nil

}
