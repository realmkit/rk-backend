package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/metadata/domain"
	"github.com/realmkit/rk-backend/module/metadata/port"
	"github.com/realmkit/rk-backend/pkg/orm"
	"github.com/realmkit/rk-backend/pkg/pagination"
	"github.com/realmkit/rk-backend/pkg/postgres/migrations"
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

// TestMetafieldDefinitionRepositoryLifecycle verifies definition read, list, update, and archive paths.
func TestMetafieldDefinitionRepositoryLifecycle(t *testing.T) {
	repository := newDefinitionRepository(t)
	definition, err := repository.Create(context.Background(), testDefinition())
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	definition.Name = "Public Motto"
	updated, err := repository.Update(context.Background(), definition, definition.Version)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.Version != definition.Version+1 {
		t.Fatalf("Update() Version = %d, want %d", updated.Version, definition.Version+1)
	}
	found, err := repository.FindByID(context.Background(), updated.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if found.Name != "Public Motto" {
		t.Fatalf("FindByID() Name = %q, want Public Motto", found.Name)
	}
	active := true
	list, err := repository.List(
		context.Background(),
		port.DefinitionFilter{OwnerType: domain.OwnerUser, Active: &active},
		testPage(),
	)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(list.Items) != 1 {
		t.Fatalf("List() items = %d, want 1", len(list.Items))
	}
	if err := repository.Archive(context.Background(), updated.ID, updated.Version); err != nil {
		t.Fatalf("Archive() error = %v", err)
	}
	if _, err := repository.FindByID(context.Background(), updated.ID); !errors.Is(err, port.ErrNotFound) {
		t.Fatalf("FindByID() archived error = %v, want %v", err, port.ErrNotFound)
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

// TestMetafieldValueRepositoryReadListDeleteAndCount verifies value query and delete paths.
func TestMetafieldValueRepositoryReadListDeleteAndCount(t *testing.T) {
	definitions, values := newDefinitionAndValueRepositories(t)
	definition, err := definitions.Create(context.Background(), testDefinition())
	if err != nil {
		t.Fatalf("Create() definition error = %v", err)
	}
	ownerID := uuid.New()
	created, _, err := values.Upsert(context.Background(), domain.MetafieldValue{
		DefinitionID: definition.ID,
		OwnerType:    definition.OwnerType,
		OwnerID:      ownerID,
		Value:        []byte(`{"value":"hello"}`),
	}, nil)
	if err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}
	found, err := values.Find(context.Background(), definition.ID, definition.OwnerType, ownerID)
	if err != nil {
		t.Fatalf("Find() error = %v", err)
	}
	if found.ID != created.ID {
		t.Fatalf("Find() ID = %s, want %s", found.ID, created.ID)
	}
	list, err := values.ListForOwner(context.Background(), definition.OwnerType, ownerID)
	if err != nil {
		t.Fatalf("ListForOwner() error = %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("ListForOwner() values = %d, want 1", len(list))
	}
	count, err := values.CountByDefinition(context.Background(), definition.ID)
	if err != nil {
		t.Fatalf("CountByDefinition() error = %v", err)
	}
	if count != 1 {
		t.Fatalf("CountByDefinition() = %d, want 1", count)
	}
	if err := values.Delete(context.Background(), definition.ID, definition.OwnerType, ownerID, created.Version); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := values.Find(context.Background(), definition.ID, definition.OwnerType, ownerID); !errors.Is(err, port.ErrNotFound) {
		t.Fatalf("Find() deleted error = %v, want %v", err, port.ErrNotFound)
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

// TestMetaobjectDefinitionRepositoryLifecycle verifies metaobject definition query and mutation paths.
func TestMetaobjectDefinitionRepositoryLifecycle(t *testing.T) {
	definitions, _ := newMetaobjectRepositories(t)
	definition, err := definitions.Create(context.Background(), testMetaobjectDefinition())
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	definition.Name = "Public Profile Card"
	updated, err := definitions.Update(context.Background(), definition, definition.Version)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.Version != definition.Version+1 {
		t.Fatalf("Update() Version = %d, want %d", updated.Version, definition.Version+1)
	}
	found, err := definitions.FindByID(context.Background(), updated.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if found.Name != "Public Profile Card" {
		t.Fatalf("FindByID() Name = %q, want Public Profile Card", found.Name)
	}
	active := true
	list, err := definitions.List(context.Background(), port.MetaobjectDefinitionFilter{Type: "profile_card", Active: &active}, testPage())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(list.Items) != 1 {
		t.Fatalf("List() items = %d, want 1", len(list.Items))
	}
	if err := definitions.Archive(context.Background(), updated.ID, updated.Version); err != nil {
		t.Fatalf("Archive() error = %v", err)
	}
	if _, err := definitions.FindByID(context.Background(), updated.ID); !errors.Is(err, port.ErrNotFound) {
		t.Fatalf("FindByID() archived error = %v, want %v", err, port.ErrNotFound)
	}
}

// TestMetaobjectEntryRepositoryLifecycle verifies metaobject entry query and mutation paths.
func TestMetaobjectEntryRepositoryLifecycle(t *testing.T) {
	definitions, entries := newMetaobjectRepositories(t)
	definition, err := definitions.Create(context.Background(), testMetaobjectDefinition())
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
	entry.DisplayName = "Updated Card"
	updated, err := entries.Update(context.Background(), entry, entry.Version)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.Version != entry.Version+1 {
		t.Fatalf("Update() Version = %d, want %d", updated.Version, entry.Version+1)
	}
	found, err := entries.FindByID(context.Background(), updated.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if found.DisplayName != "Updated Card" {
		t.Fatalf("FindByID() DisplayName = %q, want Updated Card", found.DisplayName)
	}
	list, err := entries.List(context.Background(), definition.ID, testPage())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(list.Items) != 1 {
		t.Fatalf("List() items = %d, want 1", len(list.Items))
	}
	count, err := entries.CountByDefinition(context.Background(), definition.ID)
	if err != nil {
		t.Fatalf("CountByDefinition() error = %v", err)
	}
	if count != 1 {
		t.Fatalf("CountByDefinition() = %d, want 1", count)
	}
	if err := entries.Delete(context.Background(), updated.ID, updated.Version); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := entries.FindByID(context.Background(), updated.ID); !errors.Is(err, port.ErrNotFound) {
		t.Fatalf("FindByID() deleted error = %v, want %v", err, port.ErrNotFound)
	}
}

// TestJSONScanAcceptsSupportedValues verifies JSON database scanning behavior.
func TestJSONScanAcceptsSupportedValues(t *testing.T) {
	var value JSON
	if err := value.Scan([]byte(`{"name":"Alex"}`)); err != nil {
		t.Fatalf("Scan([]byte) error = %v", err)
	}
	if string(value) != `{"name":"Alex"}` {
		t.Fatalf("value = %s, want JSON object", string(value))
	}
	if err := value.Scan(`{"name":"Sam"}`); err != nil {
		t.Fatalf("Scan(string) error = %v", err)
	}
	if string(value) != `{"name":"Sam"}` {
		t.Fatalf("value = %s, want JSON string", string(value))
	}
	if err := value.Scan(nil); err != nil {
		t.Fatalf("Scan(nil) error = %v", err)
	}
	if string(value) != "null" {
		t.Fatalf("value = %s, want null", string(value))
	}
	if err := value.Scan(42); err == nil {
		t.Fatalf("Scan(int) error = nil, want unsupported type error")
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
		Key:       "motto",
		Name:      "Motto",
		ValueType: domain.ValueSingleLineText,
		Active:    true,
		Version:   1,
	}
}

// testMetaobjectDefinition returns a valid test metaobject definition.
func testMetaobjectDefinition() domain.MetaobjectDefinition {
	return domain.MetaobjectDefinition{
		Type:    "profile_card",
		Name:    "Profile Card",
		Fields:  []domain.FieldDefinition{{Key: "motto", Name: "Motto", ValueType: domain.ValueSingleLineText}},
		Active:  true,
		Version: 1,
	}
}

// testPage returns a normalized test page.
func testPage() pagination.Page {
	return pagination.Page{Limit: pagination.DefaultLimit}
}
