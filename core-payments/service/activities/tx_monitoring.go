package activities

import (
	"context"

	"github.com/openreserveio/core/core-payments/pmtmodel"
)

type RiskScore struct {
	Score int64
}

func (act *PaymentActivity) GetTransactionMonitoringRisk(ctx context.Context, payment pmtmodel.Payment) (RiskScore, error) {

	// for now, to trigger the hold
	if payment.TargetAmount > 10000000 {
		return RiskScore{Score: 100}, nil
	}

	return RiskScore{Score: 0}, nil

}

func (act *PaymentActivity) DetermineRiskTolerance(ctx context.Context, payment pmtmodel.Payment, score RiskScore) (bool, error) {

	if score.Score > 80 {
		return true, nil
	}

	return false, nil

}
