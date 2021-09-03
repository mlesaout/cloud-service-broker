package dbencryptor

import (
	"context"

	"github.com/cloudfoundry-incubator/cloud-service-broker/db_service/models"
	"gorm.io/gorm"
)

// EncryptDB encrypts the database with the primary encryptor (which can be the No-op encryptor)
func EncryptDB(ctx context.Context, db *gorm.DB) error {
	var batch []models.ProvisionRequestDetails
	db.FindInBatches(&batch, 100, func(tx *gorm.DB, batchNumber int) error {
		for _, entry := range batch {
			details, err := entry.GetRequestDetails()
			if err != nil {
				return err
			}

			if err := entry.SetRequestDetails(details); err != nil {
				return err
			}
		}

		return tx.Save(&batch).Error
	})

	return nil
}
