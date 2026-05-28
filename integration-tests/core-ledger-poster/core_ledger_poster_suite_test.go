package core_ledger_poster_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCoreLedgerPoster(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CoreLedgerPoster Suite")
}
