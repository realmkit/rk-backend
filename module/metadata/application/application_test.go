package application

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	metadatapostgres "github.com/realmkit/rk-backend/module/metadata/adapter/postgres"
	"github.com/realmkit/rk-backend/module/metadata/domain"
	"github.com/realmkit/rk-backend/module/metadata/port"
	eventdomain "github.com/realmkit/rk-backend/pkg/events/domain"
	eventtesting "github.com/realmkit/rk-backend/pkg/events/testing"
	"github.com/realmkit/rk-backend/pkg/orm"
	"github.com/realmkit/rk-backend/pkg/pagination"
	"github.com/realmkit/rk-backend/pkg/postgres/migrations"
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
		Owner:    port.OwnerRef{Type: definition.OwnerType, ID: ownerID},
		Key:      definition.Key,
		RawValue: json.RawMessage(`"hello"`),
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
		Owner:    port.OwnerRef{Type: definition.OwnerType, ID: uuid.New()},
		Key:      definition.Key,
		RawValue: json.RawMessage(`"hello"`),
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

// TestServiceGetAndDeleteValue verifies value read and delete use cases.
func TestServiceGetAndDeleteValue(t *testing.T) {
	service := newService(t)
	definition, err := service.CreateDefinition(context.Background(), port.CreateDefinitionCommand{Definition: testDefinition()})
	if err != nil {
		t.Fatalf("CreateDefinition() error = %v", err)
	}
	ownerID := uuid.New()
	value, _, err := service.SetValue(context.Background(), port.SetValueCommand{
		Owner:    port.OwnerRef{Type: definition.OwnerType, ID: ownerID},
		Key:      definition.Key,
		RawValue: json.RawMessage(`"hello"`),
	})
	if err != nil {
		t.Fatalf("SetValue() error = %v", err)
	}
	found, err := service.GetValue(context.Background(), port.GetValueQuery{
		Owner: port.OwnerRef{Type: definition.OwnerType, ID: ownerID},
		Key:   definition.Key,
	})
	if err != nil {
		t.Fatalf("GetValue() error = %v", err)
	}
	if found.ID != value.ID {
		t.Fatalf("GetValue() ID = %s, want %s", found.ID, value.ID)
	}
	err = service.DeleteValue(context.Background(), port.DeleteValueCommand{
		Owner:           port.OwnerRef{Type: definition.OwnerType, ID: ownerID},
		Key:             definition.Key,
		ExpectedVersion: value.Version,
	})
	if err != nil {
		t.Fatalf("DeleteValue() error = %v", err)
	}
	if _, err := service.GetValue(context.Background(), port.GetValueQuery{Owner: port.OwnerRef{Type: definition.OwnerType, ID: ownerID}, Key: definition.Key}); !errors.Is(
		err,
		port.ErrNotFound,
	) {
		t.Fatalf("GetValue() deleted error = %v, want %v", err, port.ErrNotFound)
	}
}

// TestServicePublishesMetadataEvents verifies metadata writes emit event facts.
func TestServicePublishesMetadataEvents(t *testing.T) {
	events := &eventtesting.PublisherRecorder{}
	service := newServiceWithEvents(t, events)
	actor := port.Actor{ID: uuid.New()}
	definition, err := service.CreateDefinition(context.Background(), port.CreateDefinitionCommand{
		Actor:      actor,
		Definition: testDefinition(),
	})
	if err != nil {
		t.Fatalf("CreateDefinition() error = %v", err)
	}
	ownerID := uuid.New()
	value, _, err := service.SetValue(context.Background(), port.SetValueCommand{
		Actor:    actor,
		Owner:    port.OwnerRef{Type: definition.OwnerType, ID: ownerID},
		Key:      definition.Key,
		RawValue: json.RawMessage(`"hello"`),
	})
	if err != nil {
		t.Fatalf("SetValue() error = %v", err)
	}
	if err := service.DeleteValue(context.Background(), port.DeleteValueCommand{
		Actor:           actor,
		Owner:           port.OwnerRef{Type: definition.OwnerType, ID: ownerID},
		Key:             definition.Key,
		ExpectedVersion: value.Version,
	}); err != nil {
		t.Fatalf("DeleteValue() error = %v", err)
	}
	assertMetadataEventKeys(t, events.Drafts(), []string{
		"metadata.definition.created",
		"metadata.metafield.set",
		"metadata.entry.created",
		"metadata.metafield.deleted",
		"metadata.entry.deleted",
	})
}

