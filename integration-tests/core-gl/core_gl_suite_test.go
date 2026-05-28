package core_gl_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCoreGl(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CoreGl Suite")
}
