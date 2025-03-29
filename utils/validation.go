package utils

import (
	"fmt"
	"strings"

	"github.com/robfig/cron/v3"
)

// ValidateCronExpression validates a cron expression and returns a normalized version
func ValidateCronExpression(cronExpr string) (string, error) {
	// Pre-processing to handle known issues
	// Replace ** with *
	cronExpr = strings.ReplaceAll(cronExpr, "**", "*")
	
	// Ensure proper spacing in cron expression
	parts := strings.Fields(cronExpr)
	if len(parts) < 5 {
		return "", fmt.Errorf("expected at least 5 fields, found %d: %v", len(parts), parts)
	}
	
	// Handle both 5-field and 6-field formats
	if len(parts) == 5 {
		// Standard 5-field format; add seconds field
		parts = append([]string{"0"}, parts...)
	} else if len(parts) != 6 {
		return "", fmt.Errorf("expected 5 or 6 fields, found %d: %v", len(parts), parts)
	}
	
	// Replace any ** (double asterisks) in each field with *
	for i, part := range parts {
		if part == "**" {
			parts[i] = "*"
		}
		parts[i] = strings.ReplaceAll(parts[i], "**", "*")
	}
	
	// Try to parse the cron expression
	normalized := strings.Join(parts, " ")
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	_, err := parser.Parse(normalized)
	if err != nil {
		return "", fmt.Errorf("invalid cron expression: %v", err)
	}
	
	return normalized, nil
}
