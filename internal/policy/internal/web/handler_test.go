package web

import (
	"reflect"
	"testing"
)

func TestPolicyResourceCandidates(t *testing.T) {
	testCases := []struct {
		name     string
		service  string
		resource string
		want     []string
	}{
		{
			name:     "explicit task resource keeps legacy alias",
			service:  "task",
			resource: "TASK",
			want:     []string{"TASK", "task"},
		},
		{
			name:    "task service supports synced and legacy resources",
			service: "task",
			want:    []string{"task", "TASK"},
		},
		{
			name:    "task module service supports module and base aliases",
			service: "task:manager",
			want:    []string{"task:manager", "task", "TASK"},
		},
		{
			name:    "unknown service falls back to upper alias",
			service: "custom",
			want:    []string{"custom", "CUSTOM"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := policyResourceCandidates(tc.service, tc.resource)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("policyResourceCandidates(%q, %q) = %#v, want %#v", tc.service, tc.resource, got, tc.want)
			}
		})
	}
}

func TestPolicyPathCandidates(t *testing.T) {
	testCases := []struct {
		name string
		path string
		want []string
	}{
		{
			name: "internal task path remains unchanged",
			path: "/api/manager/list",
			want: []string{"/api/manager/list"},
		},
		{
			name: "nginx task prefix falls back to etask internal path",
			path: "/api/task/manager/list",
			want: []string{"/api/task/manager/list", "/api/manager/list"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := policyPathCandidates(tc.path)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("policyPathCandidates(%q) = %#v, want %#v", tc.path, got, tc.want)
			}
		})
	}
}
