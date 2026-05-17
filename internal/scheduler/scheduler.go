package scheduler

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// Task represents a function to be executed on a schedule
type Task func() error

// Scheduler handles the periodic execution of tasks
type Scheduler struct {
	cron       *cron.Cron
	logger     *zap.Logger
	statePath  string
}

// New creates a new Scheduler
func New(logger *zap.Logger) *Scheduler {
	home, _ := os.UserHomeDir()
	return &Scheduler{
		cron:      cron.New(cron.WithSeconds()),
		logger:    logger,
		statePath: filepath.Join(home, ".MacosLeanStorage.lastrun"),
	}
}

// AddTask adds a task to the scheduler with a cron expression
func (s *Scheduler) AddTask(spec string, task Task) error {
	_, err := s.cron.AddFunc(spec, func() {
		s.executeTask(task)
	})
	if err != nil {
		return fmt.Errorf("failed to add task: %w", err)
	}
	return nil
}

func (s *Scheduler) executeTask(task Task) {
	s.logger.Info("Executing scheduled task")
	if err := task(); err != nil {
		s.logger.Error("Scheduled task failed", zap.Error(err))
	} else {
		// Update last run time on success
		_ = os.WriteFile(s.statePath, []byte(time.Now().Format(time.RFC3339)), 0644)
	}
}

// CheckForMissedTasks checks if the last run was more than 24 hours ago and triggers a run
func (s *Scheduler) CheckForMissedTasks(task Task) {
	data, err := os.ReadFile(s.statePath)
	if err != nil {
		return // No previous run state
	}

	lastRun, err := time.Parse(time.RFC3339, string(data))
	if err != nil {
		return
	}

	// If it has been more than 23 hours, consider it missed (allowing for drift)
	if time.Since(lastRun) > 23*time.Hour {
		s.logger.Info("Missed scheduled task detected, catching up...", zap.Time("last_run", lastRun))
		s.executeTask(task)
	}
}

// Start begins the scheduler execution
func (s *Scheduler) Start() {
	s.logger.Info("Starting scheduler")
	s.cron.Start()
}

// Stop halts the scheduler execution
func (s *Scheduler) Stop() {
	s.logger.Info("Stopping scheduler")
	s.cron.Stop()
}
