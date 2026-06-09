package postgres

import (
	"github.com/niflaot/gamehub-go/module/forums/domain"
	"github.com/niflaot/gamehub-go/pkg/orm"
)

// categoryModelFromDomain maps category to persistence.
func categoryModelFromDomain(category domain.ForumCategory) CategoryModel {
	return CategoryModel{ID: orm.ID{ID: category.ID}, Key: string(category.Key), Name: category.Name, Description: category.Description, DisplayOrder: category.DisplayOrder, Status: string(category.Status), Version: category.Version}
}

// categoryFromModel maps persistence category to domain.
func categoryFromModel(model CategoryModel) domain.ForumCategory {
	return domain.ForumCategory{ID: model.ID.ID, Key: domain.Key(model.Key), Name: model.Name, Description: model.Description, DisplayOrder: model.DisplayOrder, Status: domain.CategoryStatus(model.Status), Version: model.Version, CreatedAt: model.CreatedAt, UpdatedAt: model.UpdatedAt}
}

// forumModelFromDomain maps forum to persistence.
func forumModelFromDomain(forum domain.Forum) ForumModel {
	return ForumModel{ID: orm.ID{ID: forum.ID}, CategoryID: forum.CategoryID, ParentForumID: forum.ParentForumID, Kind: string(forum.Kind), Key: string(forum.Key), Slug: string(forum.Slug), Name: forum.Name, Description: forum.Description, DisplayOrder: forum.DisplayOrder, Path: forum.Path, Depth: forum.Depth, ExternalURL: forum.ExternalURL, IconAssetID: forum.IconAssetID, ThreadVisibilityMode: string(forum.ThreadVisibilityMode), MaxStickyThreads: forum.MaxStickyThreads, DefaultThreadStatus: string(forum.DefaultThreadStatus), Status: string(forum.Status), Version: forum.Version}
}

// forumFromModel maps persistence forum to domain.
func forumFromModel(model ForumModel) domain.Forum {
	return domain.Forum{ID: model.ID.ID, CategoryID: model.CategoryID, ParentForumID: model.ParentForumID, Kind: domain.ForumKind(model.Kind), Key: domain.Key(model.Key), Slug: domain.Slug(model.Slug), Name: model.Name, Description: model.Description, DisplayOrder: model.DisplayOrder, Path: model.Path, Depth: model.Depth, ExternalURL: model.ExternalURL, IconAssetID: model.IconAssetID, ThreadVisibilityMode: domain.ThreadVisibilityMode(model.ThreadVisibilityMode), MaxStickyThreads: model.MaxStickyThreads, DefaultThreadStatus: domain.ThreadStatus(model.DefaultThreadStatus), Status: domain.ForumStatus(model.Status), Version: model.Version, CreatedAt: model.CreatedAt, UpdatedAt: model.UpdatedAt}
}

// statsFromModel maps persistence stats to domain.
func statsFromModel(model StatsModel) domain.ForumStats {
	return domain.ForumStats{ForumID: model.ForumID, ThreadCount: model.ThreadCount, VisibleThreadCount: model.VisibleThreadCount, PostCount: model.PostCount, VisiblePostCount: model.VisiblePostCount, LatestThreadID: model.LatestThreadID, LatestPostID: model.LatestPostID, LatestPostAuthorUserID: model.LatestPostAuthorUserID, LatestPostAt: model.LatestPostAt, UpdatedAt: model.UpdatedAt}
}
