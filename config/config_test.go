package config

import (
	"os"
	"testing"
)

func TestNewConfig(t *testing.T) {
	cfg := NewConfig()
	if cfg == nil {
		t.Error("NewConfig() returned nil")
	}
	if cfg.Users == nil {
		t.Error("NewConfig() did not initialize Users map")
	}
}

func TestCreateUser(t *testing.T) {
	cfg := NewConfig()
	user := "testuser"
	
	userData := cfg.CreateUser(user)
	if userData == nil {
		t.Error("CreateUser() returned nil")
	}
	
	if _, exists := cfg.Users[user]; !exists {
		t.Error("CreateUser() did not add user to config")
	}
	
	if userData.Cron == nil {
		t.Error("CreateUser() did not initialize Cron slice")
	}
	
	if userData.Data == nil {
		t.Error("CreateUser() did not initialize Data map")
	}
}

func TestGetUser(t *testing.T) {
	cfg := NewConfig()
	user := "testuser"
	
	// Before creating user
	if cfg.GetUser(user) != nil {
		t.Error("GetUser() did not return nil for non-existent user")
	}
	
	// Create user
	cfg.CreateUser(user)
	
	// After creating user
	if cfg.GetUser(user) == nil {
		t.Error("GetUser() returned nil for existing user")
	}
}

func TestDeleteUser(t *testing.T) {
	cfg := NewConfig()
	user := "testuser"
	
	// Before creating user
	if cfg.DeleteUser(user) {
		t.Error("DeleteUser() returned true for non-existent user")
	}
	
	// Create user
	cfg.CreateUser(user)
	
	// Delete user
	if !cfg.DeleteUser(user) {
		t.Error("DeleteUser() returned false for existing user")
	}
	
	// After deleting user
	if cfg.GetUser(user) != nil {
		t.Error("DeleteUser() did not remove user from config")
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	// Create a temporary file
	file, err := os.CreateTemp("", "config-test-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(file.Name())
	file.Close()
	
	// Create config
	cfg := NewConfig()
	user := "testuser"
	userData := cfg.CreateUser(user)
	
	// Add job
	job := &CronJob{
		ID:     "job1",
		Cron:   "* * * * *",
		URL:    "https://example.com",
		Active: true,
	}
	userData.Cron = append(userData.Cron, job)
	
	// Add data
	userData.Data["key1"] = "value1"
	userData.Data["key2"] = map[string]interface{}{
		"nested": "value",
	}
	
	// Save config
	if err := SaveConfig(cfg, file.Name()); err != nil {
		t.Fatalf("SaveConfig() failed: %v", err)
	}
	
	// Load config
	loadedCfg, err := LoadConfig(file.Name())
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}
	
	// Verify config
	if loadedCfg == nil {
		t.Fatal("LoadConfig() returned nil")
	}
	
	loadedUser := loadedCfg.GetUser(user)
	if loadedUser == nil {
		t.Fatal("LoadConfig() did not load user")
	}
	
	if len(loadedUser.Cron) != 1 {
		t.Fatal("LoadConfig() did not load jobs correctly")
	}
	
	loadedJob := loadedUser.Cron[0]
	if loadedJob.ID != job.ID || loadedJob.Cron != job.Cron || 
	   loadedJob.URL != job.URL || loadedJob.Active != job.Active {
		t.Error("LoadConfig() did not load job correctly")
	}
	
	if v, exists := loadedUser.Data["key1"]; !exists || v != "value1" {
		t.Error("LoadConfig() did not load simple data correctly")
	}
	
	if nested, exists := loadedUser.Data["key2"].(map[string]interface{}); !exists {
		t.Error("LoadConfig() did not load nested data correctly")
	} else if v, exists := nested["nested"]; !exists || v != "value" {
		t.Error("LoadConfig() did not load nested data value correctly")
	}
}
