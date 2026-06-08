package main

import (
	"context"
	"errors"
	"os"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/gamehub-go/pkg/config"
	"github.com/niflaot/gamehub-go/pkg/logger"
	"go.uber.org/zap"
)

// main starts the GameHub backend process.
func main() {
	activeLogger, loggerErr := logger.New(logger.Config{Level: "info"})
	if loggerErr != nil {
		panic(loggerErr)
	}

	var runErr error
	defer func() {
		finish(activeLogger, runErr, os.Exit)
	}()

	runErr = runArgs(&activeLogger, os.Args[1:])
}

// run initializes configuration, logging, and the HTTP server.
func run(activeLogger **zap.Logger) error {
	return runArgs(activeLogger, nil)
}

// runArgs executes the GameHub CLI with args.
func runArgs(activeLogger **zap.Logger, args []string) error {
	return execute(activeLogger, args, defaultCommandDeps())
}

// execute executes the root command with dependencies.
func execute(activeLogger **zap.Logger, args []string, deps commandDeps) error {
	cmd := newRootCommand(activeLogger, deps)
	cmd.SetArgs(args)
	return cmd.Execute()
}

// runWith runs startup using injected dependencies for testable orchestration.
func runWith(
	activeLogger **zap.Logger,
	loadConfig func() (config.Config, error),
	newLogger func(logger.Config) (*zap.Logger, error),
	newServer func(*zap.Logger, bool) *fiber.App,
	listenServer func(*fiber.App, string) error,
) error {
	return runServe(context.Background(), activeLogger, commandDeps{
		loadConfig:    loadConfig,
		newLogger:     newLogger,
		newServer:     newServer,
		listenServer:  listenServer,
		openPostgres:  nil,
		closePostgres: nil,
		newRunner:     nil,
	})
}

// listen starts the Fiber application on the configured address.
func listen(app *fiber.App, address string) error {
	return app.Listen(address)
}

// finish logs final errors, syncs the logger, and exits when needed.
func finish(log *zap.Logger, err error, exit func(int)) {
	logged := false
	if err != nil {
		log.Error("gamehub backend failed", zap.Error(err))
		logged = true
	}

	if syncErr := syncLogger(log); syncErr != nil {
		err = errors.Join(err, syncErr)
		if !logged {
			log.Error("gamehub backend failed", zap.Error(err))
			_ = syncLogger(log)
		}
	}

	if err != nil {
		exit(1)
	}
}

// syncLogger flushes pending Zap logs while tolerating unsupported sync targets.
func syncLogger(log *zap.Logger) error {
	if log == nil {
		return nil
	}

	if err := log.Sync(); err != nil && !errors.Is(err, syscall.EINVAL) {
		return err
	}

	return nil
}
