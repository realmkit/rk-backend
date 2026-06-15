package http

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/groups/adapter/httpguard"
	groupsdomain "github.com/realmkit/rk-backend/module/groups/domain"
	groupsport "github.com/realmkit/rk-backend/module/groups/port"
	"github.com/realmkit/rk-backend/module/metadata/domain"
	"github.com/realmkit/rk-backend/module/metadata/port"
	"github.com/realmkit/rk-backend/pkg/api/headers"
	"github.com/realmkit/rk-backend/pkg/api/problem"
	"github.com/realmkit/rk-backend/pkg/pagination"
)

// decodeJSON decodes the request body into target.
func decodeJSON(ctx *fiber.Ctx, target any) error {
	decoder := json.NewDecoder(strings.NewReader(string(ctx.Body())))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return problem.Error{Problem: problem.New(fiber.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")}
	}
	return nil
}

// writeJSON writes a JSON response.
func writeJSON(ctx *fiber.Ctx, status int, payload any) error {
	ctx.Set(headers.ContentType, "application/json")
	return ctx.Status(status).JSON(payload)
}

// writeNoContent writes a no content response.
func writeNoContent(ctx *fiber.Ctx) error {
	return ctx.SendStatus(fiber.StatusNoContent)
}

// handleError maps application errors to problem responses.
func handleError(ctx *fiber.Ctx, err error) error {
	var validation domain.ValidationError
	if errors.As(err, &validation) {
		payload := problem.New(fiber.StatusUnprocessableEntity, "validation_failed", "Request validation failed.")
		payload.Errors = make([]problem.FieldError, 0, len(validation.Violations))
		for _, violation := range validation.Violations {
			payload.Errors = append(payload.Errors, problem.FieldError{Field: violation.Field, Message: violation.Message})
		}
		return problem.Write(ctx, payload)
	}
	switch {
	case errors.Is(err, port.ErrNotFound):
		return problem.Write(ctx, problem.New(fiber.StatusNotFound, "metadata_not_found", "Metadata resource was not found."))
	case errors.Is(err, port.ErrPreconditionFailed):
		return problem.Write(
			ctx,
			problem.New(fiber.StatusPreconditionFailed, "metadata_precondition_failed", "Resource version did not match."),
		)
	case errors.Is(err, port.ErrConflict):
		return problem.Write(
			ctx,
			problem.New(fiber.StatusConflict, "metadata_conflict", "Metadata resource conflicts with existing state."),
		)
	case errors.Is(err, port.ErrReferenced):
		return problem.Write(ctx, problem.New(fiber.StatusConflict, "metadata_referenced", "Metadata resource has active dependents."))
	case errors.Is(err, port.ErrInactive):
		return problem.Write(ctx, problem.New(fiber.StatusConflict, "metadata_inactive", "Metadata resource is inactive."))
	default:
		return err
	}
}

// idFromParam parses a UUID route parameter.
func idFromParam(ctx *fiber.Ctx, name string) (uuid.UUID, error) {
	id, err := uuid.Parse(ctx.Params(name))
	if err != nil {
		return uuid.Nil, problem.Error{Problem: problem.New(fiber.StatusBadRequest, "invalid_path_parameter", name+" must be a UUID.")}
	}
	return id, nil
}

// ownerFromParams parses owner route parameters.
func ownerFromParams(ctx *fiber.Ctx) (port.OwnerRef, error) {
	ownerID, err := idFromParam(ctx, "owner_id")
	if err != nil {
		return port.OwnerRef{}, err
	}
	return port.OwnerRef{Type: domain.OwnerType(ctx.Params("owner_type")), ID: ownerID}, nil
}

// pageFromQuery parses pagination query parameters.
func pageFromQuery(ctx *fiber.Ctx) (pagination.Page, error) {
	page, err := pagination.New(pagination.Request{
		Limit:  ctx.QueryInt("page_size"),
		Cursor: ctx.Query("page_token"),
	})
	if err != nil {
		return pagination.Page{}, problem.Error{
			Problem: problem.New(fiber.StatusBadRequest, "invalid_pagination", "Pagination parameters are invalid."),
		}
	}
	return page, nil
}

// expectedVersion returns the required If-Match version.
func expectedVersion(ctx *fiber.Ctx) (uint64, error) {
	value := strings.Trim(ctx.Get(headers.IfMatch), `" `)
	if value == "" {
		return 0, problem.Error{Problem: problem.New(fiber.StatusPreconditionRequired, "if_match_required", "If-Match header is required.")}
	}
	version, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, problem.Error{
			Problem: problem.New(fiber.StatusBadRequest, "invalid_if_match", "If-Match must contain a numeric version."),
		}
	}
	return version, nil
}

