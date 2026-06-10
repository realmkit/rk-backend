package postgres

import (
	"encoding/json"

	"github.com/niflaot/gamehub-go/module/forums/domain"
	"github.com/niflaot/gamehub-go/pkg/orm"
)

// forumModelFromDomain maps forum to persistence.
func forumModelFromDomain(forum domain.Forum) ForumModel {
	return ForumModel{
		ID:                            orm.ID{ID: forum.ID},
		CategoryID:                    forum.CategoryID,
		ParentForumID:                 forum.ParentForumID,
		Kind:                          string(forum.Kind),
		Key:                           string(forum.Key),
		Slug:                          string(forum.Slug),
		Name:                          forum.Name,
		Description:                   forum.Description,
		DisplayOrder:                  forum.DisplayOrder,
		Path:                          forum.Path,
		Depth:                         forum.Depth,
		ExternalURL:                   forum.ExternalURL,
		IconAssetID:                   forum.IconAssetID,
		ThreadVisibilityMode:          string(forum.ThreadVisibilityMode),
		MaxStickyThreads:              forum.MaxStickyThreads,
		DefaultThreadStatus:           string(forum.DefaultThreadStatus),
		AuthorPostEditWindowSeconds:   forum.AuthorPostEditWindowSeconds,
		AuthorPostDeleteWindowSeconds: forum.AuthorPostDeleteWindowSeconds,
		Status:                        string(forum.Status),
		Version:                       forum.Version,
	}
}

// forumFromModel maps persistence forum to domain.
func forumFromModel(model ForumModel) domain.Forum {
	return domain.Forum{
		ID:                            model.ID.ID,
		CategoryID:                    model.CategoryID,
		ParentForumID:                 model.ParentForumID,
		Kind:                          domain.ForumKind(model.Kind),
		Key:                           domain.Key(model.Key),
		Slug:                          domain.Slug(model.Slug),
		Name:                          model.Name,
		Description:                   model.Description,
		DisplayOrder:                  model.DisplayOrder,
		Path:                          model.Path,
		Depth:                         model.Depth,
		ExternalURL:                   model.ExternalURL,
		IconAssetID:                   model.IconAssetID,
		ThreadVisibilityMode:          domain.ThreadVisibilityMode(model.ThreadVisibilityMode),
		MaxStickyThreads:              model.MaxStickyThreads,
		DefaultThreadStatus:           domain.ThreadStatus(model.DefaultThreadStatus),
		AuthorPostEditWindowSeconds:   model.AuthorPostEditWindowSeconds,
		AuthorPostDeleteWindowSeconds: model.AuthorPostDeleteWindowSeconds,
		Status:                        domain.ForumStatus(model.Status),
		Version:                       model.Version,
		CreatedAt:                     model.CreatedAt,
		UpdatedAt:                     model.UpdatedAt,
	}
}

// threadModelFromDomain maps thread to persistence.
func threadModelFromDomain(thread domain.Thread) ThreadModel {
	return ThreadModel{
		ID:                     orm.ID{ID: thread.ID},
		ForumID:                thread.ForumID,
		AuthorUserID:           thread.AuthorUserID,
		OpenerPostID:           thread.OpenerPostID,
		LatestPostID:           thread.LatestPostID,
		LatestPostAuthorUserID: thread.LatestPostAuthorUserID,
		LatestPostAt:           thread.LatestPostAt,
		Title:                  thread.Title,
		Slug:                   string(thread.Slug),
		Status:                 string(thread.Status),
		StickyState:            string(thread.StickyState),
		StickyOrder:            thread.StickyOrder,
		StickyUntil:            thread.StickyUntil,
		LockedReason:           thread.LockedReason,
		ReplyCount:             thread.ReplyCount,
		VisibleReplyCount:      thread.VisibleReplyCount,
		PostCount:              thread.PostCount,
		VisiblePostCount:       thread.VisiblePostCount,
		LikeCount:              thread.LikeCount,
		ViewCount:              thread.ViewCount,
		Version:                thread.Version,
	}
}

// threadFromModel maps persistence thread to domain.
func threadFromModel(model ThreadModel) domain.Thread {
	return domain.Thread{
		ID:                     model.ID.ID,
		ForumID:                model.ForumID,
		AuthorUserID:           model.AuthorUserID,
		OpenerPostID:           model.OpenerPostID,
		LatestPostID:           model.LatestPostID,
		LatestPostAuthorUserID: model.LatestPostAuthorUserID,
		LatestPostAt:           model.LatestPostAt,
		Title:                  model.Title,
		Slug:                   domain.Slug(model.Slug),
		Status:                 domain.ThreadStatus(model.Status),
		StickyState:            domain.StickyState(model.StickyState),
		StickyOrder:            model.StickyOrder,
		StickyUntil:            model.StickyUntil,
		LockedReason:           model.LockedReason,
		ReplyCount:             model.ReplyCount,
		VisibleReplyCount:      model.VisibleReplyCount,
		PostCount:              model.PostCount,
		VisiblePostCount:       model.VisiblePostCount,
		LikeCount:              model.LikeCount,
		ViewCount:              model.ViewCount,
		Version:                model.Version,
		CreatedAt:              model.CreatedAt,
		UpdatedAt:              model.UpdatedAt,
	}
}

