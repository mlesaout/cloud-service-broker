package passwords

import (
	"fmt"

	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/passwords/saltedpassword"

	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/passwords/parser"
	"gorm.io/gorm"
)

func ProcessPasswords(input string, encryptionEnabled bool, db *gorm.DB) (Passwords, error) {
	parsedPasswords, err := parse(input, encryptionEnabled)
	if err != nil {
		return Passwords{}, err
	}

	var result Passwords
	labels := make(map[string]struct{})
	for _, p := range parsedPasswords {
		entry, err := consolidate(p, db)
		if err != nil {
			return Passwords{}, err
		}

		if p.Primary {
			result.Primary = entry
		} else {
			result.Secondaries = append(result.Secondaries, entry)
		}

		labels[p.Label] = struct{}{}
	}

	previousPrimary, found, err := findPasswordMetadataForPrimary(db)
	if err != nil {
		return Passwords{}, err
	}
	if found {
		if _, ok := labels[previousPrimary.Label]; !ok {
			return Passwords{}, fmt.Errorf("the previous primary password labeled %q was not specified", previousPrimary.Label)
		}

		result.ChangedPrimary = previousPrimary.Label != result.Primary.Label
	}

	return result, nil
}

func consolidate(parsed parser.PasswordEntry, db *gorm.DB) (saltedpassword.SaltedPassword, error) {
	loaded, ok, err := findPasswordMetadataForLabel(db, parsed.Label)
	switch {
	case err != nil:
		return saltedpassword.SaltedPassword{}, err
	case ok:
		return checkRecord(loaded, parsed)
	default:
		return newRecord(parsed, db)
	}
}

func checkRecord(loaded passwordMetadata, parsed parser.PasswordEntry) (saltedpassword.SaltedPassword, error) {
	sp, err := saltedpassword.New(parsed.Label, parsed.Secret, loaded.Salt[:])
	if err != nil {
		return saltedpassword.SaltedPassword{}, err
	}

	if err := decryptCanary(sp.Encryptor, loaded.Canary, parsed.Label); err != nil {
		return saltedpassword.SaltedPassword{}, err
	}

	return sp, nil
}

func newRecord(parsed parser.PasswordEntry, db *gorm.DB) (saltedpassword.SaltedPassword, error) {
	salt, err := randomSalt()
	if err != nil {
		return saltedpassword.SaltedPassword{}, err
	}

	sp, err := saltedpassword.New(parsed.Label, parsed.Secret, salt)

	canary, err := encryptCanary(sp.Encryptor)
	if err != nil {
		return saltedpassword.SaltedPassword{}, err
	}

	err = savePasswordMetadata(db, passwordMetadata{
		Label:   parsed.Label,
		Salt:    salt,
		Canary:  canary,
		Primary: false,
	})
	if err != nil {
		return saltedpassword.SaltedPassword{}, err
	}

	return sp, nil
}
