package application

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	metadatapostgres "github.com/niflaot/gamehub-go/module/metadata/adapter/postgres"
	"github.com/niflaot/gamehub-go/module/metadata/domain"
	"github.com/niflaot/gamehub-go/module/metadata/port"
	"github.com/niflaot/gamehub-go/pkg/orm"
	"github.com/niflaot/gamehub-go/pkg/pagination"
	"github.com/niflaot/gamehub-go/pkg/postgres/migrations"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestServiceSetValueCreatesCanonicalValue verifies definition lookup and normalization.
func TestServiceSetValueCreatesCanonicalValue(t *testing.T) {
	service := newService(t)
	definition, err := service.CreateDefinition(context.Background(), port.CreateDefinitionCommand{Definition: testDefinition()})
	if err != nil {
		t.Fatalf("CreateDefinition() error = %v", err)
	}
	ownerID := uuid.New()

	value, created, err := service.SetValue(context.Background(), port.SetValueCommand{
		Owner:     port.OwnerRef{Type: definition.OwnerType, ID: ownerID},
		Namespace: definition.Namespace,
		Key:       definition.Key,
		RawValue:  json.RawMessage(`"hello"`),
	})
	if err != nil {
		t.Fatalf("SetValue() error = %v", err)
	}
	if !created {
		t.Fatalf("SetValue() created = false, want true")
	}
	if string(value.Value) != `{"value":"hello"}` {
		t.Fatalf("Value = %s", value.Value)
	}
}

// TestServiceSetValueRejectsMissingOwner verifies owner resolution is enforced.
func TestServiceSetValueRejectsMissingOwner(t *testing.T) {
	service := newServiceWithOwnerResolver(t, missingOwnerResolver{})
	definition, err := service.CreateDefinition(context.Background(), port.CreateDefinitionCommand{Definition: testDefinition()})
	if err != nil {
		t.Fatalf("CreateDefinition() error = %v", err)
	}

	_, _, err = service.SetValue(context.Background(), port.SetValueCommand{
		Owner:     port.OwnerRef{Type: definition.OwnerType, ID: uuid.New()},
		Namespace: definition.Namespace,
		Key:       definition.Key,
		RawValue:  json.RawMessage(`"hello"`),
	})
	if !errors.Is(err, port.ErrNotFound) {
		t.Fatalf("SetValue() error = %v, want %v", err, port.ErrNotFound)
	}
}

// TestServiceListValuesForOwnerIncludesEmpty verifies absent values remain visible.
func TestServiceListValuesForOwnerIncludesEmpty(t *testing.T) {
	service := newService(t)
	definition, err := service.CreateDefinition(context.Background(), port.CreateDefinitionCommand{Definition: testDefinition()})
	if err != nil {
		t.Fatalf("CreateDefinition() error = %v", err)
	}

	view, err := service.ListValuesForOwner(context.Background(), port.ListValuesForOwnerQuery{
		Owner:        port.OwnerRef{Type: definition.OwnerType, ID: uuid.New()},
		IncludeEmpty: true,
	})
	if err != nil {
		t.Fatalf("ListValuesForOwner() error = %v", err)
	}
	if len(view.Metafields) != 1 || view.Metafields[0].Value != nil {
		t.Fatalf("Metafields = %+v, want one empty value", view.Metafields)
	}
}

// TestServiceArchiveDefinitionRejectsActiveValues verifies referenced definitions are protected.
func TestServiceArchiveDefinitionRejectsActiveValues(t *testing.T) {
	service := newService(t)
	definition, err := service.CreateDefinition(context.Background(), port.CreateDefinitionCommand{Definition: testDefinition()})
	if err != nil {
		t.Fatalf("CreateDefinition() error = %v", err)
	}
	_, _, err = service.SetValue(context.Background(), port.SetValueCommand{
		Owner:     port.OwnerRef{Type: definition.OwnerType, ID: uuid.New()},
		Namespace: definition.Namespace,
		Key:       definition.Key,
		RawValue:  json.RawMessage(`"hello"`),
	})
	if err != nil {
		t.Fatalf("SetValue() error = %v", err)
	}

	err = service.ArchiveDefinition(context.Background(), port.ArchiveDefinitionCommand{
		ID:              definition.ID,
		ExpectedVersion: definition.Version,
	})
	if !errors.Is(err, port.ErrReferenced) {
		t.Fatalf("ArchiveDefinition() error = %v, want %v", err, port.ErrReferenced)
	}
}

// TestServiceCreateMetaobjectEntryValidatesFields verifies entry schema validation.
func TestServiceCreateMetaobjectEntryValidatesFields(t *testing.T) {
	service := newService(t)
	definition, err := service.CreateMetaobjectDefinition(context.Background(), port.CreateMetaobjectDefinitionCommand{
		Definition: domain.MetaobjectDefinition{
			Type:    "profile_card",
			Name:    "Profile Card",
			Fields:  []domain.FieldDefinition{{Key: "motto", Name: "Motto", ValueType: domain.ValueSingleLineText, Required: true}},
			Active:  true,
			Version: 1,
		},
	})
	if err != nil {
		t.Fatalf("CreateMetaobjectDefinition() error = %v", err)
	}

	_, err = service.CreateMetaobjectEntry(context.Background(), port.CreateMetaobjectEntryCommand{
		Entry: domain.MetaobjectEntry{
			DefinitionID: definition.ID,
			Handle:       "first_card",
			DisplayName:  "First Card",
		},
		RawFields: map[domain.Key]json.RawMessage{},
	})
	if !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("CreateMetaobjectEntry() error = %v, want %v", err, domain.ErrInvalid)
	}
}

