package validate

import (
	"fmt"
	"strings"
)

// NonEmpty returns an error if the value is empty or whitespace-only.
func NonEmpty(name, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s cannot be empty", name)
	}
	return nil
}

// Trimmed trims whitespace from the value and validates it is non-empty.
func Trimmed(name string, value *string) error {
	*value = strings.TrimSpace(*value)
	if *value == "" {
		return fmt.Errorf("%s cannot be empty", name)
	}
	return nil
}

// knownPrefixes maps RevenueCat resource types to their expected ID prefixes.
var knownPrefixes = map[string]string{
	"project":      "proj",
	"app":          "app",
	"product":      "prod",
	"entitlement":  "entl",
	"offering":     "ofrnge",
	"package":      "pkge",
	"subscription": "sub",
	"purchase":     "purch",
}

// ResourceID validates that a resource ID is non-empty and optionally checks
// for the expected prefix. Returns the trimmed value.
func ResourceID(resourceType, value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("%s ID cannot be empty", resourceType)
	}
	if prefix, ok := knownPrefixes[resourceType]; ok {
		if !strings.HasPrefix(value, prefix) {
			return value, fmt.Errorf("warning: %s ID %q does not start with expected prefix %q", resourceType, value, prefix)
		}
	}
	return value, nil
}
