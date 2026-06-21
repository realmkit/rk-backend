package application

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/punishments/domain"
	"github.com/realmkit/rk-backend/module/punishments/port"
)

// benchmarkPunishment stores the prepared punishment benchmark result.
var benchmarkPunishment domain.Punishment

// benchmarkRestrictions stores the prepared restriction benchmark result.
var benchmarkRestrictions []domain.ActiveRestriction

// benchmarkCheckResult stores the restriction check benchmark result.
var benchmarkCheckResult domain.CheckResult

// BenchmarkPrepareIssue measures punishment issue preparation and active restriction projection.
func BenchmarkPrepareIssue(b *testing.B) {
	service := Service{}
	definition := testDefinition()
	command := port.IssueCommand{
		DefinitionID:  definition.ID,
		TargetUserID:  uuid.New(),
		IssuerType:    domain.IssuerSystem,
		IssuerKey:     "realmkit",
		Reason:        "benchmark",
		PrivateReason: "benchmark private",
		Source:        "benchmark",
	}

	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		punishment, restrictions, err := service.prepareIssue(command, definition)
		if err != nil {
			b.Fatalf("prepareIssue() error = %v", err)
		}
		benchmarkPunishment = punishment
		benchmarkRestrictions = restrictions
	}
}

// BenchmarkCheckRestrictionCacheHit measures the fast cached restriction path.
func BenchmarkCheckRestrictionCacheHit(b *testing.B) {
	result := domain.CheckResult{Allowed: true}
	cache := &cacheFake{values: map[string]domain.CheckResult{"realmkit.forums.reply": result}}
	service := NewService(Dependencies{Cache: cache})
	command := port.CheckCommand{UserID: uuid.New(), ActionKey: "realmkit.forums.reply"}
	ctx := context.Background()

	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		got, err := service.CheckRestriction(ctx, command)
		if err != nil {
			b.Fatalf("CheckRestriction() error = %v", err)
		}
		benchmarkCheckResult = got
	}
}

// BenchmarkCheckRestrictionMissAllowed measures cache miss behavior for unrestricted users.
func BenchmarkCheckRestrictionMissAllowed(b *testing.B) {
	cache := &cacheFake{values: map[string]domain.CheckResult{}}
	service := NewService(Dependencies{Cases: newCaseFake(), Cache: cache})
	command := port.CheckCommand{UserID: uuid.New(), ActionKey: "realmkit.forums.reply"}
	ctx := context.Background()

	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		got, err := service.CheckRestriction(ctx, command)
		if err != nil {
			b.Fatalf("CheckRestriction() error = %v", err)
		}
		delete(cache.values, command.ActionKey)
		benchmarkCheckResult = got
	}
}
