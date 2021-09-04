package passwords

import (
	"github.com/cloudfoundry-incubator/cloud-service-broker/db_service/models"
	"gorm.io/gorm"
)

type passwordMetadata struct {
	Label   string
	Salt    []byte
	Canary  string
	Primary bool
}

func savePasswordMetadata(db *gorm.DB, p passwordMetadata) error {
	return db.Create(&models.PasswordMetadata{
		Label:   p.Label,
		Salt:    p.Salt,
		Canary:  p.Canary,
		Primary: p.Primary,
	}).Error
}

func findPasswordMetadataForLabel(db *gorm.DB, label string) (passwordMetadata, bool, error) {
	return findPasswordMetadata(db, "label = ?", label)
}

func findPasswordMetadataForPrimary(db *gorm.DB) (passwordMetadata, bool, error) {
	return findPasswordMetadata(db, `"primary" = true`)
}

func findPasswordMetadata(db *gorm.DB, query interface{}, args ...interface{}) (passwordMetadata, bool, error) {
	var receiver []models.PasswordMetadata
	if err := db.Where(query, args...).Find(&receiver).Error; err != nil {
		return passwordMetadata{}, false, err
	}

	if len(receiver) == 0 {
		return passwordMetadata{}, false, nil
	}

	return passwordMetadata{
		Label:   receiver[0].Label,
		Salt:    receiver[0].Salt,
		Canary:  receiver[0].Canary,
		Primary: receiver[0].Primary,
	}, true, nil
}
