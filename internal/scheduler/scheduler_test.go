package scheduler

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mohamedlamineallal/MacosLeanStorage/internal/utils"
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

func TestShouldRunCommand(t *testing.T) {
	s := New(zap.NewNop())
	commandName := "test-cmd"
	statePath := filepath.Join(utils.GetAppCacheDir(), fmt.Sprintf("mls-cmd-%s.lastrun", commandName))
	os.Remove(statePath) // Ensure clean state
	defer os.Remove(statePath)
	
	// Test first run (should run)
	assert.True(t, s.ShouldRunCommand(commandName, 30))
	
	// Record run
	s.UpdateCommandRunTime(commandName)
	
	// Test immediate check (should not run)
	assert.False(t, s.ShouldRunCommand(commandName, 30))
}