// TestServiceDefinitionUpdateListAndArchive verifies definition management use cases.
func TestServiceDefinitionUpdateListAndArchive(t *testing.T) {
	service := newService(t)
	definition, err := service.CreateDefinition(context.Background(), port.CreateDefinitionCommand{Definition: testDefinition()})
	if err != nil {
		t.Fatalf("CreateDefinition() error = %v", err)
	}
	definition.Name = "Public Motto"
	updated, err := service.UpdateDefinition(context.Background(), port.UpdateDefinitionCommand{
		Definition:      definition,
		ExpectedVersion: definition.Version,
	})
	if err != nil {
		t.Fatalf("UpdateDefinition() error = %v", err)
	}
	if updated.Version != definition.Version+1 {
		t.Fatalf("updated Version = %d, want %d", updated.Version, definition.Version+1)
	}

	list, err := service.ListDefinitions(context.Background(), port.ListDefinitionsQuery{Page: testPage()})
	if err != nil {
		t.Fatalf("ListDefinitions() error = %v", err)
	}
	if len(list.Items) != 1 {
		t.Fatalf("ListDefinitions() items = %d, want 1", len(list.Items))
	}

	if err := service.ArchiveDefinition(context.Background(), port.ArchiveDefinitionCommand{ID: updated.ID, ExpectedVersion: updated.Version}); err != nil {
		t.Fatalf("ArchiveDefinition() error = %v", err)
	}
}

// TestServiceMetaobjectLifecycle verifies metaobject management use cases.
func TestServiceMetaobjectLifecycle(t *testing.T) {
	service := newService(t)
	definition, err := service.CreateMetaobjectDefinition(context.Background(), port.CreateMetaobjectDefinitionCommand{
		Definition: domain.MetaobjectDefinition{
			Type:    "profile_card",
			Name:    "Profile Card",
			Fields:  []domain.FieldDefinition{{Key: "motto", Name: "Motto", ValueType: domain.ValueSingleLineText, Required: true}},
			Active:  true,
			Version: 1,
		},
	})
	if err != nil {
		t.Fatalf("CreateMetaobjectDefinition() error = %v", err)
	}
	definition.Name = "Profile Card Updated"
	updatedDefinition, err := service.UpdateMetaobjectDefinition(context.Background(), port.UpdateMetaobjectDefinitionCommand{
		Definition:      definition,
		ExpectedVersion: definition.Version,
	})
	if err != nil {
		t.Fatalf("UpdateMetaobjectDefinition() error = %v", err)
	}
	entry, err := service.CreateMetaobjectEntry(context.Background(), port.CreateMetaobjectEntryCommand{
		Entry: domain.MetaobjectEntry{
			DefinitionID: updatedDefinition.ID,
			Handle:       "first_card",
			DisplayName:  "First Card",
		},
		RawFields: map[domain.Key]json.RawMessage{"motto": json.RawMessage(`"Ready"`)},
	})
	if err != nil {
		t.Fatalf("CreateMetaobjectEntry() error = %v", err)
	}
	entry.DisplayName = "Updated Card"
	updatedEntry, err := service.UpdateMetaobjectEntry(context.Background(), port.UpdateMetaobjectEntryCommand{
		Entry:           entry,
		RawFields:       map[domain.Key]json.RawMessage{"motto": json.RawMessage(`"Still ready"`)},
		ExpectedVersion: entry.Version,
	})
	if err != nil {
		t.Fatalf("UpdateMetaobjectEntry() error = %v", err)
	}
	list, err := service.ListMetaobjectEntries(context.Background(), port.ListMetaobjectEntriesQuery{DefinitionID: updatedDefinition.ID, Page: testPage()})
	if err != nil {
		t.Fatalf("ListMetaobjectEntries() error = %v", err)
	}
	if len(list.Items) != 1 {
		t.Fatalf("ListMetaobjectEntries() items = %d, want 1", len(list.Items))
	}
	if err := service.DeleteMetaobjectEntry(context.Background(), port.DeleteMetaobjectEntryCommand{ID: updatedEntry.ID, ExpectedVersion: updatedEntry.Version}); err != nil {
		t.Fatalf("DeleteMetaobjectEntry() error = %v", err)
	}
	if err := service.ArchiveMetaobjectDefinition(context.Background(), port.ArchiveMetaobjectDefinitionCommand{ID: updatedDefinition.ID, ExpectedVersion: updatedDefinition.Version}); err != nil {
		t.Fatalf("ArchiveMetaobjectDefinition() error = %v", err)
	}
}

// missingOwnerResolver reports every owner as missing.
type missingOwnerResolver struct{}

// Exists reports whether owner exists.
func (missingOwnerResolver) Exists(context.Context, domain.OwnerType, uuid.UUID) (bool, error) {
	return false, nil
}

// newService creates a metadata service backed by test repositories.
func newService(t *testing.T) Service {
	t.Helper()
	return newServiceWithOwnerResolver(t, ExistingOwnerResolver{})
}

// newServiceWithOwnerResolver creates a metadata service with owner resolver.
func newServiceWithOwnerResolver(t *testing.T, resolver port.OwnerResolver) Service {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	if _, err := migrations.NewRunner(db, migrations.DefaultSource()).Up(context.Background()); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}
	store := orm.NewStore(db)
	return NewService(Dependencies{
		Definitions:           metadatapostgres.NewMetafieldDefinitionRepository(store),
		Values:                metadatapostgres.NewMetafieldValueRepository(store),
		MetaobjectDefinitions: metadatapostgres.NewMetaobjectDefinitionRepository(store),
		MetaobjectEntries:     metadatapostgres.NewMetaobjectEntryRepository(store),
		Owners:                resolver,
	})
}

// testDefinition returns a valid test definition.
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

// testPage returns a normalized test page.
func testPage() pagination.Page {
	return pagination.Page{Limit: pagination.DefaultLimit}
}
