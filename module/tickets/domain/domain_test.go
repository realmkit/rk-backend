package domain

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestDefinitionValidation verifies workflow definition invariants.
func TestDefinitionValidation(t *testing.T) {
	definition := Definition{
		ID:                  uuid.New(),
		Key:                 "appeal",
		Name:                "Appeal",
		Kind:                KindAppeal,
		Status:              DefinitionActive,
		RequiresPunishment:  false,
		MaxOpenPerSubmitter: -1,
	}
	err := definition.Validate()
	var validation ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("Validate() error = %v, want ValidationError", err)
	}
	if len(validation.Violations) < 2 {
		t.Fatalf("Violations = %+v, want appeal and max-open violations", validation.Violations)
	}
	definition.RequiresPunishment = true
	definition.MaxOpenPerSubmitter = 3
	if err := definition.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

// TestDefinitionNormalizeAndReportRules verifies defaults, trimming, and report target policy.
func TestDefinitionNormalizeAndReportRules(t *testing.T) {
	definition := Definition{
		Key:                "report",
		Name:               "  Reports  ",
		Description:        "  bad behavior  ",
		Kind:               KindReport,
		RequiresTargetUser: true,
		MetadataSchemaKey:  "  report_card  ",
	}.Normalize()
	if definition.Name != "Reports" || definition.Description != "bad behavior" {
		t.Fatalf("Normalize() = %+v, want trimmed strings", definition)
	}
	if definition.Status != DefinitionActive || definition.Version != 1 {
		t.Fatalf("Normalize() status/version = %s/%d, want active/1", definition.Status, definition.Version)
	}
	if definition.MetadataSchemaKey != "report_card" {
		t.Fatalf("MetadataSchemaKey = %q, want trimmed key", definition.MetadataSchemaKey)
	}
	if err := definition.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	definition.RequiresTargetUser = false
	if err := definition.Validate(); err == nil {
		t.Fatalf("report without target Validate() error = nil, want validation")
	}
}

// TestTicketValidationAndNormalize verifies defaulting and required fields.
func TestTicketValidationAndNormalize(t *testing.T) {
	ticket := Ticket{DefinitionID: uuid.New(), Title: "  Help  ", Kind: KindSupport, OpenedAt: time.Now()}.Normalize()
	if ticket.Status != StatusOpen || ticket.Priority != PriorityNormal || ticket.Version != 1 {
		t.Fatalf("Normalize() = %+v, want defaults", ticket)
	}
	if ticket.Title != "Help" {
		t.Fatalf("Title = %q, want trimmed", ticket.Title)
	}
	if err := ticket.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

// TestTicketValidationRejectsMissingCoreFields verifies invalid ticket state is reported.
func TestTicketValidationRejectsMissingCoreFields(t *testing.T) {
	err := Ticket{Kind: "unknown", Status: "wild"}.Validate()
	var validation ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("Validate() error = %v, want ValidationError", err)
	}
	if len(validation.Violations) < 4 {
		t.Fatalf("Violations = %+v, want core field violations", validation.Violations)
	}
}

// TestMessageAndEvidenceValidation verifies conversation content rules.
func TestMessageAndEvidenceValidation(t *testing.T) {
	message := Message{
		TicketID:            uuid.New(),
		AuthorRole:          RoleSubmitter,
		Visibility:          VisibilityParticipants,
		ContentDocumentJSON: json.RawMessage(`{"type":"doc"}`),
		ContentText:         "hello",
	}
	if err := message.Validate(); err != nil {
		t.Fatalf("message Validate() error = %v", err)
	}
	bad := message
	bad.ContentText = ""
	if err := bad.Validate(); err == nil {
		t.Fatalf("empty submitter message Validate() error = nil, want error")
	}
	evidence := Evidence{TicketID: message.TicketID, Visibility: VisibilityParticipants}
	if err := evidence.Validate(); err == nil {
		t.Fatalf("evidence Validate() error = nil, want missing target")
	}
	assetID := uuid.New()
	evidence.AssetID = &assetID
	if err := evidence.Validate(); err != nil {
		t.Fatalf("evidence Validate() error = %v", err)
	}
}

// TestSystemMessageAndActionValidation verifies workflow edge cases.
func TestSystemMessageAndActionValidation(t *testing.T) {
	system := Message{
		TicketID:            uuid.New(),
		AuthorRole:          RoleSystem,
		Visibility:          VisibilitySystemOnly,
		ContentDocumentJSON: json.RawMessage(`{"type":"doc"}`),
	}
	if err := system.Validate(); err != nil {
		t.Fatalf("system message Validate() error = %v", err)
	}
	if MessageVisibleToSubmitter(Message{Visibility: VisibilityStaffOnly}) {
		t.Fatalf("staff-only message unexpectedly visible to submitter")
	}

	action := Action{
		TicketID:    uuid.New(),
		Type:        ActionAssign,
		Status:      ActionCompleted,
		PayloadJSON: json.RawMessage(`{"assignee":"mod"}`),
	}
	if err := action.Validate(); err != nil {
		t.Fatalf("action Validate() error = %v", err)
	}
	action.PayloadJSON = json.RawMessage(`{`)
	if err := action.Validate(); err == nil {
		t.Fatalf("invalid action payload Validate() error = nil, want validation")
	}
}

// TestTransitionsAndSLA verifies state transitions and SLA due times.
func TestTransitionsAndSLA(t *testing.T) {
	if !CanTransition(StatusOpen, StatusEscalated) {
		t.Fatalf("CanTransition(open, escalated) = false, want true")
	}
	if !CanTransition(StatusClosed, StatusOpen) {
		t.Fatalf("CanTransition(closed, open) = false, want true")
	}
	if CanTransition(StatusClosed, StatusAccepted) {
		t.Fatalf("CanTransition(closed, accepted) = true, want false")
	}
	opened := time.Unix(100, 0).UTC()
	first, resolution := SLADueAt(opened, Definition{
		SLAFirstResponseSeconds: 10,
		SLAResolutionSeconds:    20,
	})
	if first == nil || !first.Equal(opened.Add(10*time.Second)) {
		t.Fatalf("first SLA = %v, want +10s", first)
	}
	if resolution == nil || !resolution.Equal(opened.Add(20*time.Second)) {
		t.Fatalf("resolution SLA = %v, want +20s", resolution)
	}
}

// TestNoColorFieldDocumentsMetadataOwnership guards the requested model shape.
func TestNoColorFieldDocumentsMetadataOwnership(t *testing.T) {
	body, err := json.Marshal(Definition{Key: "support", Name: "Support"})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if string(body) == "" || json.Valid(body) == false {
		t.Fatalf("definition JSON invalid: %s", body)
	}
	if strings.Contains(string(body), "color") {
		t.Fatalf("definition unexpectedly exposed color field")
	}
}
