package dao

import (
	"testing"

	"github.com/Duke1616/ecmdb/internal/task/domain"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func TestSuccessTasksByUtimeFilterUsesCompletedBeforeCutoff(t *testing.T) {
	assert.Equal(t, bson.M{
		"status":      bson.M{"$eq": domain.SUCCESS},
		"utime":       bson.M{"$lte": int64(123)},
		"mark_passed": bson.M{"$eq": false},
	}, successTasksByUtimeFilter(123))
}
