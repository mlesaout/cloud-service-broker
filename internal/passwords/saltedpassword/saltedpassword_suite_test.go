package saltedpassword_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSaltedpassword(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Salted Password Suite")
}
