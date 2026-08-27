// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vcfa

import (
	"encoding/json"
	"testing"
)

func TestTmOrgNetworkingSettingsResetPayload(t *testing.T) {
	settings := tmOrgNetworkingSettingsReset{}
	if settings.NetworkingTenancyEnabled {
		t.Fatal("networking tenancy reset = true, want false")
	}
	if settings.OrgNameForLogs != nil {
		t.Fatalf("org name for logs reset = %q, want null", *settings.OrgNameForLogs)
	}

	payload, err := json.Marshal(settings)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	want := `{"networkingTenancyEnabled":false,"orgNameForLogs":null}`
	if string(payload) != want {
		t.Fatalf("reset payload = %s, want %s", payload, want)
	}
}
