package problem

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/realmkit/rk-backend/pkg/api/headers"
	"github.com/realmkit/rk-backend/pkg/orm"
)

// ContentType is the problem response media type.
const ContentType = "application/problem+json"

// FieldError describes one invalid request field.
type FieldError struct {
	// Field is the invalid request field.
	Field string `json:"field"`

	// Message explains why the field is invalid.
	Message string `json:"message"`
}

// Problem describes a RealmKit HTTP error response.
type Problem struct {
	// Type is the stable problem type URI.
	Type string `json:"type"`

	// Title is a short human-readable error summary.
	Title string `json:"title"`

	// Status is the HTTP status code.
	Status int `json:"status"`

	// Detail is a human-readable error detail.
	Detail string `json:"detail,omitempty"`

	// Instance is the request URI that produced the problem.
	Instance string `json:"instance,omitempty"`

	// Code is the stable application error code.
	Code string `json:"code"`

	// RequestID is the request identifier.
	RequestID string `json:"request_id,omitempty"`

	// Errors contains field-level validation errors.
	Errors []FieldError `json:"errors,omitempty"`
}

// Error wraps a Problem as an error.
type Error struct {
	// Problem is the response payload.
	Problem Problem
}

// Error returns the problem detail.
func (err Error) Error() string {
	if err.Problem.Detail != "" {
		return err.Problem.Detail
	}
	return err.Problem.Title
}

// New creates a Problem payload.
func New(status int, code string, detail string) Problem {
	return Problem{
		Type:   fmt.Sprintf("https://realmkit.dev/problems/%s", code),
		Title:  titleFor(status),
		Status: status,
		Detail: detail,
		Code:   code,
	}
}

// Handler maps returned handler errors to problem responses.
func Handler(ctx *fiber.Ctx, err error) error {
	if problemErr, ok := err.(Error); ok {
		return Write(ctx, problemErr.Problem)
	}
	if payload, ok := FromContextError(err); ok {
		return Write(ctx, payload)
	}

	status := fiber.StatusInternalServerError
	detail := "Internal server error."
	if fiberErr, ok := err.(*fiber.Error); ok {
		status = fiberErr.Code
		detail = fiberErr.Message
	}

	return Write(ctx, New(status, codeFor(status), detail))
}

// FromContextError maps cancellation and deadline errors to problem responses.
func FromContextError(err error) (Problem, bool) {
	switch {
	case errors.Is(err, context.Canceled):
		return New(499, "request_cancelled", "Request was cancelled."), true
	case errors.Is(err, context.DeadlineExceeded):
		return New(fiber.StatusGatewayTimeout, "request_timeout", "Request deadline was exceeded."), true
	case errors.Is(err, orm.ErrUnavailable):
		return New(fiber.StatusServiceUnavailable, "dependency_unavailable", "A required dependency is unavailable."), true
	default:
		return Problem{}, false
	}
}

// Write writes a problem response to ctx.
func Write(ctx *fiber.Ctx, payload Problem) error {
	if payload.Instance == "" {
		payload.Instance = ctx.OriginalURL()
	}
	if payload.RequestID == "" {
		payload.RequestID = headers.CurrentRequestID(ctx)
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	ctx.Set(headers.ContentType, ContentType)
	return ctx.Status(payload.Status).Send(body)
}

// titleFor returns the default problem title for status.
func titleFor(status int) string {
	if statusText := http.StatusText(status); statusText != "" {
		return statusText
	}
	return "Error"
}

// codeFor returns the stable application error code for status.
func codeFor(status int) string {
	value := strings.ToLower(titleFor(status))
	value = strings.ReplaceAll(value, " ", "_")
	value = strings.ReplaceAll(value, "-", "_")
	return value
}
