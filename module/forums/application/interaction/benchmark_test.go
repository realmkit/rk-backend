package interaction

import (
	"testing"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/forums/domain"
	"github.com/realmkit/rk-backend/module/forums/port"
	"github.com/realmkit/rk-backend/pkg/pagination"
)

// benchmarkPostLike stores the interaction like benchmark result.
var benchmarkPostLike domain.PostLike

// benchmarkInteractionKey stores the interaction cache key benchmark result.
var benchmarkInteractionKey string

// benchmarkReadState stores the thread read state benchmark result.
var benchmarkReadState domain.ThreadReadState

// BenchmarkNewPostLike measures like aggregate construction.
func BenchmarkNewPostLike(b *testing.B) {
	post := domain.Post{ID: uuid.New(), ThreadID: uuid.New(), ForumID: uuid.New()}
	actorID := uuid.New()

	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		benchmarkPostLike = newPostLike(post, actorID)
	}
}

// BenchmarkInteractionCacheKeys measures latest-post and most-liked widget key construction.
func BenchmarkInteractionCacheKeys(b *testing.B) {
	actorID := uuid.New()
	forumID := uuid.New()
	page := pagination.Page{Cursor: "cursor:100", Limit: 25}

	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		benchmarkInteractionKey = latestPostsCacheKey(actorID, forumID, page)
		benchmarkInteractionKey = mostLikedCacheKey(actorID, forumID, page)
	}
}

// BenchmarkThreadReadState measures read-state normalization from command and thread data.
func BenchmarkThreadReadState(b *testing.B) {
	thread := domain.Thread{ID: uuid.New(), ForumID: uuid.New(), VisiblePostCount: 42}
	command := port.MarkThreadReadCommand{ActorUserID: uuid.New(), ThreadID: thread.ID}

	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		benchmarkReadState = threadReadState(command, thread)
	}
}
