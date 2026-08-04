//go:build tm || org || ALL || functional

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vcfa

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccVcfaOrgIdpImport(t *testing.T) {
	preTestChecks(t)
	defer postTestChecks(t)
	skipIfNotSysAdmin(t)
	oidcServerUrl := validateAndGetOidcServerUrl(t, testConfig)

	params := StringMap{
		"Testname":          t.Name(),
		"OrgUser":           "tf-idp-admin",
		"OrgPassword":       "long-change-ME1",
		"IdpUser":           "tf-idp-user@example.test",
		"IdpGroup":          "tf-idp-group@example.test",
		"WellKnownEndpoint": oidcServerUrl.String(),
		"Tags":              "tm org",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccVcfaOrgIdpImportPrerequisites, params)
	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccVcfaOrgIdpImportStep1, params)
	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(testAccVcfaOrgIdpImportStep2, params)

	if vcfaShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	multipleFactories := func() map[string]func() (*schema.Provider, error) {
		return map[string]func() (*schema.Provider, error){
			"vcfa": func() (*schema.Provider, error) {
				return testAccProvider, nil
			},
			"vcfatenant": func() (*schema.Provider, error) {
				return testOrgProvider(params["Testname"].(string), params["OrgUser"].(string), params["OrgPassword"].(string)), nil
			},
		}
	}
	defer cachedVCDClients.reset()

	importId := func(principalType, name string) resource.ImportStateIdFunc {
		return func(state *terraform.State) (string, error) {
			orgResource, ok := state.RootModule().Resources["vcfa_org.test"]
			if !ok {
				return "", fmt.Errorf("vcfa_org.test was not found in state")
			}
			return fmt.Sprintf("%s/%s/%s", orgResource.Primary.ID, principalType, name), nil
		}
	}

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProviderFactories: testAccProviders,
				Config:            configText1,
			},
			{
				ProviderFactories: multipleFactories(),
				Config:            configText2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcfa_org_idp_import.user", "principal_type", "user"),
					resource.TestCheckResourceAttr("vcfa_org_idp_import.user", "name", params["IdpUser"].(string)),
					resource.TestCheckResourceAttr("vcfa_org_idp_import.user", "provider_type", "OAUTH"),
					resource.TestCheckResourceAttr("vcfa_org_idp_import.user", "inherit_group_roles", "true"),
					resource.TestCheckResourceAttr("vcfa_org_idp_import.user", "role_ids.#", "1"),
					resource.TestCheckResourceAttr("vcfa_org_idp_import.group", "principal_type", "group"),
					resource.TestCheckResourceAttr("vcfa_org_idp_import.group", "name", params["IdpGroup"].(string)),
					resource.TestCheckResourceAttr("vcfa_org_idp_import.group", "provider_type", "OAUTH"),
					resource.TestCheckResourceAttr("vcfa_org_idp_import.group", "role_ids.#", "1"),
				),
			},
			{
				ProviderFactories: multipleFactories(),
				Config:            configText3,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcfa_org_idp_import.user", "inherit_group_roles", "false"),
					resource.TestCheckResourceAttr("vcfa_org_idp_import.user", "role_ids.#", "2"),
					resource.TestCheckResourceAttr("vcfa_org_idp_import.group", "role_ids.#", "2"),
				),
			},
			{
				ProviderFactories: multipleFactories(),
				ResourceName:      "vcfa_org_idp_import.user",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importId("user", params["IdpUser"].(string)),
			},
			{
				ProviderFactories: multipleFactories(),
				ResourceName:      "vcfa_org_idp_import.group",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importId("group", params["IdpGroup"].(string)),
			},
		},
	})
}

const testAccVcfaOrgIdpImportPrerequisites = `
resource "vcfa_org" "test" {
  name         = "{{.Testname}}"
  display_name = "terraform-test"
  description  = "terraform test"
  is_enabled   = true
}

data "vcfa_role" "org-admin" {
  org_id = vcfa_org.test.id
  name   = "Organization Administrator"
}

data "vcfa_role" "org-user" {
  org_id = vcfa_org.test.id
  name   = "Organization User"
}

resource "vcfa_org_local_user" "admin" {
  org_id   = vcfa_org.test.id
  role_ids = [data.vcfa_role.org-admin.id]
  username = "{{.OrgUser}}"
  password = "{{.OrgPassword}}"
}

resource "vcfa_org_oidc" "test" {
  org_id                 = vcfa_org.test.id
  enabled                = true
  prefer_id_token        = true
  client_id              = "clientId"
  client_secret          = "clientSecret"
  max_clock_skew_seconds = 60
  wellknown_endpoint     = "{{.WellKnownEndpoint}}"
  claims_mapping {
    email     = "email"
    subject   = "sub"
    full_name = "name"
    groups    = "groups"
  }
}
`

const testAccVcfaOrgIdpImportStep1 = testAccVcfaOrgIdpImportPrerequisites + `
resource "vcfa_org_idp_import" "user" {
  provider = vcfatenant

  org_id              = vcfa_org.test.id
  principal_type      = "user"
  name                = "{{.IdpUser}}"
  provider_type       = "OAUTH"
  role_ids            = [data.vcfa_role.org-user.id]
  inherit_group_roles = true

  depends_on = [vcfa_org_local_user.admin, vcfa_org_oidc.test]
}

resource "vcfa_org_idp_import" "group" {
  provider = vcfatenant

  org_id         = vcfa_org.test.id
  principal_type = "group"
  name           = "{{.IdpGroup}}"
  provider_type  = "OAUTH"
  role_ids       = [data.vcfa_role.org-user.id]

  depends_on = [vcfa_org_local_user.admin, vcfa_org_oidc.test]
}
`

const testAccVcfaOrgIdpImportStep2 = testAccVcfaOrgIdpImportPrerequisites + `
resource "vcfa_org_idp_import" "user" {
  provider = vcfatenant

  org_id              = vcfa_org.test.id
  principal_type      = "user"
  name                = "{{.IdpUser}}"
  provider_type       = "OAUTH"
  role_ids            = [data.vcfa_role.org-user.id, data.vcfa_role.org-admin.id]
  inherit_group_roles = false

  depends_on = [vcfa_org_local_user.admin, vcfa_org_oidc.test]
}

resource "vcfa_org_idp_import" "group" {
  provider = vcfatenant

  org_id         = vcfa_org.test.id
  principal_type = "group"
  name           = "{{.IdpGroup}}"
  provider_type  = "OAUTH"
  role_ids       = [data.vcfa_role.org-user.id, data.vcfa_role.org-admin.id]

  depends_on = [vcfa_org_local_user.admin, vcfa_org_oidc.test]
}
`
