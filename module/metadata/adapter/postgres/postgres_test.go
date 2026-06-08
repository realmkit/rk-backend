package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/metadata/domain"
	"github.com/niflaot/gamehub-go/module/metadata/port"
	"github.com/niflaot/gamehub-go/pkg/orm"
	"github.com/niflaot/gamehub-go/pkg/postgres/migrations"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestMetafieldDefinitionRepositoryCreateRejectsDuplicate verifies unique definition keys.
func TestMetafieldDefinitionRepositoryCreateRejectsDuplicate(t *testing.T) {
	repository := newDefinitionRepository(t)
	definition := testDefinition()

	if _, err := repository.Create(context.Background(), definition); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if _, err := repository.Create(context.Background(), definition); !errors.Is(err, port.ErrConflict) {
		t.Fatalf("Create() duplicate error = %v, want %v", err, port.ErrConflict)
	}
}

// TestMetafieldValueRepositoryUpsertCreatesAndUpdates verifies value upsert semantics.
func TestMetafieldValueRepositoryUpsertCreatesAndUpdates(t *testing.T) {
	definitions, values := newDefinitionAndValueRepositories(t)
	definition, err := definitions.Create(context.Background(), testDefinition())
	if err != nil {
		t.Fatalf("Create() definition error = %v", err)
	}

	ownerID := uuid.New()
	created, isCreated, err := values.Upsert(context.Background(), domain.MetafieldValue{
		DefinitionID: definition.ID,
		OwnerType:    definition.OwnerType,
		OwnerID:      ownerID,
		Value:        []byte(`{"value":"hello"}`),
	}, nil)
	if err != nil {
		t.Fatalf("Upsert() create error = %v", err)
	}
	if !isCreated || created.Version != 1 {
		t.Fatalf("Upsert() create = (%v, %d), want created version 1", isCreated, created.Version)
	}

	expected := created.Version
	updated, isCreated, err := values.Upsert(context.Background(), domain.MetafieldValue{
		DefinitionID: definition.ID,
		OwnerType:    definition.OwnerType,
		OwnerID:      ownerID,
		Value:        []byte(`{"value":"updated"}`),
	}, &expected)
	if err != nil {
		t.Fatalf("Upsert() update error = %v", err)
	}
	if isCreated || updated.Version != 2 {
		t.Fatalf("Upsert() update = (%v, %d), want updated version 2", isCreated, updated.Version)
	}
}

// TestMetafieldValueRepositoryUpsertRejectsStaleVersion verifies optimistic concurrency.
func TestMetafieldValueRepositoryUpsertRejectsStaleVersion(t *testing.T) {
	definitions, values := newDefinitionAndValueRepositories(t)
	definition, err := definitions.Create(context.Background(), testDefinition())
	if err != nil {
		t.Fatalf("Create() definition error = %v", err)
	}
	ownerID := uuid.New()
	_, _, err = values.Upsert(context.Background(), domain.MetafieldValue{
		DefinitionID: definition.ID,
		OwnerType:    definition.OwnerType,
		OwnerID:      ownerID,
		Value:        []byte(`{"value":"hello"}`),
	}, nil)
	if err != nil {
		t.Fatalf("Upsert() create error = %v", err)
	}

	stale := uint64(999)
	_, _, err = values.Upsert(context.Background(), domain.MetafieldValue{
		DefinitionID: definition.ID,
		OwnerType:    definition.OwnerType,
		OwnerID:      ownerID,
		Value:        []byte(`{"value":"updated"}`),
	}, &stale)
	if !errors.Is(err, port.ErrPreconditionFailed) {
		t.Fatalf("Upsert() stale error = %v, want %v", err, port.ErrPreconditionFailed)
	}
}

// TestMetaobjectEntryRepositoryCreateAndFind verifies metaobject entry round trips.
func TestMetaobjectEntryRepositoryCreateAndFind(t *testing.T) {
	definitions, entries := newMetaobjectRepositories(t)
	definition, err := definitions.Create(context.Background(), domain.MetaobjectDefinition{
		Type:    "profile_card",
		Name:    "Profile Card",
		Fields:  []domain.FieldDefinition{{Key: "motto", Name: "Motto", ValueType: domain.ValueSingleLineText}},
		Active:  true,
		Version: 1,
	})
	if err != nil {
		t.Fatalf("Create() definition error = %v", err)
	}
	entry, err := entries.Create(context.Background(), domain.MetaobjectEntry{
		DefinitionID: definition.ID,
		Handle:       "first_card",
		DisplayName:  "First Card",
		Fields:       map[domain.Key]json.RawMessage{"motto": json.RawMessage(`{"value":"hi"}`)},
		Version:      1,
	})
	if err != nil {
		t.Fatalf("Create() entry error = %v", err)
	}

	found, err := entries.FindByHandle(context.Background(), definition.ID, "first_card")
	if err != nil {
		t.Fatalf("FindByHandle() error = %v", err)
	}
	if found.ID != entry.ID {
		t.Fatalf("FindByHandle() ID = %s, want %s", found.ID, entry.ID)
	}
}

// newDefinitionRepository creates a migrated definition repository.
func newDefinitionRepository(t *testing.T) MetafieldDefinitionRepository {
	t.Helper()
	return NewMetafieldDefinitionRepository(orm.NewStore(newDB(t)))
}

// newDefinitionAndValueRepositories creates migrated definition and value repositories.
func newDefinitionAndValueRepositories(t *testing.T) (MetafieldDefinitionRepository, MetafieldValueRepository) {
	t.Helper()
	store := orm.NewStore(newDB(t))
	return NewMetafieldDefinitionRepository(store), NewMetafieldValueRepository(store)
}

// newMetaobjectRepositories creates migrated metaobject repositories.
func newMetaobjectRepositories(t *testing.T) (MetaobjectDefinitionRepository, MetaobjectEntryRepository) {
	t.Helper()
	store := orm.NewStore(newDB(t))
	return NewMetaobjectDefinitionRepository(store), NewMetaobjectEntryRepository(store)
}

// newDB creates a migrated in-memory SQLite database.
func newDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	if _, err := migrations.NewRunner(db, migrations.DefaultSource()).Up(context.Background()); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}
	return db
}

// testDefinition returns a valid test metafield definition.
func testDefinition() domain.MetafieldDefinition {
	return domain.MetafieldDefinition{
		OwnerType: domain.OwnerUser,
		Namespace: "profile",
		Key:       "motto",
		Name:      "Motto",
		ValueType: domain.ValueSingleLineText,
		Active:    true,
		Version:   1,
	}
}
