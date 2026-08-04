//go:build unit || ALL

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vcfa

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

func TestResourceVcfaOrgIdpImportSchema(t *testing.T) {
	resource := resourceVcfaOrgIdpImport()
	if err := resource.InternalValidate(nil, true); err != nil {
		t.Fatalf("resourceVcfaOrgIdpImport schema validation failed: %v", err)
	}

	for _, field := range []string{"org_id", "principal_type", "name", "provider_type"} {
		if !resource.Schema[field].ForceNew {
			t.Fatalf("schema field %q must force replacement", field)
		}
	}
	if resource.Schema["role_ids"].Type != schema.TypeSet {
		t.Fatal("role_ids must use TypeSet")
	}
}

func TestSetOrgIdpImportUserData(t *testing.T) {
	data := schema.TestResourceDataRaw(t, resourceVcfaOrgIdpImport().Schema, map[string]interface{}{
		"principal_type": "user",
	})
	user := &govcd.OpenApiUser{User: &types.OpenApiUser{
		ID:                "urn:vcloud:user:one",
		Username:          "user@example.test",
		ProviderType:      "OAUTH",
		OrgEntityRef:      &types.OpenApiReference{ID: "urn:vcloud:org:one"},
		RoleEntityRefs:    []types.OpenApiReference{{ID: "urn:vcloud:role:one"}},
		InheritGroupRoles: true,
	}}

	if err := setOrgIdpImportUserData(data, user); err != nil {
		t.Fatalf("setOrgIdpImportUserData() error = %v", err)
	}
	if data.Id() != user.User.ID {
		t.Fatalf("ID = %q, want %q", data.Id(), user.User.ID)
	}
	if got := data.Get("name").(string); got != user.User.Username {
		t.Fatalf("name = %q, want %q", got, user.User.Username)
	}
	if got := data.Get("provider_type").(string); got != user.User.ProviderType {
		t.Fatalf("provider_type = %q, want %q", got, user.User.ProviderType)
	}
	if got := data.Get("inherit_group_roles").(bool); !got {
		t.Fatal("inherit_group_roles = false, want true")
	}
	if got := data.Get("role_ids").(*schema.Set).Len(); got != 1 {
		t.Fatalf("role_ids count = %d, want 1", got)
	}
}

func TestSetOrgIdpImportGroupData(t *testing.T) {
	data := schema.TestResourceDataRaw(t, resourceVcfaOrgIdpImport().Schema, map[string]interface{}{
		"principal_type": "group",
	})
	group := &govcd.OpenApiGroup{Group: &types.OpenApiGroup{
		ID:             "urn:vcloud:group:one",
		Name:           "group@example.test",
		ProviderType:   "OAUTH",
		OrgEntityRef:   &types.OpenApiReference{ID: "urn:vcloud:org:one"},
		RoleEntityRefs: []types.OpenApiReference{{ID: "urn:vcloud:role:one"}},
	}}

	if err := setOrgIdpImportGroupData(data, group); err != nil {
		t.Fatalf("setOrgIdpImportGroupData() error = %v", err)
	}
	if data.Id() != group.Group.ID {
		t.Fatalf("ID = %q, want %q", data.Id(), group.Group.ID)
	}
	if got := data.Get("name").(string); got != group.Group.Name {
		t.Fatalf("name = %q, want %q", got, group.Group.Name)
	}
	if got := data.Get("provider_type").(string); got != group.Group.ProviderType {
		t.Fatalf("provider_type = %q, want %q", got, group.Group.ProviderType)
	}
	if got := data.Get("role_ids").(*schema.Set).Len(); got != 1 {
		t.Fatalf("role_ids count = %d, want 1", got)
	}
}
