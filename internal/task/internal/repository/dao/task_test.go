package dao

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/userreksai/ecmdb-main/internal/task/domain"
	"go.mongodb.org/mongo-driver/bson"
)

func TestSuccessTasksByUtimeFilterUsesCompletedBeforeCutoff(t *testing.T) {
	assert.Equal(t, bson.M{
		"status":      bson.M{"$eq": domain.SUCCESS},
		"utime":       bson.M{"$lte": int64(123)},
		"mark_passed": bson.M{"$eq": false},
	}, successTasksByUtimeFilter(123))
}
