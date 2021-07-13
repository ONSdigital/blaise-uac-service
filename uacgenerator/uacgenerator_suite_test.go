package uacgenerator_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestUacgenerator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Uacgenerator Suite")
}
