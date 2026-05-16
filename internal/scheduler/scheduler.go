package scheduler

import (
	"fmt"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// Task represents a function to be executed on a schedule
type Task func() error

// Scheduler handles the periodic execution of tasks
type Scheduler struct {
	cron   *cron.Cron
	logger *zap.Logger
}

// New creates a new Scheduler
func New(logger *zap.Logger) *Scheduler {
	return &Scheduler{
		cron:   cron.New(cron.WithSeconds()),
		logger: logger,
	}
}

// AddTask adds a task to the scheduler with a cron expression
func (s *Scheduler) AddTask(spec string, task Task) error {
	_, err := s.cron.AddFunc(spec, func() {
		s.logger.Info("Executing scheduled task")
		if err := task(); err != nil {
			s.logger.Error("Scheduled task failed", zap.Error(err))
		}
	})
	if err != nil {
		return fmt.Errorf("failed to add task: %w", err)
	}
	return nil
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
