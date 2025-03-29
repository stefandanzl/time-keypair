package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
)

// CronJob represents a scheduled job configuration
type CronJob struct {
	ID     string `json:"id"`
	Cron   string `json:"cron"`
	URL    string `json:"url"`
	Active bool   `json:"active"`
}

// MarshalJSON implements custom JSON marshaling for CronJob
func (c *CronJob) MarshalJSON() ([]byte, error) {
	type Alias struct {
		ID     string `json:"id"`
		Cron   string `json:"cron"`
		URL    string `json:"url"`
		Active bool   `json:"active"`
	}
	
	// Create a clean copy with fixed cron expression
	cleanCron := strings.ReplaceAll(c.Cron, "**", "*")
	
	return json.Marshal(&Alias{
		ID:     c.ID,
		Cron:   cleanCron,
		URL:    c.URL,
		Active: c.Active,
	})
}

// UserData represents a user's configuration and data
type UserData struct {
	Cron []*CronJob              `json:"cron"`
	Data map[string]interface{} `json:"data"`
}

// Config represents the entire server configuration
type Config struct {
	mutex sync.RWMutex
	Users map[string]*UserData `json:"users"`
	Changed bool // Track if config has changed since last save
}

// NewConfig creates a new empty configuration
func NewConfig() *Config {
	return &Config{
		Users: make(map[string]*UserData),
		Changed: true, // New config should be saved
	}
}

// LoadConfig loads the configuration from the given file path
func LoadConfig(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config.Users); err != nil {
		return nil, err
	}

	// Initialize the Changed flag to false since we just loaded it
	config.Changed = false

	return &config, nil
}

// SaveConfig saves the configuration to the given file path
func SaveConfig(config *Config, filePath string) error {
	config.mutex.RLock()
	// Only proceed with saving if config has changed
	if !config.Changed {
		config.mutex.RUnlock()
		return nil
	}

	data, err := json.MarshalIndent(config.Users, "", "  ")
	config.mutex.RUnlock()

	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Create directory if it doesn't exist
	dir := extractDirectory(filePath)
	if dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write config file %s: %w", filePath, err)
	}

	// Reset the Changed flag after successful save
	config.mutex.Lock()
	config.Changed = false
	config.mutex.Unlock()

	return nil
}

// extractDirectory extracts the directory part from a file path
func extractDirectory(filePath string) string {
	for i := len(filePath) - 1; i >= 0; i-- {
		if filePath[i] == '/' || filePath[i] == '\\' {
			return filePath[:i]
		}
	}
	return ""
}

// GetUser returns the user data for the given user
func (c *Config) GetUser(user string) *UserData {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.Users[user]
}

// CreateUser creates a new user
func (c *Config) CreateUser(user string) *UserData {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, exists := c.Users[user]; !exists {
		c.Users[user] = &UserData{
			Cron: make([]*CronJob, 0),
			Data: make(map[string]interface{}),
		}
		c.Changed = true
	}

	return c.Users[user]
}

// DeleteUser deletes a user
func (c *Config) DeleteUser(user string) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, exists := c.Users[user]; exists {
		delete(c.Users, user)
		c.Changed = true
		return true
	}

	return false
}

// GetAllUsers returns all user IDs
func (c *Config) GetAllUsers() []string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	users := make([]string, 0, len(c.Users))
	for user := range c.Users {
		users = append(users, user)
	}

	return users
}

// GetUserKeys returns all data keys for a given user
func (c *Config) GetUserKeys(user string) []string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	userData, exists := c.Users[user]
	if !exists {
		return nil
	}

	keys := make([]string, 0, len(userData.Data))
	for key := range userData.Data {
		keys = append(keys, key)
	}

	return keys
}

// GetUserData retrieves data for a given user and key
func (c *Config) GetUserData(user, key string) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	userData, exists := c.Users[user]
	if !exists {
		return nil, false
	}

	value, exists := userData.Data[key]
	return value, exists
}

// SetUserData sets data for a given user and key
func (c *Config) SetUserData(user, key string, value interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	userData, exists := c.Users[user]
	if !exists {
		userData = &UserData{
			Cron: make([]*CronJob, 0),
			Data: make(map[string]interface{}),
		}
		c.Users[user] = userData
	}

	userData.Data[key] = value
	c.Changed = true
}

// DeleteUserData deletes data for a given user and key
func (c *Config) DeleteUserData(user, key string) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	userData, exists := c.Users[user]
	if !exists {
		return false
	}

	if _, exists := userData.Data[key]; exists {
		delete(userData.Data, key)
		c.Changed = true
		return true
	}

	return false
}

// GetUserJobs returns all cron jobs for a given user
func (c *Config) GetUserJobs(user string) []*CronJob {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	userData, exists := c.Users[user]
	if !exists {
		return nil
	}

	return userData.Cron
}

// AddUserJob adds a cron job for a given user
func (c *Config) AddUserJob(user string, job *CronJob) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	userData, exists := c.Users[user]
	if !exists {
		userData = &UserData{
			Cron: make([]*CronJob, 0),
			Data: make(map[string]interface{}),
		}
		c.Users[user] = userData
	}

	// Check if job with this ID already exists
	for i, existingJob := range userData.Cron {
		if existingJob.ID == job.ID {
			// Replace existing job
			userData.Cron[i] = job
			c.Changed = true
			return
		}
	}

	// Add new job
	userData.Cron = append(userData.Cron, job)
	c.Changed = true
}

// DeleteUserJob deletes a cron job for a given user
func (c *Config) DeleteUserJob(user, jobID string) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	userData, exists := c.Users[user]
	if !exists {
		return false
	}

	for i, job := range userData.Cron {
		if job.ID == jobID {
			// Remove job
			userData.Cron = append(userData.Cron[:i], userData.Cron[i+1:]...)
			c.Changed = true
			return true
		}
	}

	return false
}

// GetUserJob returns a specific cron job for a given user
func (c *Config) GetUserJob(user, jobID string) (*CronJob, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	userData, exists := c.Users[user]
	if !exists {
		return nil, false
	}

	for _, job := range userData.Cron {
		if job.ID == jobID {
			return job, true
		}
	}

	return nil, false
}

// SetUserJobActive sets the active state of a cron job
func (c *Config) SetUserJobActive(user, jobID string, active bool) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	userData, exists := c.Users[user]
	if !exists {
		return false
	}

	for _, job := range userData.Cron {
		if job.ID == jobID {
			if job.Active != active {
				job.Active = active
				c.Changed = true
			}
			return true
		}
	}

	return false
}

// SetAllUserJobsActive sets the active state of all cron jobs for a user
func (c *Config) SetAllUserJobsActive(user string, active bool) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	userData, exists := c.Users[user]
	if !exists {
		return false
	}

	changed := false
	for _, job := range userData.Cron {
		if job.Active != active {
			job.Active = active
			changed = true
		}
	}

	if changed {
		c.Changed = true
	}

	return true
}
