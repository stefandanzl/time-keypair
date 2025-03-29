package api

import (
	"data-cron-server/auth"
	"data-cron-server/config"
	"data-cron-server/cron"
	"data-cron-server/utils"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// handleAdminUsers handles the admin users endpoint
func (r *Router) handleAdminUsers(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		// List all users
		users := r.config.GetAllUsers()
		respondJSON(w, users)

	case http.MethodPost:
		// Create a new user
		var userData struct {
			User string `json:"user"`
		}

		if err := json.NewDecoder(req.Body).Decode(&userData); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if userData.User == "" {
			http.Error(w, "User ID is required", http.StatusBadRequest)
			return
		}

		// Create user
		r.config.CreateUser(userData.User)
		w.WriteHeader(http.StatusCreated)

	case http.MethodDelete:
		// Delete a user
		user := getPathPart(req.URL.Path, 3) // /admin/{super_key}/users/{user}
		if user == "" {
			http.Error(w, "User ID is required in path: /admin/{super_key}/users/{user}", http.StatusBadRequest)
			return
		}

		if r.config.DeleteUser(user) {
			w.WriteHeader(http.StatusNoContent)
		} else {
			http.Error(w, "User not found", http.StatusNotFound)
		}

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleAdminConfig handles the admin config endpoint
func (r *Router) handleAdminConfig(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		// Get full config
		respondJSON(w, r.config.Users)

	case http.MethodPut:
		// Replace full config
		body, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusBadRequest)
			return
		}

		var users map[string]*config.UserData
		if err := json.Unmarshal(body, &users); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Validate cron expressions in the configuration
		for _, userData := range users {
			for _, job := range userData.Cron {
				if job.Cron != "" {
					normalizedCron, err := utils.ValidateCronExpression(job.Cron)
					if err != nil {
						http.Error(w, fmt.Sprintf("Invalid cron expression for job %s: %v", job.ID, err), http.StatusBadRequest)
						return
					}
					job.Cron = normalizedCron
				}
			}
		}

		// Replace config
		r.config.Users = users

		// Reload scheduler
		r.scheduler.Stop()
		r.scheduler = cron.NewScheduler(r.config)

		w.WriteHeader(http.StatusOK)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleStatus handles the status endpoint
func (r *Router) handleStatus(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		// Get user from context
		user, ok := auth.UserFromContext(req.Context())
		if !ok {
			http.Error(w, "User not found in context", http.StatusInternalServerError)
			return
		}

		// Get job statuses
		statuses := r.scheduler.GetAllJobStatus(user)
		respondJSON(w, statuses)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleCronJobs handles the cron jobs endpoint
func (r *Router) handleCronJobs(w http.ResponseWriter, req *http.Request) {
	// Get user from context
	user, ok := auth.UserFromContext(req.Context())
	if !ok {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}

	switch req.Method {
	case http.MethodGet:
		// List all jobs
		jobs := r.config.GetUserJobs(user)
		respondJSON(w, jobs)

	case http.MethodPost:
		// Create a new job
		var job config.CronJob
		if err := json.NewDecoder(req.Body).Decode(&job); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate job
		if job.ID == "" {
			http.Error(w, "Job ID is required", http.StatusBadRequest)
			return
		}
		if job.Cron == "" {
			http.Error(w, "Cron expression is required", http.StatusBadRequest)
			return
		}
		if job.URL == "" {
			http.Error(w, "URL is required", http.StatusBadRequest)
			return
		}

		// Validate and normalize cron expression
		normalizedCron, err := utils.ValidateCronExpression(job.Cron)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid cron expression: %v", err), http.StatusBadRequest)
			return
		}
		job.Cron = normalizedCron

		// Add job to config
		r.config.AddUserJob(user, &job)

		// Add job to scheduler
		if err := r.scheduler.AddJob(user, &job); err != nil {
			http.Error(w, fmt.Sprintf("Invalid cron expression: %v", err), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusCreated)

	case http.MethodPut:
		// Update all jobs
		var jobs []*config.CronJob
		if err := json.NewDecoder(req.Body).Decode(&jobs); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate and normalize cron expressions
		for _, job := range jobs {
			if job.Cron != "" {
				normalizedCron, err := utils.ValidateCronExpression(job.Cron)
				if err != nil {
					http.Error(w, fmt.Sprintf("Invalid cron expression for job %s: %v", job.ID, err), http.StatusBadRequest)
					return
				}
				job.Cron = normalizedCron
			}
		}

		// Remove all existing jobs
		for _, job := range r.config.GetUserJobs(user) {
			r.scheduler.RemoveJob(user, job.ID)
		}

		// Add new jobs
		userData := r.config.GetUser(user)
		if userData == nil {
			userData = r.config.CreateUser(user)
		}
		userData.Cron = jobs

		// Add jobs to scheduler
		for _, job := range jobs {
			if job.Active {
				if err := r.scheduler.AddJob(user, job); err != nil {
					http.Error(w, fmt.Sprintf("Invalid cron expression for job %s: %v", job.ID, err), http.StatusBadRequest)
					return
				}
			}
		}

		w.WriteHeader(http.StatusOK)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleCronJobActivation handles activating or deactivating a cron job
func (r *Router) handleCronJobActivation(w http.ResponseWriter, req *http.Request, activate bool) {
	// Only allow GET method for activation/deactivation endpoints
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from context
	user, ok := auth.UserFromContext(req.Context())
	if !ok {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}

	// Extract job ID from the path
	// The path is /cron/{user_key}/{job_id}/on or /cron/{user_key}/{job_id}/off
	// We need to extract the job_id part
	parts := strings.Split(strings.Trim(req.URL.Path, "/"), "/")
	if len(parts) < 3 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}
	jobID := parts[1] // cron/job_id/[on|off]

	// Get the job
	job, exists := r.config.GetUserJob(user, jobID)
	if !exists {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	// Set active state if it's different
	if job.Active != activate {
		// Update in config
		r.config.SetUserJobActive(user, jobID, activate)

		// Update in scheduler
		if activate {
			// Add to scheduler if activating
			if err := r.scheduler.AddJob(user, job); err != nil {
				http.Error(w, fmt.Sprintf("Failed to activate job: %v", err), http.StatusInternalServerError)
				return
			}
		} else {
			// Remove from scheduler if deactivating
			r.scheduler.RemoveJob(user, jobID)
		}
	}

	// Return response based on action
	action := "activated"
	if !activate {
		action = "deactivated"
	}
	
	response := struct {
		ID     string `json:"id"`
		Action string `json:"action"`
		Active bool   `json:"active"`
	}{
		ID:     jobID,
		Action: action,
		Active: activate,
	}

	respondJSON(w, response)
}

// handleCronJob handles the cron job endpoint
func (r *Router) handleCronJob(w http.ResponseWriter, req *http.Request) {
	// Get user from context
	user, ok := auth.UserFromContext(req.Context())
	if !ok {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}

	// Get job ID from path
	jobID := getPathPart(req.URL.Path, 2) // /cron/{user_key}/{job_id}
	if jobID == "" {
		http.Error(w, "Job ID is required", http.StatusBadRequest)
		return
	}

	switch req.Method {
	case http.MethodGet:
		// Get job
		job, exists := r.config.GetUserJob(user, jobID)
		if !exists {
			http.Error(w, "Job not found", http.StatusNotFound)
			return
		}

		// Get job status
		status := r.scheduler.GetJobStatus(user, jobID)

		// Combine job and status
		response := struct {
			*config.CronJob
			Status *cron.JobStatus `json:"status,omitempty"`
		}{
			CronJob: job,
			Status:  status,
		}

		respondJSON(w, response)

	case http.MethodPut:
		// Update job
		var job config.CronJob
		if err := json.NewDecoder(req.Body).Decode(&job); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Ensure job ID matches
		job.ID = jobID

		// Validate and normalize cron expression
		if job.Cron != "" {
			normalizedCron, err := utils.ValidateCronExpression(job.Cron)
			if err != nil {
				http.Error(w, fmt.Sprintf("Invalid cron expression: %v", err), http.StatusBadRequest)
				return
			}
			job.Cron = normalizedCron
		}

		// Update job in config
		r.config.AddUserJob(user, &job)

		// Update job in scheduler
		if err := r.scheduler.UpdateJob(user, &job); err != nil {
			http.Error(w, fmt.Sprintf("Invalid cron expression: %v", err), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)

	case http.MethodDelete:
		// Delete job
		if !r.config.DeleteUserJob(user, jobID) {
			http.Error(w, "Job not found", http.StatusNotFound)
			return
		}

		// Remove job from scheduler
		r.scheduler.RemoveJob(user, jobID)

		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleDataKeys handles the data keys endpoint
func (r *Router) handleDataKeys(w http.ResponseWriter, req *http.Request) {
	// Get user from context
	user, ok := auth.UserFromContext(req.Context())
	if !ok {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}

	switch req.Method {
	case http.MethodGet:
		// List all data keys
		keys := r.config.GetUserKeys(user)
		respondJSON(w, keys)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleData handles the data endpoint
func (r *Router) handleData(w http.ResponseWriter, req *http.Request) {
	// Get user from context
	user, ok := auth.UserFromContext(req.Context())
	if !ok {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}

	// Get data key from path
	dataKey := getPathPart(req.URL.Path, 2) // /data/{user_key}/{data_key}
	if dataKey == "" {
		http.Error(w, "Data key is required", http.StatusBadRequest)
		return
	}

	switch req.Method {
	case http.MethodGet:
		// Get data
		data, exists := r.config.GetUserData(user, dataKey)
		if !exists {
			http.Error(w, "Data not found", http.StatusNotFound)
			return
		}

		respondJSON(w, data)

	case http.MethodPut:
		// Update data
		var data interface{}
		if err := json.NewDecoder(req.Body).Decode(&data); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Set data
		r.config.SetUserData(user, dataKey, data)

		w.WriteHeader(http.StatusOK)

	case http.MethodDelete:
		// Delete data
		if !r.config.DeleteUserData(user, dataKey) {
			http.Error(w, "Data not found", http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// respondJSON responds with JSON
func respondJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

// getPathPart gets a part from a URL path
func getPathPart(path string, index int) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if index < len(parts) {
		return parts[index]
	}
	return ""
}
