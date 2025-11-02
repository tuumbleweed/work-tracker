package email

import (
	"fmt"
	"slices"
	"time"

	er "work-tracker/src/pkg/error"
)

const timeout time.Duration = 30*time.Second

// Provider represents supported email providers.
type Provider string

const (
	ProviderMailgun   Provider = "mailgun"
	ProviderSendGrid  Provider = "sendgrid"
	ProviderAmazonSES Provider = "amazonses"
)

var AllowedProviders = []Provider{ProviderMailgun, ProviderSendGrid, ProviderAmazonSES}

// IsValidProvider checks if the given string matches a known provider.
// Returns error if not valid
func IsValidProvider(provider Provider) (e *er.Error) {
	if slices.Contains(AllowedProviders, provider) {
		return nil
	}

	allowed := make([]string, len(AllowedProviders))
	for i, prov := range AllowedProviders {
		allowed[i] = string(prov)
	}

	return er.NewError(
		fmt.Errorf("Unsupported provider: '%s'", provider),
		fmt.Sprintf("Provider must be among those: %v", AllowedProviders),
		provider,
	)
}
