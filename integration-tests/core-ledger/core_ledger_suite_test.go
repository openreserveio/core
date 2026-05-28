package core_ledger_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCoreLedger(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CoreLedger Suite")
}
