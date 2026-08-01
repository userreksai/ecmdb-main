package service

import (
	"reflect"
	"testing"

	"github.com/Duke1616/ecmdb/internal/role/internal/domain"
)

func TestFilterAccessibleModelUIDs(t *testing.T) {
	testCases := []struct {
		name      string
		roles     []domain.Role
		modelUIDs []string
		want      []string
	}{
		{
			name:      "ordinary role without grants cannot access models",
			roles:     []domain.Role{{Code: "viewer"}},
			modelUIDs: []string{"host", "database"},
			want:      []string{},
		},
		{
			name:      "single role can access explicitly granted models",
			roles:     []domain.Role{{Code: "viewer", AllowedModelUIDs: []string{"host", "network"}}},
			modelUIDs: []string{"host", "database", "network"},
			want:      []string{"host", "network"},
		},
		{
			name: "multiple roles combine access with union semantics",
			roles: []domain.Role{
				{Code: "role-a", AllowedModelUIDs: []string{"host"}},
				{Code: "role-b", AllowedModelUIDs: []string{"database"}},
			},
			modelUIDs: []string{"host", "database", "network"},
			want:      []string{"host", "database"},
		},
		{
			name:      "admin role can access every current and future model",
			roles:     []domain.Role{{Code: domain.AdminRole}},
			modelUIDs: []string{"host", "database", "new-model"},
			want:      []string{"host", "database", "new-model"},
		},
		{
			name:      "new model stays hidden until explicitly granted",
			roles:     []domain.Role{{Code: "viewer", AllowedModelUIDs: []string{"host"}}},
			modelUIDs: []string{"host", "new-model"},
			want:      []string{"host"},
		},
		{
			name:      "empty roles grant no model access",
			roles:     nil,
			modelUIDs: []string{"host"},
			want:      []string{},
		},
		{
			name:      "candidate model identifiers are normalized and deduplicated",
			roles:     []domain.Role{{Code: "viewer", AllowedModelUIDs: []string{"host", "database"}}},
			modelUIDs: []string{" host ", "host", "", "database"},
			want:      []string{"host", "database"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := filterAccessibleModelUIDs(tc.roles, tc.modelUIDs)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("filterAccessibleModelUIDs() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestNormalizeModelUIDs(t *testing.T) {
	got := normalizeModelUIDs([]string{" host ", "host", "", "database"})
	want := []string{"host", "database"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("normalizeModelUIDs() = %v, want %v", got, want)
	}
}
