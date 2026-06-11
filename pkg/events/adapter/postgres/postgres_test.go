package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/pkg/events/domain"
	"github.com/realmkit/rk-backend/pkg/events/port"
	"github.com/realmkit/rk-backend/pkg/orm"
	"github.com/realmkit/rk-backend/pkg/pagination"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestRepositoryPublishGetAndList verifies event persistence lifecycle.
func TestRepositoryPublishGetAndList(t *testing.T) {
	repo := newEventRepository(t)
	event, err := repo.Publish(context.Background(), testEventDraft(), testEventNow())
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}
	got, err := repo.Get(context.Background(), event.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Key != domain.EventForumsThreadCreated || len(got.Scopes) != 1 {
		t.Fatalf("event = %+v, want thread event with scope", got)
	}
	list, err := repo.List(context.Background(), port.ListFilter{Producer: domain.ProducerForums}, pagination.Page{Limit: 10})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(list.Items) != 1 {
		t.Fatalf("List() = %d items, want 1", len(list.Items))
	}
}

// TestRepositoryListIsBounded verifies operator event lists cannot return unbounded rows.
func TestRepositoryListIsBounded(t *testing.T) {
	repo := newEventRepository(t)
	for index := 0; index < 3; index++ {
		draft := testEventDraft()
		id := uuid.New()
		draft.AggregateID = &id
		draft.Payload = map[string]any{"thread_id": id.String()}
		if _, err := repo.Publish(context.Background(), draft, testEventNow().Add(time.Duration(index)*time.Second)); err != nil {
			t.Fatalf("Publish(%d) error = %v", index, err)
		}
	}

	list, err := repo.List(context.Background(), port.ListFilter{}, pagination.Page{Limit: 2})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(list.Items) != 2 {
		t.Fatalf("List() items = %d, want bounded limit 2", len(list.Items))
	}
}

// TestRepositoryClaimAndStatusUpdates verifies dispatch state transitions.
func TestRepositoryClaimAndStatusUpdates(t *testing.T) {
	repo := newEventRepository(t)
	event, _ := repo.Publish(context.Background(), testEventDraft(), testEventNow())

	claimed, err := repo.Claim(context.Background(), "worker", 10, testEventNow(), testEventNow().Add(time.Minute))
	if err != nil {
		t.Fatalf("Claim() error = %v", err)
	}
	if len(claimed) != 1 || claimed[0].AttemptCount != 1 {
		t.Fatalf("claimed = %+v, want one claimed attempt", claimed)
	}
	if err := repo.MarkProcessed(context.Background(), event.ID, testEventNow()); err != nil {
		t.Fatalf("MarkProcessed() error = %v", err)
	}
	processed, err := repo.Get(context.Background(), event.ID)
	if err != nil {
		t.Fatalf("Get processed error = %v", err)
	}
	if processed.Status != domain.StatusProcessed {
		t.Fatalf("status = %s, want processed", processed.Status)
	}
}

// TestRepositoryClaimDoesNotReclaimProcessingEvent verifies retry workers cannot double-claim one event.
func TestRepositoryClaimDoesNotReclaimProcessingEvent(t *testing.T) {
	repo := newEventRepository(t)
	if _, err := repo.Publish(context.Background(), testEventDraft(), testEventNow()); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}
	if _, err := repo.Claim(context.Background(), "worker-1", 1, testEventNow(), testEventNow().Add(time.Minute)); err != nil {
		t.Fatalf("Claim() error = %v", err)
	}

	claimed, err := repo.Claim(context.Background(), "worker-2", 1, testEventNow(), testEventNow().Add(time.Minute))
	if err != nil {
		t.Fatalf("Claim() second error = %v", err)
	}
	if len(claimed) != 0 {
		t.Fatalf("claimed second = %+v, want no duplicate claim", claimed)
	}
}

// TestRepositoryReplayAndCancel verifies operator transitions.
func TestRepositoryReplayAndCancel(t *testing.T) {
	repo := newEventRepository(t)
	event, _ := repo.Publish(context.Background(), testEventDraft(), testEventNow())
	if err := repo.Cancel(context.Background(), event.ID, testEventNow()); err != nil {
		t.Fatalf("Cancel() error = %v", err)
	}
	if err := repo.Replay(context.Background(), event.ID, testEventNow()); err != nil {
		t.Fatalf("Replay() error = %v", err)
	}
	got, _ := repo.Get(context.Background(), event.ID)
	if got.Status != domain.StatusPending {
		t.Fatalf("status = %s, want pending", got.Status)
	}
}

// newEventRepository creates a SQLite-backed repository.
func newEventRepository(t *testing.T) Repository {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite error = %v", err)
	}
	if err := db.AutoMigrate(&EventModel{}, &ScopeModel{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	return NewRepository(orm.NewStore(db))
}

// testEventDraft returns a valid event draft.
func testEventDraft() domain.Draft {
	id := uuid.New()
	return domain.Draft{
		Key:           domain.EventForumsThreadCreated,
		SchemaVersion: 1,
		Producer:      domain.ProducerForums,
		AggregateType: "forum_thread",
		AggregateID:   &id,
		Payload:       map[string]any{"thread_id": id.String()},
		Metadata:      map[string]any{},
		Scopes:        []domain.Scope{{Type: domain.ScopeThread, ID: id.String()}},
	}
}

// testEventNow returns deterministic time.
func testEventNow() time.Time {
	return time.Unix(300, 0).UTC()
}
