package models_test

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"io"
	"reflect"
	"strings"

	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/encryption"

	"github.com/cloudfoundry-incubator/cloud-service-broker/db_service/models"
	"github.com/cloudfoundry-incubator/cloud-service-broker/db_service/models/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func newKey() [32]byte {
	dbKey := make([]byte, 32)
	io.ReadFull(rand.Reader, dbKey)
	return sha256.Sum256(dbKey)
}

var _ = Describe("Db", func() {
	var encryptor models.Encryptor

	AfterEach(func() {
		models.SetEncryptor(nil)
	})

	Describe("ServiceInstanceDetails", func() {
		Context("GCM encryptor", func() {
			BeforeEach(func() {
				key := newKey()
				encryptor = encryption.NewGCMEncryptor(&key)
				models.SetEncryptor(encryptor)
			})

			Describe("SetOtherDetails", func() {
				It("marshalls json content", func() {
					otherDetails := map[string]interface{}{
						"some": []interface{}{"json", "blob", "here"},
					}
					details := models.ServiceInstanceDetails{}

					err := details.SetOtherDetails(otherDetails)

					Expect(err).NotTo(HaveOccurred())
					decryptedDetails, _ := encryptor.Decrypt([]byte(details.OtherDetails))
					Expect(string(decryptedDetails)).To(Equal(`{"some":["json","blob","here"]}`))
				})

				It("marshalls nil into json null", func() {
					details := models.ServiceInstanceDetails{}

					err := details.SetOtherDetails(nil)

					Expect(err).NotTo(HaveOccurred())
					decryptedDetails, _ := encryptor.Decrypt([]byte(details.OtherDetails))
					Expect(string(decryptedDetails)).To(Equal("null"))
				})
			})

			Describe("GetOtherDetails", func() {
				It("decrypts and unmarshalls json content", func() {
					encryptedDetails, _ := encryptor.Encrypt([]byte(`{"some":["json","blob","here"]}`))
					serviceInstanceDetails := models.ServiceInstanceDetails{
						OtherDetails: string(encryptedDetails),
					}

					var actualOtherDetails map[string]interface{}
					err := serviceInstanceDetails.GetOtherDetails(&actualOtherDetails)

					Expect(err).NotTo(HaveOccurred())

					var arrayOfInterface []interface{}
					arrayOfInterface = append(arrayOfInterface, "json", "blob", "here")
					expectedOtherDetails := map[string]interface{}{
						"some": arrayOfInterface,
					}
					Expect(actualOtherDetails).To(Equal(expectedOtherDetails))
				})

				It("returns nil if is empty", func() {
					serviceInstanceDetails := models.ServiceInstanceDetails{}

					var actualOtherDetails map[string]interface{}
					err := serviceInstanceDetails.GetOtherDetails(&actualOtherDetails)

					Expect(err).NotTo(HaveOccurred())

					Expect(actualOtherDetails).To(BeNil())
				})

			})

			It("Can decrypt what it had previously encrypted", func() {
				serviceInstanceDetails := models.ServiceInstanceDetails{}
				input := map[string]interface{}{
					"some": []string{"json", "blob", "here"},
				}
				serviceInstanceDetails.SetOtherDetails(input)

				var actualOtherDetails map[string]interface{}
				err := serviceInstanceDetails.GetOtherDetails(&actualOtherDetails)

				Expect(err).NotTo(HaveOccurred())

				var arrayOfInterface []interface{}
				arrayOfInterface = append(arrayOfInterface, "json", "blob", "here")
				expectedOtherDetails := map[string]interface{}{
					"some": arrayOfInterface,
				}

				Expect(actualOtherDetails).To(Equal(expectedOtherDetails))
			})
		})

		Context("Noop encryptor", func() {
			BeforeEach(func() {
				encryptor = encryption.NewNoopEncryptor()
				models.SetEncryptor(encryptor)
			})

			Describe("SetOtherDetails", func() {
				It("marshalls json content", func() {
					otherDetails := map[string]interface{}{
						"some": []interface{}{"json", "blob", "here"},
					}
					details := models.ServiceInstanceDetails{}

					err := details.SetOtherDetails(otherDetails)

					Expect(err).NotTo(HaveOccurred())
					Expect(details.OtherDetails).To(Equal(`{"some":["json","blob","here"]}`))
				})

				It("marshalls nil into json null", func() {
					details := models.ServiceInstanceDetails{}

					err := details.SetOtherDetails(nil)

					Expect(err).NotTo(HaveOccurred())
					Expect(details.OtherDetails).To(Equal("null"))
				})
			})

			Describe("GetOtherDetails", func() {
				It("unmarshalls json content", func() {
					serviceInstanceDetails := models.ServiceInstanceDetails{
						OtherDetails: `{"some":["json","blob","here"]}`,
					}

					var actualOtherDetails map[string]interface{}
					err := serviceInstanceDetails.GetOtherDetails(&actualOtherDetails)

					Expect(err).NotTo(HaveOccurred())

					var arrayOfInterface []interface{}
					arrayOfInterface = append(arrayOfInterface, "json", "blob", "here")
					expectedOtherDetails := map[string]interface{}{
						"some": arrayOfInterface,
					}
					Expect(actualOtherDetails).To(Equal(expectedOtherDetails))
				})

				It("returns nil if is empty", func() {
					serviceInstanceDetails := models.ServiceInstanceDetails{}

					var actualOtherDetails map[string]interface{}
					err := serviceInstanceDetails.GetOtherDetails(&actualOtherDetails)

					Expect(err).NotTo(HaveOccurred())

					Expect(actualOtherDetails).To(BeNil())
				})

			})
		})

		Describe("errors", func() {
			Describe("SetOtherDetails", func() {
				It("returns an error if it cannot marshall", func() {

					details := models.ServiceInstanceDetails{}

					err := details.SetOtherDetails(struct {
						F func()
					}{F: func() {}})

					Expect(err).ToNot(BeNil(), "Should have returned an error")
					Expect(details.OtherDetails).To(BeEmpty())
				})

				Context("When there are errors while encrypting", func() {
					BeforeEach(func() {
						fakeEncryptor := &fakes.FakeEncryptor{}
						fakeEncryptor.EncryptReturns(nil, errors.New("some error"))

						encryptor = fakeEncryptor
						models.SetEncryptor(encryptor)
					})

					It("returns an error", func() {
						details := models.ServiceInstanceDetails{}
						var someDetails []byte

						err := details.SetOtherDetails(someDetails)

						Expect(err).To(MatchError("some error"))
					})
				})
			})

			Describe("GetOtherDetails", func() {
				Context("When there are errors while unmarshalling", func() {
					BeforeEach(func() {
						fakeEncryptor := &fakes.FakeEncryptor{}
						fakeEncryptor.DecryptReturns([]byte(`{"some":"badjson", "here"]}`), nil)

						encryptor = fakeEncryptor
						models.SetEncryptor(encryptor)
					})

					It("returns an error", func() {
						serviceInstanceDetails := models.ServiceInstanceDetails{
							OtherDetails: "something not nil",
						}

						var actualOtherDetails map[string]interface{}
						err := serviceInstanceDetails.GetOtherDetails(&actualOtherDetails)

						Expect(err).To(MatchError(ContainSubstring("invalid character")))

						Expect(actualOtherDetails).To(BeNil())
					})
				})

				Context("When there are errors while decrypting", func() {
					BeforeEach(func() {
						fakeEncryptor := &fakes.FakeEncryptor{}
						fakeEncryptor.DecryptReturns(nil, errors.New("some error"))

						encryptor = fakeEncryptor
						models.SetEncryptor(encryptor)
					})

					It("returns an error", func() {
						details := models.ServiceInstanceDetails{
							OtherDetails: "something not nil",
						}

						var actualOtherDetails map[string]interface{}
						err := details.GetOtherDetails(&actualOtherDetails)

						Expect(err).To(MatchError("some error"))
					})
				})
			})
		})
	})

	Describe("ProvisionRequestDetails", func() {
		Context("GCM encryptor", func() {
			BeforeEach(func() {
				key := newKey()
				encryptor = encryption.NewGCMEncryptor(&key)
				models.SetEncryptor(encryptor)
			})

			Describe("SetRequestDetails", func() {
				It("encrypts and sets the details", func() {
					details := models.ProvisionRequestDetails{}

					rawMessage := []byte(`{"key":"value"}`)
					details.SetRequestDetails(rawMessage)

					decryptedDetails, _ := encryptor.Decrypt([]byte(details.RequestDetails))
					Expect(string(decryptedDetails)).To(Equal(`{"key":"value"}`))
				})

				It("converts nil to the empty string", func() {
					details := models.ProvisionRequestDetails{}

					details.SetRequestDetails(nil)

					decryptedDetails, _ := encryptor.Decrypt([]byte(details.RequestDetails))
					Expect(decryptedDetails).To(BeEmpty())
				})

				It("converts empty array to the empty string", func() {
					details := models.ProvisionRequestDetails{}
					var rawMessage []byte
					details.SetRequestDetails(rawMessage)

					decryptedDetails, _ := encryptor.Decrypt([]byte(details.RequestDetails))
					Expect(decryptedDetails).To(BeEmpty())
				})
			})

			Describe("GetRequestDetails", func() {
				It("gets as RawMessage", func() {
					encryptedDetails, _ := encryptor.Encrypt([]byte(`{"some":["json","blob","here"]}`))
					requestDetails := models.ProvisionRequestDetails{
						RequestDetails: string(encryptedDetails),
					}

					details, err := requestDetails.GetRequestDetails()

					rawMessage := json.RawMessage(`{"some":["json","blob","here"]}`)

					Expect(err).To(BeNil())
					Expect(details).To(Equal(rawMessage))
				})
			})

			It("Can decrypt what it had previously encrypted", func() {
				details := models.ProvisionRequestDetails{}

				rawMessage := json.RawMessage(`{"key":"value"}`)
				details.SetRequestDetails(rawMessage)

				actualDetails, err := details.GetRequestDetails()

				Expect(err).To(BeNil())
				Expect(actualDetails).To(Equal(rawMessage))
			})
		})

		Context("Noop encryptor", func() {
			BeforeEach(func() {
				encryptor = encryption.NewNoopEncryptor()
				models.SetEncryptor(encryptor)
			})

			Describe("SetRequestDetails", func() {
				It("sets the details", func() {
					details := models.ProvisionRequestDetails{}

					rawMessage := []byte(`{"key":"value"}`)
					details.SetRequestDetails(rawMessage)

					Expect(details.RequestDetails).To(Equal("{\"key\":\"value\"}"))
				})

				It("converts nil to the empty string", func() {
					details := models.ProvisionRequestDetails{}

					details.SetRequestDetails(nil)

					Expect(details.RequestDetails).To(BeEmpty())
				})

				It("converts empty array to the empty string", func() {
					details := models.ProvisionRequestDetails{}
					var rawMessage []byte
					details.SetRequestDetails(rawMessage)

					Expect(details.RequestDetails).To(BeEmpty())
				})
			})

			Describe("GetRequestDetails", func() {
				It("gets as RawMessage", func() {
					requestDetails := models.ProvisionRequestDetails{
						RequestDetails: `{"some":["json","blob","here"]}`,
					}

					details, err := requestDetails.GetRequestDetails()

					rawMessage := json.RawMessage(`{"some":["json","blob","here"]}`)

					Expect(err).To(BeNil())
					Expect(details).To(Equal(rawMessage))
				})
			})
		})

		Describe("errors", func() {
			Context("SetRequestDetails", func() {
				BeforeEach(func() {
					fakeEncryptor := &fakes.FakeEncryptor{}
					fakeEncryptor.EncryptReturns(nil, errors.New("some error"))

					encryptor = fakeEncryptor
					models.SetEncryptor(encryptor)
				})

				It("returns an error when there are errors while encrypting", func() {
					details := models.ProvisionRequestDetails{}
					var rawMessage []byte

					err := details.SetRequestDetails(rawMessage)

					Expect(err).To(MatchError("some error"))
				})
			})

			Context("GetRequestDetails", func() {
				BeforeEach(func() {
					fakeEncryptor := &fakes.FakeEncryptor{}
					fakeEncryptor.DecryptReturns(nil, errors.New("some error"))

					encryptor = fakeEncryptor
					models.SetEncryptor(encryptor)
				})

				It("returns an error when there are errors while decrypting", func() {
					requestDetails := models.ProvisionRequestDetails{
						RequestDetails: "some string",
					}

					details, err := requestDetails.GetRequestDetails()

					Expect(err).To(MatchError("some error"))
					Expect(details).To(BeNil())

				})
			})

		})
	})

	Describe("ConfigureEncryption", func() {
		Context("No key provided", func() {
			When("Key is empty", func() {
				It("Skips encryption", func() {
					encryptor := models.ConfigureEncryption("")

					Expect(encryptor).To(Equal(encryption.NewNoopEncryptor()))
				})
			})

			When("Key is blank", func() {
				It("Skips encryption", func() {
					encryptor := models.ConfigureEncryption("    \t   \n")

					Expect(encryptor).To(Equal(encryption.NewNoopEncryptor()))
				})
			})
		})

		Context("Key provided", func() {
			When("Key is valid", func() {
				It("Sets up encryptor with the key", func() {
					encryptor := models.ConfigureEncryption("one-key-here-with-32-bytes-in-it")

					Expect(reflect.TypeOf(encryptor).Name()).To(Equal("GCMEncryptor"))
					gcmEncryptor, _ := encryptor.(encryption.GCMEncryptor)
					Expect(strings.TrimSpace(string(gcmEncryptor.Key[:]))).To(Equal("one-key-here-with-32-bytes-in-it"))
				})
			})

			When("Key has surrounding spaces", func() {
				It("skips encryption", func() {
					encryptor := models.ConfigureEncryption("\t  one-key-here  \n")

					Expect(encryptor).To(Equal(encryption.NewNoopEncryptor()))
				})
			})
		})
	})
})
