package models

import (
	"encoding/json"
	"fmt"
	"sync"
)

// CronJob defines a scheduled HTTP request
type CronJob struct {
	ID     string `json:"id"`
	Cron   string `json:"cron"`
	URL    string `json:"url"`
	Active bool   `json:"active"`
}

// UserConfig holds user-specific configuration and data
type UserConfig struct {
	CronJobs []CronJob              `json:"cron"`
	Data     map[string]interface{} `json:"data"`
	ApiKey   string                 `json:"-"` // Not stored in JSON
}

// AppConfig represents the full application configuration
type AppConfig struct {
	Users map[string]*UserConfig `json:"users"`
	mu    sync.RWMutex           // For thread-safe access
}

// NewAppConfig creates a new empty configuration
func NewAppConfig() *AppConfig {
	return &AppConfig{
		Users: make(map[string]*UserConfig),
	}
}

// GetUser retrieves a user's configuration, returns nil if not found
func (c *AppConfig) GetUser(username string) *UserConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Users[username]
}

// AddUser adds a new user to the configuration
func (c *AppConfig) AddUser(username, apiKey string) *UserConfig {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	user := &UserConfig{
		CronJobs: []CronJob{},
		Data:     make(map[string]interface{}),
		ApiKey:   apiKey,
	}
	c.Users[username] = user
	return user
}

// RemoveUser removes a user from the configuration
func (c *AppConfig) RemoveUser(username string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if _, exists := c.Users[username]; exists {
		delete(c.Users, username)
		return true
	}
	return false
}

// LoadFromJSON loads configuration from JSON data
func (c *AppConfig) LoadFromJSON(data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	return json.Unmarshal(data, &c.Users)
}

// SaveToJSON saves the configuration as JSON
func (c *AppConfig) SaveToJSON() ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	return json.MarshalIndent(c.Users, "", "  ")
}

// ValidateUserKey checks if the provided API key is valid for the given username
func (c *AppConfig) ValidateUserKey(username, apiKey string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	user, exists := c.Users[username]
	if !exists {
		return false
	}
	return user.ApiKey == apiKey
}

// AddCronJob adds a new cron job for a user
func (c *AppConfig) AddCronJob(username string, job CronJob) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	user, exists := c.Users[username]
	if !exists {
		return fmt.Errorf("user %s not found", username)
	}
	
	// Check for duplicate ID
	for _, existingJob := range user.CronJobs {
		if existingJob.ID == job.ID {
			return fmt.Errorf("cron job with ID %s already exists", job.ID)
		}
	}
	
	user.CronJobs = append(user.CronJobs, job)
	return nil
}

// UpdateCronJob updates an existing cron job
func (c *AppConfig) UpdateCronJob(username, jobID string, job CronJob) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	user, exists := c.Users[username]
	if !exists {
		return fmt.Errorf("user %s not found", username)
	}
	
	for i, existingJob := range user.CronJobs {
		if existingJob.ID == jobID {
			user.CronJobs[i] = job
			return nil
		}
	}
	
	return fmt.Errorf("cron job with ID %s not found", jobID)
}

// GetUserData retrieves a specific data value for a user
func (c *AppConfig) GetUserData(username, key string) (interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	user, exists := c.Users[username]
	if !exists {
		return nil, fmt.Errorf("user %s not found", username)
	}
	
	value, exists := user.Data[key]
	if !exists {
		return nil, fmt.Errorf("data key %s not found for user %s", key, username)
	}
	
	return value, nil
}

// SetUserData sets a data value for a user
func (c *AppConfig) SetUserData(username, key string, value interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	user, exists := c.Users[username]
	if !exists {
		return fmt.Errorf("user %s not found", username)
	}
	
	user.Data[key] = value
	return nil
}