// assertMetadataEventKeys verifies event draft key order.
func assertMetadataEventKeys(t *testing.T, drafts []eventdomain.Draft, want []string) {
	t.Helper()
	if len(drafts) != len(want) {
		t.Fatalf("event count = %d, want %d", len(drafts), len(want))
	}
	for index, key := range want {
		if string(drafts[index].Key) != key {
			t.Fatalf("event[%d] = %s, want %s", index, drafts[index].Key, key)
		}
	}
}

// TestServiceSetValueValidatesReferenceLists verifies list reference targets are checked.
func TestServiceSetValueValidatesReferenceLists(t *testing.T) {
	service := newServiceWithReferenceResolver(t, missingReferenceResolver{})
	definition, err := service.CreateDefinition(context.Background(), port.CreateDefinitionCommand{Definition: domain.MetafieldDefinition{
		OwnerType: domain.OwnerUser,
		Key:       "friends",
		Name:      "Friends",
		ValueType: domain.ValueOwnerReference,
		List:      true,
		Active:    true,
		Version:   1,
	}})
	if err != nil {
		t.Fatalf("CreateDefinition() error = %v", err)
	}
	raw, err := json.Marshal([]domain.OwnerReference{{Type: domain.OwnerUser, ID: uuid.New()}})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	_, _, err = service.SetValue(context.Background(), port.SetValueCommand{
		Owner:    port.OwnerRef{Type: definition.OwnerType, ID: uuid.New()},
		Key:      definition.Key,
		RawValue: raw,
	})
	if !errors.Is(err, port.ErrNotFound) {
		t.Fatalf("SetValue() reference error = %v, want %v", err, port.ErrNotFound)
	}
}

