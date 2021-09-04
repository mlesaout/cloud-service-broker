package passwords_test

import (
	"github.com/cloudfoundry-incubator/cloud-service-broker/db_service/models"
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/passwords"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var _ = Describe("Passwords struct", func() {
	var db *gorm.DB

	BeforeEach(func() {
		var err error
		db, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		Expect(err).NotTo(HaveOccurred())
		Expect(db.Migrator().CreateTable(&models.PasswordMetadata{})).NotTo(HaveOccurred())
	})

	Describe("SecondariesSecrets()", func() {
		It("returns the secrets for the secondaries", func() {
			const password = `[{"label":"barfoo","password":{"secret":"veryverysecretpassword"},"primary":false},{"label":"barbaz","password":{"secret":"anotherveryverysecretpassword"}},{"label":"bazquz","password":{"secret":"yetanotherveryverysecretpassword"},"primary":true}]`
			passwds, err := passwords.ProcessPasswords(password, true, db)
			Expect(err).NotTo(HaveOccurred())

			Expect(passwds.Secrets()).To(Equal([]string{"yetanotherveryverysecretpassword", "veryverysecretpassword", "anotherveryverysecretpassword"}))
		})
	})
})
