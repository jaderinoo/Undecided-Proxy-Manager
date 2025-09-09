package services

import (
	"fmt"
	"sync"
	"time"
)

// ScheduledJob represents a scheduled DNS update job
type ScheduledJob struct {
	RecordID    int
	Interval    time.Duration
	StopChan    chan bool
	LastStarted time.Time
	NextUpdate  time.Time
	IsPaused    bool
	PausedAt    *time.Time
}

// SchedulerService manages scheduled DNS update jobs
type SchedulerService struct {
	dnsService *DNSService
	jobs       map[int]*ScheduledJob // RecordID -> Job
	mutex      sync.RWMutex
}

// NewSchedulerService creates a new scheduler service
func NewSchedulerService(dnsService *DNSService) *SchedulerService {
	return &SchedulerService{
		dnsService: dnsService,
		jobs:       make(map[int]*ScheduledJob),
	}
}

// StartScheduledJob starts a scheduled job for a DNS record
func (s *SchedulerService) StartScheduledJob(recordID int, refreshRateMinutes int) error {
	if refreshRateMinutes <= 0 {
		return fmt.Errorf("refresh rate must be positive")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Stop existing job if it exists
	if job, exists := s.jobs[recordID]; exists {
		job.StopChan <- true
		delete(s.jobs, recordID)
	}

	// Create new job
	interval := time.Duration(refreshRateMinutes) * time.Minute
	now := time.Now()
	job := &ScheduledJob{
		RecordID:    recordID,
		Interval:    interval,
		StopChan:    make(chan bool, 1),
		LastStarted: now,
		NextUpdate:  now.Add(interval),
	}

	s.jobs[recordID] = job

	// Start the job in a goroutine
	go s.runScheduledJob(job)

	fmt.Printf("Started scheduled DNS update job for record %d with interval %v\n", recordID, interval)
	return nil
}

// StopScheduledJob stops a scheduled job for a DNS record
func (s *SchedulerService) StopScheduledJob(recordID int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if job, exists := s.jobs[recordID]; exists {
		job.StopChan <- true
		delete(s.jobs, recordID)
		fmt.Printf("Stopped scheduled DNS update job for record %d\n", recordID)
	}
}

// PauseScheduledJob pauses a scheduled job for a DNS record
func (s *SchedulerService) PauseScheduledJob(recordID int) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if job, exists := s.jobs[recordID]; exists {
		if !job.IsPaused {
			job.IsPaused = true
			now := time.Now()
			job.PausedAt = &now
			fmt.Printf("Paused scheduled DNS update job for record %d\n", recordID)
		}
		return nil
	}
	return fmt.Errorf("job not found for record %d", recordID)
}

// ResumeScheduledJob resumes a paused scheduled job for a DNS record
func (s *SchedulerService) ResumeScheduledJob(recordID int) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if job, exists := s.jobs[recordID]; exists {
		if job.IsPaused {
			job.IsPaused = false
			// Adjust next update time based on how long it was paused
			if job.PausedAt != nil {
				pausedDuration := time.Since(*job.PausedAt)
				job.NextUpdate = job.NextUpdate.Add(pausedDuration)
			}
			job.PausedAt = nil
			fmt.Printf("Resumed scheduled DNS update job for record %d\n", recordID)
		}
		return nil
	}
	return fmt.Errorf("job not found for record %d", recordID)
}

// runScheduledJob runs the actual scheduled job
func (s *SchedulerService) runScheduledJob(job *ScheduledJob) {
	ticker := time.NewTicker(job.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Check if job is paused
			if job.IsPaused {
				fmt.Printf("Scheduled job for record %d is paused, skipping update\n", job.RecordID)
				continue
			}

			// Update the next update time immediately when timer fires
			job.NextUpdate = time.Now().Add(job.Interval)

			// Perform the DNS update
			fmt.Printf("Running scheduled DNS update for record %d\n", job.RecordID)
			response, err := s.dnsService.UpdateDNSRecord(job.RecordID)
			if err != nil {
				fmt.Printf("Scheduled DNS update failed for record %d: %v\n", job.RecordID, err)
			} else if response != nil {
				if response.Success {
					fmt.Printf("Scheduled DNS update successful for record %d: %s (IP: %s)\n",
						job.RecordID, response.Message, response.NewIP)
				} else {
					fmt.Printf("Scheduled DNS update failed for record %d: %s\n",
						job.RecordID, response.Message)
				}
			}
		case <-job.StopChan:
			fmt.Printf("Scheduled job for record %d stopped\n", job.RecordID)
			return
		}
	}
}

// UpdateScheduledJob updates or creates a scheduled job for a DNS record
func (s *SchedulerService) UpdateScheduledJob(recordID int, refreshRateMinutes *int) error {
	// If refresh rate is nil or 0, stop the job
	if refreshRateMinutes == nil || *refreshRateMinutes <= 0 {
		s.StopScheduledJob(recordID)
		return nil
	}

	// Start or update the job
	return s.StartScheduledJob(recordID, *refreshRateMinutes)
}

// JobInfo contains information about a scheduled job
type JobInfo struct {
	Interval    time.Duration `json:"interval"`
	LastStarted time.Time     `json:"last_started"`
	NextUpdate  time.Time     `json:"next_update"`
	IsPaused    bool          `json:"is_paused"`
	PausedAt    *time.Time    `json:"paused_at,omitempty"`
}

// GetActiveJobs returns information about active scheduled jobs
func (s *SchedulerService) GetActiveJobs() map[int]JobInfo {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	activeJobs := make(map[int]JobInfo)
	for recordID, job := range s.jobs {
		activeJobs[recordID] = JobInfo{
			Interval:    job.Interval,
			LastStarted: job.LastStarted,
			NextUpdate:  job.NextUpdate,
			IsPaused:    job.IsPaused,
			PausedAt:    job.PausedAt,
		}
	}
	return activeJobs
}

// StopAllJobs stops all scheduled jobs
func (s *SchedulerService) StopAllJobs() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for recordID, job := range s.jobs {
		job.StopChan <- true
		fmt.Printf("Stopped scheduled DNS update job for record %d\n", recordID)
	}
	s.jobs = make(map[int]*ScheduledJob)
}

// LoadAndStartJobs loads all DNS records with refresh rates and starts their scheduled jobs
func (s *SchedulerService) LoadAndStartJobs() error {
	// Get all DNS configs
	configs, err := s.dnsService.DbService.GetDNSConfigs()
	if err != nil {
		return fmt.Errorf("failed to get DNS configs: %w", err)
	}

	for _, config := range configs {
		if !config.IsActive {
			continue
		}

		// Get all records for this config
		records, err := s.dnsService.DbService.GetDNSRecords(config.ID)
		if err != nil {
			fmt.Printf("Warning: Failed to get records for config %d: %v\n", config.ID, err)
			continue
		}

		// Start jobs for records with refresh rates
		for _, record := range records {
			if record.IsActive && record.DynamicDNSRefreshRate != nil && *record.DynamicDNSRefreshRate > 0 {
				err := s.StartScheduledJob(record.ID, *record.DynamicDNSRefreshRate)
				if err != nil {
					fmt.Printf("Warning: Failed to start scheduled job for record %d: %v\n", record.ID, err)
				}
			}
		}
	}

	return nil
}
