package encryption_test

import (
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/encryption"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
)

var _ = Describe("Encryption Config", func() {
	AfterEach(func() {
		viper.Reset()
	})

	Describe("GetEncryptionKey", func() {
		Describe("encryption is not enabled", func() {
			BeforeEach(func() {
				viper.Set("encryption.enabled", false)
			})
			It("should return empty key", func() {
				key, err := encryption.GetEncryptionKey()
				Expect(err).ToNot(HaveOccurred())
				Expect(key).To(BeEmpty())
			})

			It("should return error when a primary key is also provided", func() {
				viper.Set("encryption.keys", "[{\"encryption_key\": {\"secret\":\"thisisAveryLongstring\"},\"guid\":\"dae1dd13-53ed-4c90-8c11-7383b767d5c3\",\"label\":\"foo-foo\",\"primary\":true}]")

				_, err := encryption.GetEncryptionKey()
				Expect(err).To(MatchError("encryption is disabled, but a primary encryption key was provided"))
			})
		})

		Describe("encryption is enabled", func() {
			BeforeEach(func() {
				viper.Set("encryption.enabled", true)
			})

			It("should return the primary password", func() {
				viper.Set("encryption.keys", "[{\"encryption_key\": {\"secret\":\"thisisAveryLongstring\"},\"guid\":\"80e767c6-0599-11ec-b9bf-c36874088e33\",\"label\":\"foo-foo\",\"primary\":true}]")

				key, err := encryption.GetEncryptionKey()
				Expect(err).ToNot(HaveOccurred())
				Expect(key).ToNot(BeEmpty())
				Expect(key).ToNot(Equal("bar"))
			})

			It("should return the same key for the same password and label", func() {
				viper.Set("encryption.keys", "[{\"encryption_key\": {\"secret\":\"thisisAveryLongstring\"},\"guid\":\"dae1dd13-53ed-4c90-8c11-7383b767d5c3\",\"label\":\"foo-foo\",\"primary\":true}]")
				key1, err := encryption.GetEncryptionKey()
				Expect(err).ToNot(HaveOccurred())

				viper.Set("encryption.keys", "[{\"encryption_key\": {\"secret\":\"thisisAveryLongstring\"},\"guid\":\"aa13c938-04fd-11ec-9401-77c8cddeb97d\",\"label\":\"foo-foo\",\"primary\":true}]")
				key2, err := encryption.GetEncryptionKey()
				Expect(err).ToNot(HaveOccurred())

				Expect(key1).To(Equal(key2))
			})

			It("should return the different key for different label", func() {
				viper.Set("encryption.keys", "[{\"encryption_key\": {\"secret\":\"thisisAveryLongstring\"},\"guid\":\"80e767c6-0599-11ec-b9bf-c36874088e33\",\"label\":\"foo-foo-1\",\"primary\":true}]")
				key1, err := encryption.GetEncryptionKey()
				Expect(err).ToNot(HaveOccurred())

				viper.Set("encryption.keys", "[{\"encryption_key\": {\"secret\":\"thisisAveryLongstring\"},\"guid\":\"80e767c6-0599-11ec-b9bf-c36874088e33\",\"label\":\"foo-foo-2\",\"primary\":true}]")
				key2, err := encryption.GetEncryptionKey()
				Expect(err).ToNot(HaveOccurred())

				Expect(key1).ToNot(Equal(key2))
			})

			Describe("invalid encryption keys block", func() {
				It("should fail when no encryption keys are provided", func() {
					_, err := encryption.GetEncryptionKey()
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError("encryption is enabled, but there was an error validating encryption keys: but no encryption keys were provided"))
				})

				Describe("invalid encryption key", func() {
					It("should fail when guid is missing", func() {
						viper.Set("encryption.keys", "[{\"encryption_key\": {\"secret\":\"thisisAveryLongstring\"},\"label\":\"foo-foo\",\"primary\":true}]")

						_, err := encryption.GetEncryptionKey()
						Expect(err).To(HaveOccurred())
						Expect(err).To(MatchError("encryption is enabled, but there was an error validating encryption keys: field must be a UUID: Key[0].guid"))
					})

					It("should fail when label is missing", func() {
						viper.Set("encryption.keys", "[{\"encryption_key\": {\"secret\":\"thisisAveryLongstring\"},\"guid\":\"dae1dd13-53ed-4c90-8c11-7383b767d5c3\",\"primary\":true}]")

						_, err := encryption.GetEncryptionKey()
						Expect(err).To(HaveOccurred())
						Expect(err).To(MatchError("encryption is enabled, but there was an error validating encryption keys: missing field(s): Key[0].label"))
					})

					It("should fail when password is missing", func() {
						viper.Set("encryption.keys", "[{\"encryption_key\": {\"secret\":\"\"},\"guid\":\"dae1dd13-53ed-4c90-8c11-7383b767d5c3\",\"label\":\"foo-foo\",\"primary\":true}]")

						_, err := encryption.GetEncryptionKey()
						Expect(err).To(HaveOccurred())
						Expect(err).To(MatchError("encryption is enabled, but there was an error validating encryption keys: missing field(s): Key[0].encryption_key.secret"))
					})

					It("should fail when label has wrong length", func() {
						viper.Set("encryption.keys", "[{\"encryption_key\": {\"secret\":\"thisisAveryLongstring\"},\"guid\":\"dae1dd13-53ed-4c90-8c11-7383b767d5c3\",\"label\":\"foo\",\"primary\":true}]")

						_, err := encryption.GetEncryptionKey()
						Expect(err).To(HaveOccurred())
						Expect(err).To(MatchError("encryption is enabled, but there was an error validating encryption keys: field must be 5-1024 chars long: foo: Key[0].label"))
					})

					It("should fail when password has wrong length", func() {
						viper.Set("encryption.keys", "[{\"encryption_key\": {\"secret\":\"short\"},\"guid\":\"dae1dd13-53ed-4c90-8c11-7383b767d5c3\",\"label\":\"foo-foo\",\"primary\":true}]")

						_, err := encryption.GetEncryptionKey()
						Expect(err).To(HaveOccurred())
						Expect(err).To(MatchError("encryption is enabled, but there was an error validating encryption keys: field must be 20-1024 chars long: short: Key[0].encryption_key.secret"))
					})
				})

				It("should fail when no encryption key is marked as primary", func() {
					viper.Set("encryption.keys", "[{\"encryption_key\": {\"secret\":\"thisisAveryLongstring\"},\"guid\":\"dae1dd13-53ed-4c90-8c11-7383b767d5c3\",\"label\":\"foo-foo\",\"primary\":false}]")

					_, err := encryption.GetEncryptionKey()
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError("encryption is enabled, but there was an error validating encryption keys: no encryption key is marked as primary"))
				})

				It("should fail when two encryption keys are marked as primary", func() {
					viper.Set("encryption.keys", "[{\"encryption_key\": {\"secret\":\"thisisAveryLongstring\"},\"guid\":\"dae1dd13-53ed-4c90-8c11-7383b767d5c3\",\"label\":\"label-1\",\"primary\":true},{\"encryption_key\": {\"secret\":\"thisIs-anotherlongstring\"},\"guid\":\"80e767c6-0599-11ec-b9bf-c36874088e33\",\"label\":\"label-2\",\"primary\":true}]")

					_, err := encryption.GetEncryptionKey()
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError("encryption is enabled, but there was an error validating encryption keys: several encryption keys are marked as primary"))
				})

				It("should fail when two encryption keys have same labels", func() {
					viper.Set("encryption.keys", "[{\"encryption_key\": {\"secret\":\"thisisAveryLongstring\"},\"guid\":\"dae1dd13-53ed-4c90-8c11-7383b767d5c3\",\"label\":\"same-same\",\"primary\":false},{\"encryption_key\": {\"secret\":\"thisIs-anotherlongstring\"},\"guid\":\"80e767c6-0599-11ec-b9bf-c36874088e33\",\"label\":\"same-same\",\"primary\":true}]")

					_, err := encryption.GetEncryptionKey()
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError("encryption is enabled, but there was an error validating encryption keys: duplicated value, must be unique: same-same: Key[1].label"))
				})

				It("should fail when two encryption keys have same IDs", func() {
					viper.Set("encryption.keys", "[{\"encryption_key\": {\"secret\":\"thisisAveryLongstring\"},\"guid\":\"dae1dd13-53ed-4c90-8c11-7383b767d5c3\",\"label\":\"foo-foo\",\"primary\":true},{\"encryption_key\": {\"secret\":\"thisIs-anotherlongstring\"},\"guid\":\"dae1dd13-53ed-4c90-8c11-7383b767d5c3\",\"label\":\"wow-wow\",\"primary\":false}]")

					_, err := encryption.GetEncryptionKey()
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError("encryption is enabled, but there was an error validating encryption keys: duplicated value, must be unique: dae1dd13-53ed-4c90-8c11-7383b767d5c3: Key[1].guid"))
				})

			})
		})
	})
})
