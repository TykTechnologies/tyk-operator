package environment

import (
	"os"
	"strings"
)

// Read reads environment variable based on given key. If no environment variable is found,
// it returns given default value.
func Read(key, defaultValue string) string {
	v := strings.TrimSpace(os.Getenv(strings.TrimSpace(key)))
	if v == "" {
		v = defaultValue
	}

	return strings.TrimSpace(v)
}
