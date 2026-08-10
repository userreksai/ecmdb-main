package dao

import (
	"reflect"
	"testing"

	"github.com/Duke1616/ecmdb/internal/resource/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

func TestBuildResourceSearchFilterSupportsExactMatchingForAllFields(t *testing.T) {
	filter := buildResourceSearchFilter("domain", []string{"platform", "status"}, []domain.SearchCondition{{
		Keyword:   "godaddy",
		MatchType: domain.SearchMatchTypeExact,
	}})
	andConditions := filter["$and"].(bson.A)
	conditions, ok := andConditions[0].(bson.M)["$or"].([]bson.M)
	if !ok || len(conditions) != 2 {
		t.Fatalf("expected one exact condition per field, got %#v", filter)
	}
	if conditions[0]["platform"] != "godaddy" || conditions[1]["status"] != "godaddy" {
		t.Fatalf("unexpected all-field exact conditions: %#v", conditions)
	}
}

func TestBuildResourceSearchFilterSupportsFuzzyMatchingForSelectedField(t *testing.T) {
	filter := buildResourceSearchFilter("domain", []string{"platform", "status"}, []domain.SearchCondition{{
		FieldUID:  "status",
		Keyword:   "正常",
		MatchType: domain.SearchMatchTypeFuzzy,
	}})
	andConditions := filter["$and"].(bson.A)
	conditions, ok := andConditions[0].(bson.M)["$or"].([]bson.M)
	if !ok || len(conditions) != 2 {
		t.Fatalf("expected fuzzy string and converted-value conditions, got %#v", filter)
	}
	fieldCondition, ok := conditions[0]["status"].(bson.M)
	if !ok {
		t.Fatalf("fuzzy condition targets an unexpected field: %#v", conditions[0])
	}
	if _, ok = fieldCondition["$regex"]; !ok {
		t.Fatalf("expected selected-field regex matching, got %#v", fieldCondition)
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

func TestBuildResourceSearchFilterCombinesMixedMatchTypesWithAnd(t *testing.T) {
	filter := buildResourceSearchFilter("domain", []string{"platform", "status"}, []domain.SearchCondition{
		{FieldUID: "platform", Keyword: "godaddy", MatchType: domain.SearchMatchTypeExact},
		{FieldUID: "status", Keyword: "正常", MatchType: domain.SearchMatchTypeFuzzy},
	})

	conditions, ok := filter["$and"].(bson.A)
	if !ok || len(conditions) != 2 {
		t.Fatalf("expected two AND conditions, got %#v", filter)
	}
	exactConditions := conditions[0].(bson.M)["$or"].([]bson.M)
	if exactConditions[0]["platform"] != "godaddy" {
		t.Fatalf("expected exact platform matching, got %#v", conditions[0])
	}
	fuzzyConditions := conditions[1].(bson.M)["$or"].([]bson.M)
	statusCondition, ok := fuzzyConditions[0]["status"].(bson.M)
	if !ok {
		t.Fatalf("expected fuzzy status matching, got %#v", conditions[1])
	}
	if _, ok = statusCondition["$regex"]; !ok {
		t.Fatalf("expected a status regex, got %#v", statusCondition)
	}
}

func TestBuildResourceSearchFilterCombinesCommaSeparatedKeywordsWithOr(t *testing.T) {
	filter := buildResourceSearchFilter("domain", []string{"platform", "status"}, []domain.SearchCondition{
		{FieldUID: "platform", Keyword: "腾讯云国际, 华纳云", MatchType: domain.SearchMatchTypeFuzzy},
		{FieldUID: "status", Keyword: "正常", MatchType: domain.SearchMatchTypeFuzzy},
	})

	conditions, ok := filter["$and"].(bson.A)
	if !ok || len(conditions) != 2 {
		t.Fatalf("expected two AND condition groups, got %#v", filter)
	}
	platformConditions := conditions[0].(bson.M)["$or"].([]bson.M)
	if len(platformConditions) != 4 {
		t.Fatalf("expected comma-separated platform values to use OR, got %#v", conditions[0])
	}
	firstRegex := platformConditions[0]["platform"].(bson.M)["$regex"].(primitive.Regex)
	secondRegex := platformConditions[2]["platform"].(bson.M)["$regex"].(primitive.Regex)
	if firstRegex.Pattern != "腾讯云国际" || secondRegex.Pattern != "华纳云" {
		t.Fatalf("unexpected fuzzy platform values: %#v", platformConditions)
	}
	statusConditions := conditions[1].(bson.M)["$or"].([]bson.M)
	if len(statusConditions) == 0 {
		t.Fatalf("expected the second row to remain an AND condition group, got %#v", conditions[1])
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
