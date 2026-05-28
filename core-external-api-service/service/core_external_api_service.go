package service

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
	"github.com/openreserveio/core/core-external-api-service/generated/glmodel"
	"github.com/openreserveio/core/core-external-api-service/generated/model"
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
	engine := gin.Default()
	engine.GET("/", func(c *gin.Context) {
		c.String(200, "Hello World!")
	})
	apiService.Gin = engine

	return &apiService, nil
}

func (apiService *CoreExternalApiService) Start() {
	err := apiService.Gin.Run(fmt.Sprintf("%s:%s", apiService.HttpListenHost, apiService.HttpListenPort))
	log.Infof("Ending Core External API Service:  %v", err)
}
