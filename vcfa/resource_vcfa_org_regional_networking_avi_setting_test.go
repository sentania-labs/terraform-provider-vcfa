//go:build tm || org || ALL || functional

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vcfa

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccVcfaOrgRegionalNetworkingAviSetting(t *testing.T) {
	preTestChecks(t)
	defer postTestChecks(t)
	skipIfNotSysAdmin(t)

	nsxManagerHcl, nsxManagerHclRef := getNsxManagerHcl(t)
	vCenterHcl, vCenterHclRef := getVCenterHcl(t, nsxManagerHclRef)
	regionHcl, regionHclRef := getRegionHcl(t, vCenterHclRef, nsxManagerHclRef)
	ipSpaceHcl, ipSpaceHclRef := getIpSpaceHcl(t, regionHclRef, "1", "1")
	providerGatewayHcl, providerGatewayHclRef := getProviderGatewayHcl(t, regionHclRef, ipSpaceHclRef)
	edgeClusterHcl, _ := getEdgeClusterHcl(t, nsxManagerHclRef, regionHclRef)

	params := StringMap{
		"Testname":          t.Name(),
		"RegionId":          fmt.Sprintf("%s.id", regionHclRef),
		"ProviderGatewayId": fmt.Sprintf("%s.id", providerGatewayHclRef),
		"Tags":              "tm org",
	}
	testParamsNotEmpty(t, params)

	skipBinaryTest := "# skip-binary-test: prerequisite buildup for acceptance tests"
	configText0 := templateFill(vCenterHcl+nsxManagerHcl+skipBinaryTest, params)
	params["FuncName"] = t.Name() + "-step0"

	preRequisites := vCenterHcl + nsxManagerHcl + regionHcl + ipSpaceHcl + providerGatewayHcl + edgeClusterHcl
	configText1 := templateFill(preRequisites+testAccVcfaOrgRegionalNetworkingAviSettingStep1, params)
	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(preRequisites+testAccVcfaOrgRegionalNetworkingAviSettingStep2, params)

	debugPrintf("#[DEBUG] CONFIGURATION step1: %s\n", configText1)
	debugPrintf("#[DEBUG] CONFIGURATION step2: %s\n", configText2)
	if vcfaShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText0,
			},
			{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcfa_org_regional_networking_avi_setting.test", "id"),
					resource.TestCheckResourceAttr("vcfa_org_regional_networking_avi_setting.test", "active", "true"),
					resource.TestCheckResourceAttr("vcfa_org_regional_networking_avi_setting.test", "service_engine_group_mode", "TENANT_MANAGED"),
					resource.TestCheckResourceAttr("vcfa_org_regional_networking_avi_setting.test", "service_engine_quota", "60"),
					resource.TestCheckResourceAttr("vcfa_org_regional_networking_avi_setting.test", "application_limit", "10"),
					resource.TestCheckResourceAttr("vcfa_org_regional_networking_avi_setting.test", "service_engine_group_refs.#", "0"),
					resource.TestCheckResourceAttrPair("vcfa_org_regional_networking_avi_setting.test", "regional_networking_setting_id", "vcfa_org_regional_networking.test", "id"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcfa_org_regional_networking_avi_setting.test", "active", "true"),
					resource.TestCheckResourceAttr("vcfa_org_regional_networking_avi_setting.test", "service_engine_group_mode", "TENANT_MANAGED"),
					resource.TestCheckResourceAttr("vcfa_org_regional_networking_avi_setting.test", "service_engine_quota", "80"),
					resource.TestCheckResourceAttr("vcfa_org_regional_networking_avi_setting.test", "application_limit", "20"),
				),
			},
			{
				ResourceName:      "vcfa_org_regional_networking_avi_setting.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     fmt.Sprintf("%s%s%s", params["Testname"].(string), ImportSeparator, params["Testname"].(string)),
			},
		},
	})
}

const testAccVcfaOrgRegionalNetworkingAviSettingStep1 = testAccVcfaOrgRegionalNetworkingStep1 + `
resource "vcfa_org_regional_networking_avi_setting" "test" {
  regional_networking_setting_id = vcfa_org_regional_networking.test.id
  active                          = true
  service_engine_group_mode       = "TENANT_MANAGED"
  service_engine_quota            = 60
  application_limit               = 10
}
`

const testAccVcfaOrgRegionalNetworkingAviSettingStep2 = testAccVcfaOrgRegionalNetworkingStep1 + `
resource "vcfa_org_regional_networking_avi_setting" "test" {
  regional_networking_setting_id = vcfa_org_regional_networking.test.id
  active                          = true
  service_engine_group_mode       = "TENANT_MANAGED"
  service_engine_quota            = 80
  application_limit               = 20
}
`
