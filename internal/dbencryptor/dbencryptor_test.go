package dbencryptor_test

import (
	"context"

	"github.com/cloudfoundry-incubator/cloud-service-broker/db_service"
	"github.com/cloudfoundry-incubator/cloud-service-broker/db_service/models"
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/dbencryptor"
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/encryption"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pborman/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var _ = Describe("EncryptDB", func() {
	var (
		db                      *gorm.DB
		provisionRequestDetails models.ProvisionRequestDetails
	)

	BeforeEach(func() {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		Expect(err).NotTo(HaveOccurred())
		Expect(db_service.RunMigrations(db)).NotTo(HaveOccurred())

		db_service.DbConnection = db
		models.SetEncryptor(models.ConfigureEncryption(""))

		provisionRequestDetails = models.ProvisionRequestDetails{ServiceInstanceId: uuid.New()}
		Expect(provisionRequestDetails.SetRequestDetails([]byte(`{"a":"secret"}`))).NotTo(HaveOccurred())
		Expect(db_service.CreateProvisionRequestDetails(context.TODO(), &provisionRequestDetails)).NotTo(HaveOccurred())
	})

	const serviceInstanceFKQuery = "service_instance_id = ?"

	findRecord := func(dest interface{}, query, guid string) {
		err := db.Where(query, guid).First(dest).Error
		ExpectWithOffset(1, err).NotTo(HaveOccurred())
	}

	persistedRequestDetails := func(serviceInstanceGUID string) string {
		record := models.ProvisionRequestDetails{}
		findRecord(&record, serviceInstanceFKQuery, serviceInstanceGUID)
		return record.RequestDetails
	}

	It("encrypts the database", func() {
		models.SetEncryptor(dbencryptor.NewCompoundEncryptor(
			models.ConfigureEncryption("one-key-here-with-32-bytes-in-it"),
			encryption.NoopEncryptor{},
		))

		Expect(persistedRequestDetails(provisionRequestDetails.ServiceInstanceId)).To(Equal(`{"a":"secret"}`))

		By("running the encryption")
		Expect(dbencryptor.EncryptDB(context.TODO(), db)).NotTo(HaveOccurred())

		Expect(persistedRequestDetails(provisionRequestDetails.ServiceInstanceId)).NotTo(Equal(`{"a":"secret"}`))
	})
})
