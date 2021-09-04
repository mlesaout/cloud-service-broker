package encryption_test

import (
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/encryption"
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/encryption/compoundencryptor"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/encryption/gcmencryptor"
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/encryption/noopencryptor"
)

var _ = Describe("Encryption", func() {
	When("valid key is provided", func() {
		It("returns a GCM Encryptor", func() {
			const validKey = "one-key-here-with-32-bytes-in-it"
			Expect(encryption.EncryptorFromKeys(validKey)).To(BeAssignableToTypeOf(gcmencryptor.GCMEncryptor{}))
		})
	})

	When("no key provided", func() {
		It("returns a no-op encryptor", func() {
			Expect(encryption.EncryptorFromKeys()).To(BeAssignableToTypeOf(noopencryptor.NoopEncryptor{}))
		})
	})

	When("blank key provided", func() {
		It("returns a no-op encryptor", func() {
			Expect(encryption.EncryptorFromKeys(" ")).To(BeAssignableToTypeOf(noopencryptor.NoopEncryptor{}))
		})
	})

	When("multiple keys provided", func() {
		It("returns a compound encryptor", func() {
			const validKey1 = "one-key-here-with-32-bytes-in-it"
			const validKey2 = "another-32-great-bytes-inside-it"
			Expect(encryption.EncryptorFromKeys(validKey1, " ", validKey2)).To(BeAssignableToTypeOf(compoundencryptor.CompoundEncryptor{}))
		})
	})
})
