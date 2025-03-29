package cron

import (
	"data-cron-server/config"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

// JobStatus tracks the execution status of a cron job
type JobStatus struct {
	LastRun     time.Time `json:"last_run"`
	LastSuccess bool      `json:"last_success"`
	LastError   string    `json:"last_error,omitempty"`
	NextRun     time.Time `json:"next_run,omitempty"`
}

// Scheduler manages cron jobs
type Scheduler struct {
	cron       *cron.Cron
	config     *config.Config
	entryIDs   map[string]map[string]cron.EntryID // user -> jobID -> entryID
	jobStatus  map[string]map[string]*JobStatus   // user -> jobID -> status
	httpClient *http.Client
	mutex      sync.RWMutex
}

// NewScheduler creates a new scheduler
func NewScheduler(cfg *config.Config) *Scheduler {
	scheduler := &Scheduler{
		cron:       cron.New(cron.WithSeconds()),
		config:     cfg,
		entryIDs:   make(map[string]map[string]cron.EntryID),
		jobStatus:  make(map[string]map[string]*JobStatus),
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	// Load existing jobs from config
	scheduler.loadAllJobs()

	// Start the cron scheduler
	scheduler.cron.Start()

	return scheduler
}

// loadAllJobs loads all jobs from the configuration
func (s *Scheduler) loadAllJobs() {
	for user, userData := range s.config.Users {
		for _, job := range userData.Cron {
			if job.Active {
				s.AddJob(user, job)
			}
		}
	}
}

// AddJob adds a job to the scheduler
func (s *Scheduler) AddJob(user string, job *config.CronJob) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Create user maps if they don't exist
	if _, exists := s.entryIDs[user]; !exists {
		s.entryIDs[user] = make(map[string]cron.EntryID)
	}
	if _, exists := s.jobStatus[user]; !exists {
		s.jobStatus[user] = make(map[string]*JobStatus)
	}

	// Remove existing job if it exists
	if entryID, exists := s.entryIDs[user][job.ID]; exists {
		s.cron.Remove(entryID)
	}

	// Only add the job if it's active
	if job.Active {
		// Create job function
		jobFunc := func() {
			s.executeJob(user, job)
		}

		// Add job to cron
		entryID, err := s.cron.AddFunc(job.Cron, jobFunc)
		if err != nil {
			return err
		}

		// Store entry ID
		s.entryIDs[user][job.ID] = entryID

		// Initialize job status
		s.jobStatus[user][job.ID] = &JobStatus{
			NextRun: s.getNextRunTime(entryID),
		}
	}

	return nil
}

// RemoveJob removes a job from the scheduler
func (s *Scheduler) RemoveJob(user, jobID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if userEntries, exists := s.entryIDs[user]; exists {
		if entryID, exists := userEntries[jobID]; exists {
			s.cron.Remove(entryID)
			delete(userEntries, jobID)
		}
	}
}

// executeJob executes a job by making an HTTP GET request
func (s *Scheduler) executeJob(user string, job *config.CronJob) {
	s.mutex.Lock()
	status := s.jobStatus[user][job.ID]
	status.LastRun = time.Now()
	s.mutex.Unlock()

	// Make HTTP request
	resp, err := s.httpClient.Get(job.URL)
	
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if err != nil {
		status.LastSuccess = false
		status.LastError = err.Error()
		log.Printf("Job %s for user %s failed: %v", job.ID, user, err)
	} else {
		resp.Body.Close()
		status.LastSuccess = resp.StatusCode >= 200 && resp.StatusCode < 300
		if !status.LastSuccess {
			status.LastError = "HTTP Status: " + resp.Status
			log.Printf("Job %s for user %s returned non-success status: %s", job.ID, user, resp.Status)
		} else {
			status.LastError = ""
			log.Printf("Job %s for user %s completed successfully", job.ID, user)
		}
	}

	// Update next run time
	if entryID, exists := s.entryIDs[user][job.ID]; exists {
		status.NextRun = s.getNextRunTime(entryID)
	}
}

// getNextRunTime gets the next run time for a cron entry
func (s *Scheduler) getNextRunTime(entryID cron.EntryID) time.Time {
	entry := s.cron.Entry(entryID)
	return entry.Next
}

// GetJobStatus gets the status of a job
func (s *Scheduler) GetJobStatus(user, jobID string) *JobStatus {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if userStatus, exists := s.jobStatus[user]; exists {
		return userStatus[jobID]
	}

	return nil
}

// GetAllJobStatus gets the status of all jobs for a user
func (s *Scheduler) GetAllJobStatus(user string) map[string]*JobStatus {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if userStatus, exists := s.jobStatus[user]; exists {
		// Create a copy to avoid concurrent access issues
		statusCopy := make(map[string]*JobStatus)
		for jobID, status := range userStatus {
			statusCopy[jobID] = status
		}
		return statusCopy
	}

	return nil
}

// UpdateJob updates a job in the scheduler
func (s *Scheduler) UpdateJob(user string, job *config.CronJob) error {
	// First remove the job if it exists
	s.RemoveJob(user, job.ID)

	// Then add it back if it's active
	if job.Active {
		return s.AddJob(user, job)
	}

	return nil
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	s.cron.Stop()
}
