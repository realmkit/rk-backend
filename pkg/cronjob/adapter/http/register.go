package http

import (
	"github.com/gofiber/fiber/v2"
	groupsport "github.com/realmkit/rk-backend/module/groups/port"
	"github.com/realmkit/rk-backend/pkg/cronjob/application"
)

// Services contains cron HTTP services.
type Services struct {
	// Cron manages job definitions and runs.
	Cron application.Service

	// Checker checks group-backed permissions.
	Checker groupsport.Checker
}

// Register registers cron job routes.
func Register(router fiber.Router, services Services) {
	handler := handler{services: services}
	group := router.Group("/cronjobs")
	group.Get("", handler.listDefinitions)
	group.Post("/locks/repair", handler.repairLocks)
	group.Get("/:job_key", handler.getDefinition)
	group.Get("/:job_key/runs", handler.listRuns)
	group.Post("/:job_key/run", handler.runNow)
	group.Post("/:job_key/pause", handler.pause)
	group.Post("/:job_key/resume", handler.resume)
}

// handler contains cron route dependencies.
type handler struct {
	services Services
}
