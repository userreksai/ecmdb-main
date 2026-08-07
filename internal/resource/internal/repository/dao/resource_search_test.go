package dao

import (
	"reflect"
	"testing"

	"github.com/Duke1616/ecmdb/internal/resource/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
)

func TestBuildResourceSearchFilterUsesFuzzyMatchingForAllFields(t *testing.T) {
	filter := buildResourceSearchFilter("certificate", []string{"name", "remaining_days"}, []domain.SearchCondition{{Keyword: "2"}})
	andConditions, ok := filter["$and"].(bson.A)
	if !ok || len(andConditions) != 1 {
		t.Fatalf("expected one search condition, got %#v", filter)
	}
	conditions, ok := andConditions[0].(bson.M)["$or"].([]bson.M)
	if !ok || len(conditions) == 0 {
		t.Fatalf("expected all-field fuzzy conditions, got %#v", filter)
	}
	fieldCondition, ok := conditions[0]["name"].(bson.M)
	if !ok {
		t.Fatalf("expected a field condition, got %#v", conditions[0])
	}
	if _, ok = fieldCondition["$regex"]; !ok {
		t.Fatalf("expected regex matching, got %#v", fieldCondition)
	}
}

func TestBuildResourceSearchFilterUsesExactMatchingForSelectedField(t *testing.T) {
	filter := buildResourceSearchFilter("certificate", []string{"remaining_days"}, []domain.SearchCondition{{FieldUID: "remaining_days", Keyword: "2"}})
	andConditions, ok := filter["$and"].(bson.A)
	if !ok || len(andConditions) != 1 {
		t.Fatalf("expected one search condition, got %#v", filter)
	}
	conditions, ok := andConditions[0].(bson.M)["$or"].([]bson.M)
	if !ok || len(conditions) != 3 {
		t.Fatalf("expected string and numeric exact conditions, got %#v", filter)
	}
	for _, condition := range conditions {
		if _, ok = condition["remaining_days"]; !ok {
			t.Fatalf("exact condition targets an unexpected field: %#v", condition)
		}
	}
}

func TestParseNumericComparison(t *testing.T) {
	testCases := []struct {
		keyword  string
		operator string
		value    float64
	}{
		{keyword: ">2", operator: "$gt", value: 2},
		{keyword: "< 2.5", operator: "$lt", value: 2.5},
		{keyword: ">= -3", operator: "$gte", value: -3},
		{keyword: "<=10", operator: "$lte", value: 10},
	}

	for _, testCase := range testCases {
		operator, value, ok := parseNumericComparison(testCase.keyword)
		if !ok || operator != testCase.operator || value != testCase.value {
			t.Fatalf("unexpected comparison parse for %q: %q, %v, %v", testCase.keyword, operator, value, ok)
		}
	}
	if _, _, ok := parseNumericComparison(">not-a-number"); ok {
		t.Fatal("invalid numeric comparison should fall back to exact matching")
	}
}

func TestNumericComparisonConvertsStoredStringsAndNumbers(t *testing.T) {
	filter := buildResourceSearchFilter("certificate", []string{"remaining_days"}, []domain.SearchCondition{{FieldUID: "remaining_days", Keyword: ">2"}})
	searchConditions, ok := filter["$and"].(bson.A)
	if !ok || len(searchConditions) != 1 {
		t.Fatalf("expected one search condition, got %#v", filter)
	}
	expression, ok := searchConditions[0].(bson.M)["$expr"].(bson.M)
	if !ok {
		t.Fatalf("expected a numeric comparison expression, got %#v", filter)
	}
	andConditions, ok := expression["$and"].(bson.A)
	if !ok || len(andConditions) != 2 {
		t.Fatalf("expected conversion guard and comparison, got %#v", expression)
	}
	comparison := andConditions[1].(bson.M)
	operands := comparison["$gt"].(bson.A)
	conversion := operands[0].(bson.M)["$convert"].(bson.M)
	if conversion["to"] != "double" || conversion["input"] != "$remaining_days" {
		t.Fatalf("unexpected numeric conversion: %#v", conversion)
	}
}

func TestBuildResourceSearchFilterCombinesMultipleConditionsWithAnd(t *testing.T) {
	filter := buildResourceSearchFilter("domain", []string{"platform", "status"}, []domain.SearchCondition{
		{FieldUID: "platform", Keyword: "godaddy"},
		{FieldUID: "status", Keyword: "已注册且正常"},
	})

	conditions, ok := filter["$and"].(bson.A)
	if !ok || len(conditions) != 2 {
		t.Fatalf("expected two AND conditions, got %#v", filter)
	}
	for index, field := range []string{"platform", "status"} {
		exactConditions, ok := conditions[index].(bson.M)["$or"].([]bson.M)
		if !ok || len(exactConditions) == 0 {
			t.Fatalf("expected exact matching for %s, got %#v", field, conditions[index])
		}
		if _, ok = exactConditions[0][field]; !ok {
			t.Fatalf("condition %d targets an unexpected field: %#v", index, exactConditions[0])
		}
	}
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
