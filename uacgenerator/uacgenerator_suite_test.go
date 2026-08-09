package uacgenerator_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestUacService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "UacService Suite")
}
