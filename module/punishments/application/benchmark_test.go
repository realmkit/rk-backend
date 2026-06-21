package application

import (
	"testing"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/punishments/domain"
	"github.com/realmkit/rk-backend/module/punishments/port"
)

// benchmarkPunishment stores the prepared punishment benchmark result.
var benchmarkPunishment domain.Punishment

// benchmarkRestrictions stores the prepared restriction benchmark result.
var benchmarkRestrictions []domain.ActiveRestriction

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
