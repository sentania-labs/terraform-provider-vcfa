---
page_title: "VMware Cloud Foundation Automation: vcfa_org_regional_networking_avi_setting"
subcategory: ""
description: |-
  Provides a resource to manage Organization Regional Networking Avi settings in VMware Cloud Foundation Automation.
---

# vcfa_org_regional_networking_avi_setting

Provides a resource to manage Avi load balancer delegation for an [Organization Regional Networking][vcfa_org_regional_networking]
configuration in VMware Cloud Foundation Automation.

_Used by: **Provider**_

Deleting this resource disables Avi integration by updating the setting with `active = false`.

## Example Usage

```hcl
resource "vcfa_org_regional_networking_avi_setting" "demo" {
  regional_networking_setting_id = vcfa_org_regional_networking.demo.id
  active                          = true
  service_engine_group_mode       = "TENANT_MANAGED"
  service_engine_quota            = 60
}
```

Provider-managed service engine groups can be selected explicitly:

```hcl
resource "vcfa_org_regional_networking_avi_setting" "demo" {
  regional_networking_setting_id = vcfa_org_regional_networking.demo.id
  active                          = true
  service_engine_group_mode       = "PROVIDER_MANAGED"
  service_engine_quota            = 60
  service_engine_group_refs       = [
    "urn:vcloud:serviceEngineGroup:00000000-0000-0000-0000-000000000000",
  ]
}
```

## Argument Reference

The following arguments are supported:

- `regional_networking_setting_id` - (Required) The ID of the [Organization Regional Networking][vcfa_org_regional_networking] configuration.
- `active` - (Required) Whether Avi integration is enabled.
- `service_engine_group_mode` - (Required) Service engine group management mode. Supported values are `TENANT_MANAGED` and `PROVIDER_MANAGED`.
- `service_engine_quota` - (Required) Service engine quota assigned to the Organization.
- `service_engine_group_refs` - (Optional) A set of service engine group IDs. At least one ID is required when `service_engine_group_mode` is `PROVIDER_MANAGED`, and the set must be empty for `TENANT_MANAGED`.
## Attribute Reference

- `application_limit` - Application limit reported by VMware Cloud Foundation Automation. A null API value is stored as `0`.
- `status` - Realization status returned by VMware Cloud Foundation Automation.

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the
state. It does not generate configuration. However, an experimental feature in Terraform 1.5+ allows
also code generation. See [Importing resources][importing-resources] for more information.

An existing Organization Regional Networking Avi setting can be [imported][docs-import] using the
Organization name and Regional Networking configuration name:

```shell
terraform import vcfa_org_regional_networking_avi_setting.imported my-org-name.my-regional-configuration-name
```

_NOTE_: The default separator `.` can be changed using the provider's `import_separator` argument or environment variable `VCFA_IMPORT_SEPARATOR`.

[docs-import]: https://www.terraform.io/docs/import
[importing-resources]: /providers/vmware/vcfa/latest/docs/guides/importing_resources
[vcfa_org_regional_networking]: /providers/vmware/vcfa/latest/docs/resources/org_regional_networking
