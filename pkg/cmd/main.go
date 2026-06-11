package main

import (
	"errors"
	"fmt"
	"os"
	"syscall"

	"github.com/realmkit/rk-backend/pkg/cli"
	"github.com/realmkit/rk-backend/pkg/logger"
	"go.uber.org/zap"
)

// main starts the RealmKit backend process.
func main() {
	activeLogger, loggerErr := logger.New(logger.Config{Level: "info"})
	if loggerErr != nil {
		fmt.Fprintf(os.Stderr, "create startup logger: %v\n", loggerErr)
		os.Exit(1)
	}

	var runErr error
	defer func() {
		finish(activeLogger, runErr, os.Exit)
	}()

	runErr = cli.Run(os.Args[1:], &activeLogger)
}

// finish logs final errors, syncs the logger, and exits when needed.
func finish(log *zap.Logger, err error, exit func(int)) {
	logged := false
	if err != nil {
		log.Error("realmkit backend failed", zap.Error(err))
		logged = true
	}

	if syncErr := syncLogger(log); syncErr != nil {
		err = errors.Join(err, syncErr)
		if !logged {
			log.Error("realmkit backend failed", zap.Error(err))
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

	if err := log.Sync(); err != nil && !unsupportedSyncError(err) {
		return err
	}

	return nil
}

// unsupportedSyncError reports whether Zap cannot sync the current output target.
func unsupportedSyncError(err error) bool {
	return errors.Is(err, syscall.EINVAL) || errors.Is(err, syscall.EBADF) || errors.Is(err, syscall.ENOTTY)
}
