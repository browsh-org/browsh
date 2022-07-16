package browsh

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestBrowshUnits(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Unit test")
}
