//go:build unit || ALL

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vcfa

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

func TestGetTmOrgRegionalNetworkingAviSettingType(t *testing.T) {
	tests := []struct {
		name      string
		config    map[string]interface{}
		wantError bool
		wantRefs  int
		wantNil   bool
	}{
		{
			name: "TenantManaged",
			config: map[string]interface{}{
				"active":                    true,
				"service_engine_group_mode": "TENANT_MANAGED",
				"service_engine_quota":      60,
			},
			wantNil: true,
		},
		{
			name: "ProviderManaged",
			config: map[string]interface{}{
				"active":                    true,
				"service_engine_group_mode": "PROVIDER_MANAGED",
				"service_engine_quota":      60,
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
			wantNil: true,
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
			if (got.ServiceEngineGroupRefs == nil) != test.wantNil {
				t.Fatalf("ServiceEngineGroupRefs nil = %v, want %v", got.ServiceEngineGroupRefs == nil, test.wantNil)
			}
			if !got.Active && (got.ServiceEngineGroupMode != "" || got.ServiceEngineQuota != 0 || got.ApplicationLimit != nil) {
				t.Fatal("inactive configuration must use null Avi settings")
			}
		})
	}
}

func TestSetTmOrgRegionalNetworkingAviSettingDataWithNullableFields(t *testing.T) {
	data := schema.TestResourceDataRaw(t, resourceVcfaOrgRegionalNetworkingAviSetting().Schema, map[string]interface{}{
		"regional_networking_setting_id": "urn:vcloud:regionalNetworkingSetting:one",
		"active":                         true,
		"service_engine_group_mode":      "TENANT_MANAGED",
		"service_engine_quota":           60,
	})
	setting := &types.TmRegionalNetworkingAviSetting{
		Active:                 true,
		ServiceEngineGroupMode: "TENANT_MANAGED",
		ServiceEngineQuota:     60,
	}

	if err := setTmOrgRegionalNetworkingAviSettingData(data, setting); err != nil {
		t.Fatalf("setTmOrgRegionalNetworkingAviSettingData() error = %v", err)
	}
	if got := data.Get("application_limit").(int); got != 0 {
		t.Fatalf("application_limit = %d, want 0 for null API value", got)
	}
	if got := data.Get("service_engine_group_refs").(*schema.Set).Len(); got != 0 {
		t.Fatalf("service_engine_group_refs count = %d, want 0", got)
	}
	if got := data.Get("status").(string); got != "" {
		t.Fatalf("status = %q, want empty string for null API value", got)
	}
}

func TestSetTmOrgRegionalNetworkingAviSettingDataPreservesInactiveConfiguration(t *testing.T) {
	data := schema.TestResourceDataRaw(t, resourceVcfaOrgRegionalNetworkingAviSetting().Schema, map[string]interface{}{
		"regional_networking_setting_id": "urn:vcloud:regionalNetworkingSetting:one",
		"active":                         true,
		"service_engine_group_mode":      "PROVIDER_MANAGED",
		"service_engine_quota":           60,
		"service_engine_group_refs":      []interface{}{"urn:vcloud:serviceEngineGroup:one"},
	})

	if err := setTmOrgRegionalNetworkingAviSettingData(data, &types.TmRegionalNetworkingAviSetting{Active: false}); err != nil {
		t.Fatalf("setTmOrgRegionalNetworkingAviSettingData() error = %v", err)
	}
	if got := data.Get("active").(bool); got {
		t.Fatal("active = true, want false")
	}
	if got := data.Get("service_engine_group_mode").(string); got != "PROVIDER_MANAGED" {
		t.Fatalf("service_engine_group_mode = %q, want PROVIDER_MANAGED", got)
	}
	if got := data.Get("service_engine_quota").(int); got != 60 {
		t.Fatalf("service_engine_quota = %d, want 60", got)
	}
	if got := data.Get("service_engine_group_refs").(*schema.Set).Len(); got != 1 {
		t.Fatalf("service_engine_group_refs count = %d, want 1", got)
	}
}
