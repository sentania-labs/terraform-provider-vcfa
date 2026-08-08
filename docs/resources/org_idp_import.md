---
page_title: "VMware Cloud Foundation Automation: vcfa_org_idp_import"
subcategory: ""
description: |-
  Provides a resource to import an identity provider user or group into a VMware Cloud Foundation Automation Organization.
---

# vcfa_org_idp_import

Provides a resource to import an LDAP or OIDC user or group into an [Organization][vcfa_org] and assign roles without using the portal.

_Used by: **Provider**_

The resource uses the system-scoped provider with Organization tenant context. The target Organization must have an OIDC or LDAP identity
provider configured before principals can be imported. VCFA may otherwise return a misleading `GROUP_USER_MANAGE` authorization error.

The `name` must match the identity provider claim byte for byte, including case. Changing only the case replaces the imported principal.

## Example Usage

```hcl
data "vcfa_role" "org-admin" {
  org_id = vcfa_org.demo.id
  name   = "Organization Administrator"
}

resource "vcfa_org_idp_import" "labadmins" {
  org_id         = vcfa_org.demo.id
  principal_type = "group"
  name           = "labadmins@example.test"
  provider_type  = "OAUTH"
  role_ids       = [data.vcfa_role.org-admin.id]
}
```

An imported user can inherit roles from its imported groups:

```hcl
resource "vcfa_org_idp_import" "operator" {
  org_id              = vcfa_org.demo.id
  principal_type      = "user"
  name                = "operator@example.test"
  provider_type       = "OAUTH"
  role_ids            = [data.vcfa_role.org-admin.id]
  inherit_group_roles = true
}
```

## Argument Reference

The following arguments are supported:

- `org_id` - (Required) The ID of the [Organization][vcfa_org] receiving the imported principal.
- `principal_type` - (Required) Principal type. Supported values are `group` and `user`. Changing this value replaces the resource.
- `name` - (Required) Principal name as supplied by the identity provider. The value is case sensitive, and changing it replaces the resource.
- `provider_type` - (Required) Identity provider type. Supported values are `LDAP` and `OAUTH`. VCFA uses `OAUTH` for OIDC providers.
- `role_ids` - (Required) A set of [Role][vcfa_role] IDs assigned directly to the principal.
- `inherit_group_roles` - (Optional) Whether an imported user inherits roles from its groups. Defaults to `false` and is not sent for groups.

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the
state. It does not generate configuration. However, an experimental feature in Terraform 1.5+ allows
also code generation. See [Importing resources][importing-resources] for more information.

An existing imported principal can be [imported][docs-import] with its Organization ID, principal type, and case-sensitive name:

```shell
terraform import vcfa_org_idp_import.imported 'urn:vcloud:org:00000000-0000-0000-0000-000000000000.group.labadmins@example.test'
```

The import syntax is `<org_id><separator><principal_type><separator><name>`. The separator defaults to `.` and follows the provider's `import_separator` setting or `VCFA_IMPORT_SEPARATOR` environment variable.

[docs-import]: https://www.terraform.io/docs/import
[importing-resources]: /providers/vmware/vcfa/latest/docs/guides/importing_resources
[vcfa_org]: /providers/vmware/vcfa/latest/docs/resources/org
[vcfa_role]: /providers/vmware/vcfa/latest/docs/data-sources/role
