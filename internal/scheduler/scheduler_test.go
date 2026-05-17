package scheduler

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestScheduler(t *testing.T) {
	s := New(zap.NewNop())

	executed := make(chan bool, 1)
	task := func() error {
		executed <- true
		return nil
	}

	// Every second
	err := s.AddTask("* * * * * *", task)
	assert.NoError(t, err)

	s.Start()
	defer s.Stop()

	select {
	case <-executed:
		// Success
	case <-time.After(2 * time.Second):
		t.Fatal("Task not executed in time")
	}
}
