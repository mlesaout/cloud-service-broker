package passwords

import "github.com/cloudfoundry-incubator/cloud-service-broker/internal/passwords/saltedpassword"

type Passwords struct {
	Primary        saltedpassword.SaltedPassword
	Secondaries    []saltedpassword.SaltedPassword
	ChangedPrimary bool
}

// Secrets returns a slice with the primary secret first, followed by secondary secrets
func (p Passwords) Secrets() []string {
	result := []string{p.Primary.Secret}
	for _, s := range p.Secondaries {
		result = append(result, s.Secret)
	}
	return result
}
