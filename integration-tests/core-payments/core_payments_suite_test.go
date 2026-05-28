package core_payments_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCorePayments(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CorePayments Suite")
}
