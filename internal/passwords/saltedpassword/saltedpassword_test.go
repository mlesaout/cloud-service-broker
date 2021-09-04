package saltedpassword_test

import (
	"strings"

	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/encryption/gcmencryptor"

	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/passwords/saltedpassword"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Salted Password", func() {
	It("can be created successfully", func() {
		_, err := saltedpassword.New("label", "reallyreallygoodpassword", []byte("one-salt-here-with-32bytes-in-it"))
		Expect(err).NotTo(HaveOccurred())
	})

	It("has an Encryptor", func() {
		sp, err := saltedpassword.New("label", "reallyreallygoodpassword", []byte("one-salt-here-with-32bytes-in-it"))
		Expect(err).NotTo(HaveOccurred())

		Expect(sp.Encryptor).NotTo(BeNil())
		Expect(sp.Encryptor).To(BeAssignableToTypeOf(gcmencryptor.GCMEncryptor{}))
		Expect(sp.Encryptor.Encrypt([]byte("foo"))).NotTo(Equal("foo"))
		Expect(sp.Encryptor.Decrypt("9m7DhwzZIoFMSAV53E3Kia821GrL1mdP3in3h6Zrdg==")).To(Equal([]byte("foo")))
	})

	When("the salt is not 32 bytes", func() {
		It("returns an error", func() {
			_, err := saltedpassword.New("label", "reallyreallygoodpassword", []byte("one-salt-here-with-33-bytes-in-it"))
			Expect(err).To(MatchError("salt must be 32 bytes"))
		})
	})

	When("the password is outside the length range", func() {
		It("returns an error", func() {
			_, err := saltedpassword.New("label", "012345678912345678", []byte("one-salt-here-with-32bytes-in-it"))
			Expect(err).To(MatchError("password must be at least 20 characters"))

			_, err = saltedpassword.New("label", strings.Repeat("a", 1025), []byte("one-salt-here-with-32bytes-in-it"))
			Expect(err).To(MatchError("password must not be more than 1024 characters"))
		})
	})
})
