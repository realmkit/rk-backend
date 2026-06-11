package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/groups/domain"
	"github.com/niflaot/gamehub-go/module/groups/port"
	"github.com/niflaot/gamehub-go/pkg/orm"
	"github.com/niflaot/gamehub-go/pkg/pagination"
	"gorm.io/gorm"
)

// TupleRepository stores relation tuples in PostgreSQL.
type TupleRepository struct {
	store orm.Store
}

// NewTupleRepository creates a tuple repository.
func NewTupleRepository(store orm.Store) TupleRepository {
	return TupleRepository{store: store}
}

// Create stores a tuple.
func (repository TupleRepository) Create(ctx context.Context, tuple domain.RelationTuple) (domain.RelationTuple, error) {
	if existing, err := repository.findEquivalent(ctx, tuple); err == nil {
		return existing, port.ErrConflict
	} else if !errors.Is(err, port.ErrNotFound) {
		return domain.RelationTuple{}, err
	}
	model := tupleModelFromDomain(tuple)
	if err := repository.store.DB(ctx).Create(&model).Error; err != nil {
		return domain.RelationTuple{}, port.ErrConflict
	}
	return tupleFromModel(model), nil
}

// FindByID returns one tuple.
func (repository TupleRepository) FindByID(ctx context.Context, id uuid.UUID) (domain.RelationTuple, error) {
	var model RelationTupleModel
	if err := repository.store.DB(ctx).First(&model, "id = ?", id).Error; err != nil {
		return domain.RelationTuple{}, mapError(err)
	}
	return tupleFromModel(model), nil
}

// List returns matching tuples.
func (repository TupleRepository) List(
	ctx context.Context,
	filter port.TupleFilter,
	page pagination.Page,
) (pagination.Result[domain.RelationTuple], error) {
	query := applyTupleFilter(repository.store.DB(ctx).Model(&RelationTupleModel{}), filter).Order("created_at asc").Limit(page.Limit + 1)
	var models []RelationTupleModel
	if err := query.Find(&models).Error; err != nil {
		return pagination.Result[domain.RelationTuple]{}, err
	}
	return tuplePage(models, page.Limit), nil
}

// Delete soft deletes one tuple.
func (repository TupleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := repository.store.DB(ctx).Delete(&RelationTupleModel{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return port.ErrNotFound
	}
	return nil
}

// findEquivalent returns a matching active tuple.
func (repository TupleRepository) findEquivalent(ctx context.Context, tuple domain.RelationTuple) (domain.RelationTuple, error) {
	var model RelationTupleModel
	err := repository.store.DB(ctx).
		Where("object_type = ?", tuple.ObjectType).
		Where("object_id = ?", tuple.ObjectID).
		Where("relation = ?", tuple.Relation).
		Where("subject_type = ?", tuple.SubjectType).
		Where("subject_id = ?", tuple.SubjectID).
		Where("subject_relation = ?", tuple.SubjectRelation).
		First(&model).
		Error
	if err != nil {
		return domain.RelationTuple{}, mapError(err)
	}
	return tupleFromModel(model), nil
}

// applyTupleFilter applies tuple filters.
func applyTupleFilter(query *gorm.DB, filter port.TupleFilter) *gorm.DB {
	if filter.ObjectType != "" {
		query = query.Where("object_type = ?", filter.ObjectType)
	}
	if filter.ObjectID != uuid.Nil {
		query = query.Where("object_id = ?", filter.ObjectID)
	}
	if filter.Relation != "" {
		query = query.Where("relation = ?", filter.Relation)
	}
	if filter.SubjectType != "" {
		query = query.Where("subject_type = ?", filter.SubjectType)
	}
	if filter.SubjectID != uuid.Nil {
		query = query.Where("subject_id = ?", filter.SubjectID)
	}
	return query
}

// tuplePage maps tuple models into a page.
func tuplePage(models []RelationTupleModel, limit int) pagination.Result[domain.RelationTuple] {
	next := ""
	if len(models) > limit {
		next = models[limit-1].ID.ID.String()
		models = models[:limit]
	}
	items := make([]domain.RelationTuple, 0, len(models))
	for _, model := range models {
		items = append(items, tupleFromModel(model))
	}
	return pagination.Result[domain.RelationTuple]{Items: items, NextCursor: next}
}