// optionalExpectedVersion returns the optional If-Match version.
func optionalExpectedVersion(ctx *fiber.Ctx) (*uint64, error) {
	if strings.TrimSpace(ctx.Get(headers.IfMatch)) == "" {
		return nil, nil
	}
	version, err := expectedVersion(ctx)
	if err != nil {
		return nil, err
	}
	return &version, nil
}

// requireIdempotency verifies Idempotency-Key is present.
func requireIdempotency(ctx *fiber.Ctx) error {
	if strings.TrimSpace(ctx.Get(headers.IdempotencyKey)) == "" {
		return problem.Error{
			Problem: problem.New(fiber.StatusBadRequest, "idempotency_key_required", "Idempotency-Key header is required."),
		}
	}
	return nil
}

// setETag writes a weak version ETag.
func setETag(ctx *fiber.Ctx, version uint64) {
	ctx.Set(headers.ETag, `"`+strconv.FormatUint(version, 10)+`"`)
}

// activeFromPointer returns active defaulting to true.
func activeFromPointer(active *bool) bool {
	if active == nil {
		return true
	}
	return *active
}

// requireMetadata verifies one metadata permission target.
func requireMetadata(ctx *fiber.Ctx, checker groupsport.Checker, target httpguard.Target) error {
	_, err := httpguard.Require(ctx, checker, target)
	return err
}

// definitionReadTarget returns the metadata definition read target.
func definitionReadTarget() httpguard.Target {
	return httpguard.All(groupsdomain.PermissionMetadataReadDefinitions, groupsdomain.ObjectMetadata)
}

// definitionManageTarget returns the metadata definition management target.
func definitionManageTarget() httpguard.Target {
	return httpguard.All(groupsdomain.PermissionMetadataManageDefinitions, groupsdomain.ObjectMetadata)
}

// entryReadTarget returns the metaobject entry read target.
func entryReadTarget() httpguard.Target {
	return httpguard.All(groupsdomain.PermissionMetadataReadEntries, groupsdomain.ObjectMetadata)
}

// entryWriteTarget returns the metaobject entry write target.
func entryWriteTarget() httpguard.Target {
	return httpguard.All(groupsdomain.PermissionMetadataWriteEntries, groupsdomain.ObjectMetadata)
}

// ownerReadTarget returns a metadata owner read target.
func ownerReadTarget(owner port.OwnerRef) httpguard.Target {
	if owner.Type == domain.OwnerUser {
		return httpguard.Object(groupsdomain.PermissionMetadataReadUser, groupsdomain.ObjectUser, owner.ID)
	}
	return httpguard.All(groupsdomain.PermissionMetadataReadValues, groupsdomain.ObjectMetadata)
}

// ownerWriteTarget returns a metadata owner write target.
func ownerWriteTarget(owner port.OwnerRef) httpguard.Target {
	if owner.Type == domain.OwnerUser {
		return httpguard.Object(groupsdomain.PermissionMetadataWriteUser, groupsdomain.ObjectUser, owner.ID)
	}
	return httpguard.All(groupsdomain.PermissionMetadataWriteValues, groupsdomain.ObjectMetadata)
}
