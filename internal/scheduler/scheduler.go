package scheduler

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// Task represents a function to be executed on a periodic schedule.
type Task func() error

// Scheduler handles the periodic execution of tasks and maintains state about task execution.
// It uses a cron-based scheduling system and tracks the last run time to handle missed tasks.
type Scheduler struct {
	cron      *cron.Cron
	logger    *zap.Logger
	statePath string
}

// New creates a new Scheduler instance and initializes the state path for tracking task execution.
func New(logger *zap.Logger) *Scheduler {
	s := &Scheduler{
		cron:   cron.New(cron.WithSeconds()),
		logger: logger,
	}
	s.statePath = filepath.Join(s.getBaseStatePath(), "mls-global.lastrun")
	return s
}

// getBaseStatePath returns the path where command run states are stored, ensuring the directory exists.
func (s *Scheduler) getBaseStatePath() string {
	cache, err := os.UserCacheDir()
	if err != nil {
		cache = os.TempDir()
	}
	path := filepath.Join(cache, "mls")
	_ = os.MkdirAll(path, 0755)
	return path
}

// ShouldRunCommand determines if a command should be executed based on its name and configured interval.
// It checks a state file to see when the command was last run.
func (s *Scheduler) ShouldRunCommand(commandName string, intervalDays int) bool {
	if intervalDays <= 0 {
		return true
	}

	statePath := filepath.Join(s.getBaseStatePath(), fmt.Sprintf("mls-cmd-%s.lastrun", commandName))
	data, err := os.ReadFile(statePath)

	if err != nil {
		return true
	}

	lastRun, err := time.Parse(time.RFC3339, string(data))
	if err != nil {
		return true
	}

	return time.Since(lastRun) >= time.Duration(intervalDays)*24*time.Hour
}

// UpdateCommandRunTime records the current time as the last run time for the specified command.
func (s *Scheduler) UpdateCommandRunTime(commandName string) {
	statePath := filepath.Join(s.getBaseStatePath(), fmt.Sprintf("mls-cmd-%s.lastrun", commandName))
	_ = os.WriteFile(statePath, []byte(time.Now().Format(time.RFC3339)), 0644)
}

// GetNextRunTime calculates when a command will next be eligible to run.
func (s *Scheduler) GetNextRunTime(commandName string, intervalDays int) (time.Time, error) {
	if intervalDays <= 0 {
		return time.Now(), nil
	}

	statePath := filepath.Join(s.getBaseStatePath(), fmt.Sprintf("mls-cmd-%s.lastrun", commandName))
	data, err := os.ReadFile(statePath)
	if err != nil {
		return time.Now(), nil
	}

	lastRun, err := time.Parse(time.RFC3339, string(data))
	if err != nil {
		return time.Now(), nil
	}

	return lastRun.Add(time.Duration(intervalDays) * 24 * time.Hour), nil
}

// AddTask schedules a task to be executed according to a cron-style specification string.
func (s *Scheduler) AddTask(spec string, task Task) error {
	_, err := s.cron.AddFunc(spec, func() {
		s.executeTask(task)
	})
	if err != nil {
		return fmt.Errorf("failed to add task: %w", err)
	}
	return nil
}

// executeTask runs the provided task, logs its execution, and updates the last run state on success.
func (s *Scheduler) executeTask(task Task) {
	s.logger.Info("Executing scheduled task")
	if err := task(); err != nil {
		s.logger.Error("Scheduled task failed", zap.Error(err))
	} else {
		// Update global last-run state only after a successful task execution
		_ = os.WriteFile(s.statePath, []byte(time.Now().Format(time.RFC3339)), 0644)
	}
}

// CheckForMissedTasks evaluates if a scheduled task was missed (e.g., computer was off) and runs it if necessary.
// It considers a task missed if the last run was more than 23 hours ago.
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

// Start begins the scheduler's execution loop.
func (s *Scheduler) Start() {
	s.cron.Start()
}

// Stop halts the scheduler's execution loop.
func (s *Scheduler) Stop() {
	s.cron.Stop()
}
