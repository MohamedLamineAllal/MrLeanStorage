package engine

import (
	"os/exec"
	"strings"

	"github.com/mohamedlamineallal/MacosLeanStorage/internal/config"
	"github.com/mohamedlamineallal/MacosLeanStorage/internal/scheduler"
	"go.uber.org/zap"
)

type CommandHooks struct {
	BeforeHandleCommand    func(name string, command string, shouldExecuteCommand bool)
	AfterHandleCommand     func(name string, command string, err error)
	BeforeExecutingCommand func(name string, command string)
	AfterExecutingCommand  func(name string, command string, err error)
}

type CommandHandler struct {
	engine    *Engine
	scheduler *scheduler.Scheduler
	logger    *zap.Logger
}

func NewCommandHandler(engine *Engine, scheduler *scheduler.Scheduler, logger *zap.Logger) *CommandHandler {
	return &CommandHandler{
		engine:    engine,
		scheduler: scheduler,
		logger:    logger,
	}
}

func (ch *CommandHandler) Handle(t config.TargetConfig, hooks CommandHooks) {
	shouldRunCommand := ch.scheduler.ShouldRunCommand(t.Name, t.IntervalDays)

	if hooks.BeforeHandleCommand != nil {
		hooks.BeforeHandleCommand(t.Name, t.Command, shouldRunCommand)
	}

	err := error(nil)

	if shouldRunCommand {
		if hooks.BeforeExecutingCommand != nil {
			hooks.BeforeExecutingCommand(t.Name, t.Command)
		}

		err = ch.ExecuteCommand(t.Command)

		if err == nil && !ch.engine.Cleaner().DryRun() {
			ch.scheduler.UpdateCommandRunTime(t.Name)
		}

		if hooks.AfterExecutingCommand != nil {
			hooks.AfterExecutingCommand(t.Name, t.Command, err)
		}
	}

	if hooks.AfterHandleCommand != nil {
		hooks.AfterHandleCommand(t.Name, t.Command, err)
	}

}

// ExecuteCommand runs a shell command unless dry-run mode is enabled.
func (ch *CommandHandler) ExecuteCommand(command string) error {
	if ch.engine.cleaner.DryRun() {
		return nil
	}

	parts := strings.Fields(command)
	cmd := exec.Command(parts[0], parts[1:]...)
	err := cmd.Run()
	if err != nil {
		ch.logger.Error("Failed to execute command", zap.String("command", command), zap.Error(err))
		return err
	}
	return nil
}
