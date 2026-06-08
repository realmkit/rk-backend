package problem

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/gamehub-go/pkg/api/headers"
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

// Problem describes a GameHub HTTP error response.
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
		Type:   fmt.Sprintf("https://gamehub.dev/problems/%s", code),
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

	status := fiber.StatusInternalServerError
	detail := "Internal server error."
	if fiberErr, ok := err.(*fiber.Error); ok {
		status = fiberErr.Code
		detail = fiberErr.Message
	}

	return Write(ctx, New(status, codeFor(status), detail))
}

// Write writes a problem response to ctx.
func Write(ctx *fiber.Ctx, payload Problem) error {
	if payload.Instance == "" {
		payload.Instance = ctx.OriginalURL()
	}
	if payload.RequestID == "" {
		payload.RequestID = headers.CurrentRequestID(ctx)
	}
	ctx.Set(headers.ContentType, ContentType)
	return ctx.Status(payload.Status).JSON(payload)
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
