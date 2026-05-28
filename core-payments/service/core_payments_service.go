package service

import (
	log "github.com/sirupsen/logrus"
)

type CorePaymentsService struct {
}

func NewCorePaymentsService(masterConfig map[string]interface{}) (*CorePaymentsService, error) {

	log.Infof("Starting Core Payments Service with config: %v", masterConfig)
	return nil, nil

}