// TestServiceSetValueValidatesMetaobjectReferences verifies metaobject reference targets are checked.
func TestServiceSetValueValidatesMetaobjectReferences(t *testing.T) {
	service := newServiceWithReferenceResolver(t, missingReferenceResolver{})
	definition, err := service.CreateDefinition(context.Background(), port.CreateDefinitionCommand{Definition: domain.MetafieldDefinition{
		OwnerType: domain.OwnerUser,
		Key:       "card",
		Name:      "Card",
		ValueType: domain.ValueMetaobjectReference,
		Active:    true,
		Version:   1,
	}})
	if err != nil {
		t.Fatalf("CreateDefinition() error = %v", err)
	}
	raw, err := json.Marshal(domain.MetaobjectReference{DefinitionID: uuid.New(), EntryID: uuid.New()})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	_, _, err = service.SetValue(context.Background(), port.SetValueCommand{
		Owner:    port.OwnerRef{Type: definition.OwnerType, ID: uuid.New()},
		Key:      definition.Key,
		RawValue: raw,
	})
	if !errors.Is(err, port.ErrNotFound) {
		t.Fatalf("SetValue() reference error = %v, want %v", err, port.ErrNotFound)
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
		Owner:    port.OwnerRef{Type: definition.OwnerType, ID: uuid.New()},
		Key:      definition.Key,
		RawValue: json.RawMessage(`"hello"`),
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
	found, err := service.GetDefinition(context.Background(), port.GetDefinitionQuery{ID: updated.ID})
	if err != nil {
		t.Fatalf("GetDefinition() error = %v", err)
	}
	if found.ID != updated.ID {
		t.Fatalf("GetDefinition() ID = %s, want %s", found.ID, updated.ID)
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
	definitions, err := service.ListMetaobjectDefinitions(context.Background(), port.ListMetaobjectDefinitionsQuery{Page: testPage()})
	if err != nil {
		t.Fatalf("ListMetaobjectDefinitions() error = %v", err)
	}
	if len(definitions.Items) != 1 {
		t.Fatalf("ListMetaobjectDefinitions() items = %d, want 1", len(definitions.Items))
	}
	foundDefinition, err := service.GetMetaobjectDefinition(
		context.Background(),
		port.GetMetaobjectDefinitionQuery{ID: updatedDefinition.ID},
	)
	if err != nil {
		t.Fatalf("GetMetaobjectDefinition() error = %v", err)
	}
	if foundDefinition.ID != updatedDefinition.ID {
		t.Fatalf("GetMetaobjectDefinition() ID = %s, want %s", foundDefinition.ID, updatedDefinition.ID)
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
	foundEntry, err := service.GetMetaobjectEntry(context.Background(), port.GetMetaobjectEntryQuery{ID: updatedEntry.ID})
	if err != nil {
		t.Fatalf("GetMetaobjectEntry() error = %v", err)
	}
	if foundEntry.ID != updatedEntry.ID {
		t.Fatalf("GetMetaobjectEntry() ID = %s, want %s", foundEntry.ID, updatedEntry.ID)
	}
	list, err := service.ListMetaobjectEntries(
		context.Background(),
		port.ListMetaobjectEntriesQuery{DefinitionID: updatedDefinition.ID, Page: testPage()},
	)
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

// missingReferenceResolver reports every reference as missing.
type missingReferenceResolver struct{}

// OwnerExists reports whether owner reference exists.
func (missingReferenceResolver) OwnerExists(context.Context, domain.OwnerReference) (bool, error) {
	return false, nil
}

// MetaobjectEntryExists reports whether metaobject entry reference exists.
func (missingReferenceResolver) MetaobjectEntryExists(context.Context, domain.MetaobjectReference) (bool, error) {
	return false, nil
}

// TestExistingResolversReportExistence verifies default resolvers accept references.
func TestExistingResolversReportExistence(t *testing.T) {
	ownerID := uuid.New()
	if ok, err := (ExistingOwnerResolver{}).Exists(context.Background(), domain.OwnerUser, ownerID); err != nil || !ok {
		t.Fatalf("ExistingOwnerResolver.Exists() = (%v, %v), want true nil", ok, err)
	}
	if ok, err := (ExistingReferenceResolver{}).OwnerExists(context.Background(), domain.OwnerReference{Type: domain.OwnerUser, ID: ownerID}); err != nil ||
		!ok {
		t.Fatalf("ExistingReferenceResolver.OwnerExists() = (%v, %v), want true nil", ok, err)
	}
	if ok, err := (ExistingReferenceResolver{}).MetaobjectEntryExists(context.Background(), domain.MetaobjectReference{DefinitionID: uuid.New(), EntryID: uuid.New()}); err != nil ||
		!ok {
		t.Fatalf("ExistingReferenceResolver.MetaobjectEntryExists() = (%v, %v), want true nil", ok, err)
	}
}

// newService creates a metadata service backed by test repositories.
func newService(t *testing.T) Service {
	t.Helper()
	return newServiceWithOwnerResolver(t, ExistingOwnerResolver{})
}

// newServiceWithReferenceResolver creates a metadata service with reference resolver.
func newServiceWithReferenceResolver(t *testing.T, resolver port.ReferenceResolver) Service {
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
		References:            resolver,
	})
}

// newServiceWithEvents creates a metadata service with an event recorder.
func newServiceWithEvents(t *testing.T, events *eventtesting.PublisherRecorder) Service {
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
		Events:                events,
	})
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