// postModelFromDomain maps post to persistence.
func postModelFromDomain(post domain.Post) PostModel {
	return PostModel{
		ID:                  orm.ID{ID: post.ID},
		ThreadID:            post.ThreadID,
		ForumID:             post.ForumID,
		AuthorUserID:        post.AuthorUserID,
		Sequence:            post.Sequence,
		Status:              string(post.Status),
		ContentFormat:       string(post.ContentFormat),
		ContentDocumentJSON: string(post.ContentDocumentJSON),
		ContentText:         post.ContentText,
		ContentChecksum:     post.ContentChecksum,
		EditedAt:            post.EditedAt,
		EditedByUserID:      post.EditedByUserID,
		EditCount:           post.EditCount,
		LikeCount:           post.LikeCount,
		ReplyReferenceCount: post.ReplyReferenceCount,
		Version:             post.Version,
	}
}

// postFromModel maps persistence post to domain.
func postFromModel(model PostModel) domain.Post {
	return domain.Post{
		ID:                  model.ID.ID,
		ThreadID:            model.ThreadID,
		ForumID:             model.ForumID,
		AuthorUserID:        model.AuthorUserID,
		Sequence:            model.Sequence,
		Status:              domain.PostStatus(model.Status),
		ContentFormat:       domain.ContentFormat(model.ContentFormat),
		ContentDocumentJSON: json.RawMessage(model.ContentDocumentJSON),
		ContentText:         model.ContentText,
		ContentChecksum:     model.ContentChecksum,
		EditedAt:            model.EditedAt,
		EditedByUserID:      model.EditedByUserID,
		EditCount:           model.EditCount,
		LikeCount:           model.LikeCount,
		ReplyReferenceCount: model.ReplyReferenceCount,
		Version:             model.Version,
		CreatedAt:           model.CreatedAt,
		UpdatedAt:           model.UpdatedAt,
	}
}

// revisionModelFromDomain maps revision to persistence.
func revisionModelFromDomain(revision domain.PostRevision) PostRevisionModel {
	return PostRevisionModel{
		ID:                          orm.ID{ID: revision.ID},
		PostID:                      revision.PostID,
		EditedByUserID:              revision.EditedByUserID,
		PreviousContentDocumentJSON: string(revision.PreviousContentDocumentJSON),
		PreviousContentText:         revision.PreviousContentText,
		EditReason:                  revision.EditReason,
		CreatedAt:                   revision.CreatedAt,
	}
}

// revisionFromModel maps persistence revision to domain.
func revisionFromModel(model PostRevisionModel) domain.PostRevision {
	return domain.PostRevision{
		ID:                          model.ID.ID,
		PostID:                      model.PostID,
		EditedByUserID:              model.EditedByUserID,
		PreviousContentDocumentJSON: json.RawMessage(model.PreviousContentDocumentJSON),
		PreviousContentText:         model.PreviousContentText,
		EditReason:                  model.EditReason,
		CreatedAt:                   model.CreatedAt,
	}
}

// referenceModelFromDomain maps reference to persistence.
func referenceModelFromDomain(reference domain.PostReference) PostReferenceModel {
	return PostReferenceModel{
		ID:            orm.ID{ID: reference.ID},
		SourcePostID:  reference.SourcePostID,
		TargetPostID:  reference.TargetPostID,
		TargetUserID:  reference.TargetUserID,
		TargetAssetID: reference.TargetAssetID,
		ReferenceType: string(reference.ReferenceType),
		QuoteExcerpt:  reference.QuoteExcerpt,
		LinkURL:       reference.LinkURL,
		CreatedAt:     reference.CreatedAt,
	}
}

// referenceFromModel maps persistence reference to domain.
func referenceFromModel(model PostReferenceModel) domain.PostReference {
	return domain.PostReference{
		ID:            model.ID.ID,
		SourcePostID:  model.SourcePostID,
		TargetPostID:  model.TargetPostID,
		TargetUserID:  model.TargetUserID,
		TargetAssetID: model.TargetAssetID,
		ReferenceType: domain.ReferenceType(model.ReferenceType),
		QuoteExcerpt:  model.QuoteExcerpt,
		LinkURL:       model.LinkURL,
		CreatedAt:     model.CreatedAt,
	}
}
