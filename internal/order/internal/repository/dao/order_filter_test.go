package dao

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func TestBuildOrderFilter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		userID   string
		statuses []int
		want     bson.M
	}{
		{
			name:     "user and status are combined",
			userID:   "alice",
			statuses: []int{1, 2},
			want: bson.M{
				"create_by": "alice",
				"status":    bson.M{"$in": []int{1, 2}},
			},
		},
		{
			name:     "status only",
			statuses: []int{2},
			want:     bson.M{"status": bson.M{"$in": []int{2}}},
		},
		{
			name:   "user only",
			userID: "alice",
			want:   bson.M{"create_by": "alice"},
		},
		{
			name: "no filters",
			want: bson.M{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, buildOrderFilter(tt.userID, tt.statuses))
		})
	}
}
