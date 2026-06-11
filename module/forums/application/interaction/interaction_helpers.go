package interaction

import (
	"context"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/forums/domain"
	"github.com/realmkit/rk-backend/module/forums/port"
	"github.com/realmkit/rk-backend/pkg/pagination"
)

func newPostLike(post domain.Post, actorUserID uuid.UUID) domain.PostLike {
	return domain.PostLike{
		ID:        uuid.New(),
		PostID:    post.ID,
		ThreadID:  post.ThreadID,
		ForumID:   post.ForumID,
		UserID:    actorUserID,
		CreatedAt: time.Now().UTC(),
	}
}

func (service Service) postLikeSummary(
	ctx context.Context,
	postID uuid.UUID,
	likedByActor bool,
) (domain.PostLikeSummary, error) {
	updated, err := service.posts.FindByID(ctx, postID)
	if err != nil {
		return domain.PostLikeSummary{}, err
	}
	return domain.PostLikeSummary{
		PostID:       updated.ID,
		LikeCount:    updated.LikeCount,
		LikedByActor: likedByActor,
	}, nil
}

func threadReadState(
	command port.MarkThreadReadCommand,
	thread domain.Thread,
) domain.ThreadReadState {
	sequence := command.LastReadPostSequence
	if sequence < 1 {
		sequence = thread.VisiblePostCount
	}
	return domain.ThreadReadState{
		ID:                   uuid.New(),
		UserID:               command.ActorUserID,
		ForumID:              thread.ForumID,
		ThreadID:             thread.ID,
		LastReadPostSequence: sequence,
		LastReadAt:           time.Now().UTC(),
	}
}

func latestPostsCacheKey(
	actorUserID uuid.UUID,
	forumID uuid.UUID,
	page pagination.Page,
) string {
	return "forums:latest:v1:" +
		widgetScope(forumID) +
		":" +
		actorScope(actorUserID) +
		":" +
		page.Cursor +
		":" +
		strconv.Itoa(page.Limit)
}

func mostLikedCacheKey(
	actorUserID uuid.UUID,
	forumID uuid.UUID,
	page pagination.Page,
) string {
	return "forums:most-liked:v1:" +
		forumID.String() +
		":all:" +
		actorScope(actorUserID) +
		":" +
		page.Cursor +
		":" +
		strconv.Itoa(page.Limit)
}

func widgetScope(forumID uuid.UUID) string {
	if forumID == uuid.Nil {
		return "global:all"
	}
	return "forum:" + forumID.String()
}

func actorScope(actorUserID uuid.UUID) string {
	if actorUserID == uuid.Nil {
		return "anonymous"
	}
	return "user:" + actorUserID.String()
}
