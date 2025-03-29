package utils

import (
	"fmt"
	"strings"

	"github.com/robfig/cron/v3"
)

// ValidateCronExpression validates a cron expression and returns a normalized version
func ValidateCronExpression(cronExpr string) (string, error) {
	// Try to parse the cron expression
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	_, err := parser.Parse(cronExpr)
	if err != nil {
		return "", fmt.Errorf("invalid cron expression: %v", err)
	}

	// Normalize the cron expression format
	parts := strings.Fields(cronExpr)
	if len(parts) != 6 {
		return "", fmt.Errorf("expected exactly 6 fields, found %d: %v", len(parts), parts)
	}

	// Make sure there are no double asterisks
	for i, part := range parts {
		if part == "**" {
			parts[i] = "*"
		}
	}

	return strings.Join(parts, " "), nil
}
