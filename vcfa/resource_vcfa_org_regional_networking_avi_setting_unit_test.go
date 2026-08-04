//go:build unit || ALL

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vcfa

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestGetTmOrgRegionalNetworkingAviSettingType(t *testing.T) {
	tests := []struct {
		name      string
		config    map[string]interface{}
		wantError bool
		wantRefs  int
	}{
		{
			name: "TenantManaged",
			config: map[string]interface{}{
				"active":                    true,
				"service_engine_group_mode": "TENANT_MANAGED",
				"service_engine_quota":      60,
				"application_limit":         10,
			},
		},
		{
			name: "ProviderManaged",
			config: map[string]interface{}{
				"active":                    true,
				"service_engine_group_mode": "PROVIDER_MANAGED",
				"service_engine_quota":      60,
				"application_limit":         10,
				"service_engine_group_refs": []interface{}{"urn:vcloud:serviceEngineGroup:one"},
			},
			wantRefs: 1,
		},
		{
			name: "ProviderManagedRequiresReference",
			config: map[string]interface{}{
				"active":                    true,
				"service_engine_group_mode": "PROVIDER_MANAGED",
				"service_engine_quota":      60,
			},
			wantError: true,
		},
		{
			name: "TenantManagedRejectsReference",
			config: map[string]interface{}{
				"active":                    true,
				"service_engine_group_mode": "TENANT_MANAGED",
				"service_engine_quota":      60,
				"service_engine_group_refs": []interface{}{"urn:vcloud:serviceEngineGroup:one"},
			},
			wantError: true,
		},
		{
			name: "InactiveUsesNullSettings",
			config: map[string]interface{}{
				"active":                    false,
				"service_engine_group_mode": "TENANT_MANAGED",
				"service_engine_quota":      60,
			},
		},
	}

	resourceSchema := resourceVcfaOrgRegionalNetworkingAviSetting().Schema
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data := schema.TestResourceDataRaw(t, resourceSchema, test.config)
			got, err := getTmOrgRegionalNetworkingAviSettingType(data)
			if (err != nil) != test.wantError {
				t.Fatalf("getTmOrgRegionalNetworkingAviSettingType() error = %v, wantError %v", err, test.wantError)
			}
			if test.wantError {
				return
			}
			if got.Active != test.config["active"].(bool) {
				t.Fatalf("Active = %v, want %v", got.Active, test.config["active"])
			}
			if len(got.ServiceEngineGroupRefs) != test.wantRefs {
				t.Fatalf("ServiceEngineGroupRefs count = %d, want %d", len(got.ServiceEngineGroupRefs), test.wantRefs)
			}
			if !got.Active && (got.ServiceEngineGroupMode != nil || got.ServiceEngineQuota != nil || got.ApplicationLimit != nil) {
				t.Fatal("inactive configuration must use null Avi settings")
			}
		})
	}
}
