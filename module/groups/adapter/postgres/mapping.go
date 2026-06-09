package postgres

import (
	"github.com/niflaot/gamehub-go/module/groups/domain"
	"github.com/niflaot/gamehub-go/pkg/orm"
)

// groupModelFromDomain maps domain group to persistence.
func groupModelFromDomain(group domain.Group) GroupModel {
	return GroupModel{ID: orm.ID{ID: group.ID}, Key: string(group.Key), Name: group.Name, Description: group.Description, Color: string(group.Color), Weight: group.Weight, Status: string(group.Status), IconAssetID: group.IconAssetID, Version: group.Version}
}

// groupFromModel maps persistence group to domain.
func groupFromModel(model GroupModel) domain.Group {
	return domain.Group{ID: model.ID.ID, Key: domain.Key(model.Key), Name: model.Name, Description: model.Description, Color: domain.Color(model.Color), Weight: model.Weight, Status: domain.GroupStatus(model.Status), IconAssetID: model.IconAssetID, Version: model.Version, CreatedAt: model.CreatedAt, UpdatedAt: model.UpdatedAt}
}

// membershipModelFromDomain maps domain membership to persistence.
func membershipModelFromDomain(membership domain.Membership) MembershipModel {
	return MembershipModel{ID: orm.ID{ID: membership.ID}, GroupID: membership.GroupID, UserID: membership.UserID, Status: string(membership.Status), AssignedByUserID: membership.AssignedByUserID, AssignedReason: membership.AssignedReason, StartsAt: membership.StartsAt, ExpiresAt: membership.ExpiresAt, Version: membership.Version}
}

// membershipFromModel maps persistence membership to domain.
func membershipFromModel(model MembershipModel) domain.Membership {
	return domain.Membership{ID: model.ID.ID, GroupID: model.GroupID, UserID: model.UserID, Status: domain.MembershipStatus(model.Status), AssignedByUserID: model.AssignedByUserID, AssignedReason: model.AssignedReason, StartsAt: model.StartsAt, ExpiresAt: model.ExpiresAt, Version: model.Version, CreatedAt: model.CreatedAt, UpdatedAt: model.UpdatedAt}
}

// tupleModelFromDomain maps domain tuple to persistence.
func tupleModelFromDomain(tuple domain.RelationTuple) RelationTupleModel {
	return RelationTupleModel{ID: orm.ID{ID: tuple.ID}, ObjectType: string(tuple.ObjectType), ObjectID: tuple.ObjectID, Relation: string(tuple.Relation), SubjectType: string(tuple.SubjectType), SubjectID: tuple.SubjectID, SubjectRelation: string(tuple.SubjectRelation), CreatedByUserID: tuple.CreatedByUserID}
}

// tupleFromModel maps persistence tuple to domain.
func tupleFromModel(model RelationTupleModel) domain.RelationTuple {
	return domain.RelationTuple{ID: model.ID.ID, ObjectType: domain.ObjectType(model.ObjectType), ObjectID: model.ObjectID, Relation: domain.Relation(model.Relation), SubjectType: domain.SubjectType(model.SubjectType), SubjectID: model.SubjectID, SubjectRelation: domain.Relation(model.SubjectRelation), CreatedByUserID: model.CreatedByUserID, CreatedAt: model.CreatedAt}
}
