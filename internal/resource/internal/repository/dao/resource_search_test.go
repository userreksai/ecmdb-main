package dao

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestBuildExactSearchConditionsExcludesResourceIDForFieldSearch(t *testing.T) {
	conditions := buildExactSearchConditions([]string{"remaining_days"}, "2", false)

	for _, condition := range conditions {
		if _, ok := condition["id"]; ok {
			t.Fatal("field-specific search must not match the resource ID")
		}
	}
}

func TestBuildExactSearchConditionsIncludesResourceIDForAllFieldSearch(t *testing.T) {
	conditions := buildExactSearchConditions([]string{"remaining_days"}, "2", true)

	for _, condition := range conditions {
		if value, ok := condition["id"]; ok && value == int64(2) {
			return
		}
	}
	t.Fatal("all-field search should continue to match the resource ID")
}

func TestBuildEmptySearchConditionsCoversEmptyNullMissingAndEmptyArray(t *testing.T) {
	conditions := buildEmptySearchConditions([]string{"certificate_expires_at"})

	if len(conditions) != 3 {
		t.Fatalf("expected 3 empty-value conditions, got %d", len(conditions))
	}
	expectedValues := []interface{}{"", nil, bson.M{"$size": 0}}
	for index, condition := range conditions {
		value, ok := condition["certificate_expires_at"]
		if !ok || !reflect.DeepEqual(value, expectedValues[index]) {
			t.Fatalf("unexpected empty-value condition at index %d: %#v", index, condition)
		}
	}
}
