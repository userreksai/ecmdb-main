package service

import (
	"context"
	"testing"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
)

const testRBACModel = `[request_definition]
r = sub, obj, act, res

[policy_definition]
p = sub, obj, act, res, eft

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow)) && !some(where (p.eft == deny))

[matchers]
m = (g(r.sub, p.sub) && keyMatch2(r.obj, p.obj) && r.act == p.act && r.res == p.res) || r.sub == "root"`

func newTestService(t *testing.T) *service {
	t.Helper()

	m, err := model.NewModelFromString(testRBACModel)
	if err != nil {
		t.Fatalf("new model: %v", err)
	}

	enforcer, err := casbin.NewSyncedEnforcer(m)
	if err != nil {
		t.Fatalf("new enforcer: %v", err)
	}

	return &service{enforcer: enforcer}
}

func TestAuthorizeAdminRoleBypassesEndpointPolicies(t *testing.T) {
	svc := newTestService(t)
	if _, err := svc.enforcer.AddGroupingPolicy("1", adminRoleCode); err != nil {
		t.Fatalf("add admin role: %v", err)
	}

	result, err := svc.Authorize(context.Background(), "1", "/api/manager/list", "POST", "TASK")
	if err != nil {
		t.Fatalf("authorize: %v", err)
	}
	if !result.Allowed {
		t.Fatalf("admin role should bypass endpoint policies, got denied: %#v", result)
	}
}

func TestAuthorizeNormalRoleStillUsesEndpointPolicies(t *testing.T) {
	svc := newTestService(t)
	if _, err := svc.enforcer.AddGroupingPolicy("2", "viewer"); err != nil {
		t.Fatalf("add viewer role: %v", err)
	}
	if _, err := svc.enforcer.AddPolicy("viewer", "/api/manager/list", "POST", "TASK", "allow"); err != nil {
		t.Fatalf("add viewer policy: %v", err)
	}

	allowed, err := svc.Authorize(context.Background(), "2", "/api/manager/list", "POST", "TASK")
	if err != nil {
		t.Fatalf("authorize allowed path: %v", err)
	}
	if !allowed.Allowed {
		t.Fatalf("viewer should be allowed by matching policy, got denied: %#v", allowed)
	}

	denied, err := svc.Authorize(context.Background(), "2", "/api/manager/executions", "POST", "TASK")
	if err != nil {
		t.Fatalf("authorize denied path: %v", err)
	}
	if denied.Allowed {
		t.Fatalf("viewer should still be denied without matching policy: %#v", denied)
	}
}
