package application

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/metadata/domain"
	"github.com/realmkit/rk-backend/module/metadata/port"
)

// benchmarkOwnerMetadata stores the metadata view benchmark result.
var benchmarkOwnerMetadata port.OwnerMetadataView

// BenchmarkOwnerMetadataView measures owner metadata view assembly for definition-heavy owners.
func BenchmarkOwnerMetadataView(b *testing.B) {
	owner := port.OwnerRef{Type: domain.OwnerUser, ID: uuid.New()}
	definitions := make([]domain.MetafieldDefinition, 64)
	values := make([]domain.MetafieldValue, 32)
	for index := range definitions {
		definitions[index] = domain.MetafieldDefinition{
			ID:        uuid.New(),
			OwnerType: domain.OwnerUser,
			Key:       domain.Key("field_" + strconv.Itoa(index)),
			Name:      "Field " + strconv.Itoa(index),
			ValueType: domain.ValueSingleLineText,
			Active:    true,
			Version:   1,
		}
		if index < len(values) {
			values[index] = domain.MetafieldValue{
				ID:           uuid.New(),
				DefinitionID: definitions[index].ID,
				OwnerType:    domain.OwnerUser,
				OwnerID:      owner.ID,
				Value:        json.RawMessage(`"value"`),
				Version:      1,
			}
		}
	}

	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		benchmarkOwnerMetadata = ownerMetadataView(owner, definitions, values, true)
	}
}
