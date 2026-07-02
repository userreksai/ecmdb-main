package web

import "testing"

func TestPolicyResource(t *testing.T) {
	testCases := []struct {
		name     string
		service  string
		resource string
		want     string
	}{
		{
			name:     "explicit resource wins",
			service:  "task",
			resource: "TASK",
			want:     "TASK",
		},
		{
			name:    "task service maps to task resource",
			service: "task",
			want:    "TASK",
		},
		{
			name:    "task module service maps to task resource",
			service: "task:manager",
			want:    "TASK",
		},
		{
			name:    "unknown service falls back to uppercase",
			service: "custom",
			want:    "CUSTOM",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := policyResource(tc.service, tc.resource)
			if got != tc.want {
				t.Fatalf("policyResource(%q, %q) = %q, want %q", tc.service, tc.resource, got, tc.want)
			}
		})
	}
}
