// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vcfa

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelOrgIdpImport = "Organization IdP Principal"

func resourceVcfaOrgIdpImport() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcfaOrgIdpImportCreate,
		ReadContext:   resourceVcfaOrgIdpImportRead,
		UpdateContext: resourceVcfaOrgIdpImportUpdate,
		DeleteContext: resourceVcfaOrgIdpImportDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcfaOrgIdpImportImport,
		},

		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: fmt.Sprintf("Parent %s ID for %s", labelVcfaOrg, labelOrgIdpImport),
			},
			"principal_type": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				Description:      fmt.Sprintf("Principal type for %s", labelOrgIdpImport),
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"group", "user"}, false)),
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: fmt.Sprintf("Name of %s as supplied by the identity provider", labelOrgIdpImport),
			},
			"provider_type": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				Description:      fmt.Sprintf("Identity provider type for %s", labelOrgIdpImport),
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"LDAP", "OAUTH"}, false)),
			},
			"role_ids": {
				Type:        schema.TypeSet,
				Required:    true,
				MinItems:    1,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: fmt.Sprintf("%ss assigned directly to %s", labelVcfaRole, labelOrgIdpImport),
			},
			"inherit_group_roles": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: fmt.Sprintf("Whether %s inherits roles from its groups", labelOrgIdpImport),
			},
		},
	}
}

func resourceVcfaOrgIdpImportCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	if d.Get("principal_type").(string) == "group" && d.Get("inherit_group_roles").(bool) {
		return diag.Errorf("'inherit_group_roles' can only be enabled for user principals")
	}
	tenantContext, err := getTenantContextFromOrgId(tmClient, d.Get("org_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	roleIds := convertSchemaSetToSliceOfStrings(d.Get("role_ids").(*schema.Set))
	roleRefs := convertSliceOfStringsToOpenApiReferenceIds(roleIds)
	orgRef := &types.OpenApiReference{ID: tenantContext.OrgId, Name: tenantContext.OrgName}

	switch d.Get("principal_type").(string) {
	case "user":
		config := &types.OpenApiUser{
			Username:          d.Get("name").(string),
			NameInSource:      d.Get("name").(string),
			ProviderType:      d.Get("provider_type").(string),
			OrgEntityRef:      orgRef,
			RoleEntityRefs:    roleRefs,
			InheritGroupRoles: d.Get("inherit_group_roles").(bool),
		}
		created, err := tmClient.CreateUser(config, tenantContext)
		if err != nil {
			return diag.Errorf("error creating %s user: %s", labelOrgIdpImport, err)
		}
		d.SetId(created.User.ID)
	case "group":
		config := &types.OpenApiGroup{
			Name:           d.Get("name").(string),
			NameInSource:   d.Get("name").(string),
			ProviderType:   d.Get("provider_type").(string),
			OrgEntityRef:   orgRef,
			RoleEntityRefs: roleRefs,
		}
		created, err := tmClient.CreateGroup(config, tenantContext)
		if err != nil {
			return diag.Errorf("error creating %s group: %s", labelOrgIdpImport, err)
		}
		d.SetId(created.Group.ID)
	default:
		return diag.Errorf("unsupported principal type %q", d.Get("principal_type").(string))
	}

	return resourceVcfaOrgIdpImportRead(ctx, d, meta)
}

func resourceVcfaOrgIdpImportRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	tenantContext, err := getTenantContextFromOrgId(tmClient, d.Get("org_id").(string))
	if err != nil {
		if govcd.ContainsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	switch d.Get("principal_type").(string) {
	case "user":
		user, err := tmClient.GetUserById(d.Id(), tenantContext)
		if err != nil {
			if govcd.ContainsNotFound(err) {
				d.SetId("")
				return nil
			}
			return diag.Errorf("error retrieving %s user: %s", labelOrgIdpImport, err)
		}
		if err := setOrgIdpImportUserData(d, user); err != nil {
			return diag.FromErr(err)
		}
	case "group":
		group, err := tmClient.GetGroupById(d.Id(), tenantContext)
		if err != nil {
			if govcd.ContainsNotFound(err) {
				d.SetId("")
				return nil
			}
			return diag.Errorf("error retrieving %s group: %s", labelOrgIdpImport, err)
		}
		if err := setOrgIdpImportGroupData(d, group); err != nil {
			return diag.FromErr(err)
		}
	default:
		return diag.Errorf("unsupported principal type %q", d.Get("principal_type").(string))
	}

	return nil
}

func resourceVcfaOrgIdpImportUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	if d.Get("principal_type").(string) == "group" && d.Get("inherit_group_roles").(bool) {
		return diag.Errorf("'inherit_group_roles' can only be enabled for user principals")
	}
	tenantContext, err := getTenantContextFromOrgId(tmClient, d.Get("org_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	roleIds := convertSchemaSetToSliceOfStrings(d.Get("role_ids").(*schema.Set))
	roleRefs := convertSliceOfStringsToOpenApiReferenceIds(roleIds)

	switch d.Get("principal_type").(string) {
	case "user":
		user, err := tmClient.GetUserById(d.Id(), tenantContext)
		if err != nil {
			return diag.Errorf("error retrieving %s user: %s", labelOrgIdpImport, err)
		}
		config := *user.User
		config.RoleEntityRefs = roleRefs
		config.InheritGroupRoles = d.Get("inherit_group_roles").(bool)
		if _, err := user.Update(&config); err != nil {
			return diag.Errorf("error updating %s user: %s", labelOrgIdpImport, err)
		}
	case "group":
		group, err := tmClient.GetGroupById(d.Id(), tenantContext)
		if err != nil {
			return diag.Errorf("error retrieving %s group: %s", labelOrgIdpImport, err)
		}
		config := *group.Group
		config.RoleEntityRefs = roleRefs
		if _, err := group.Update(&config); err != nil {
			return diag.Errorf("error updating %s group: %s", labelOrgIdpImport, err)
		}
	default:
		return diag.Errorf("unsupported principal type %q", d.Get("principal_type").(string))
	}

	return resourceVcfaOrgIdpImportRead(ctx, d, meta)
}

func resourceVcfaOrgIdpImportDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	tenantContext, err := getTenantContextFromOrgId(tmClient, d.Get("org_id").(string))
	if err != nil {
		if govcd.ContainsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	switch d.Get("principal_type").(string) {
	case "user":
		user, err := tmClient.GetUserById(d.Id(), tenantContext)
		if err != nil {
			if govcd.ContainsNotFound(err) {
				d.SetId("")
				return nil
			}
			return diag.Errorf("error retrieving %s user: %s", labelOrgIdpImport, err)
		}
		if err := user.Delete(); err != nil {
			return diag.Errorf("error deleting %s user: %s", labelOrgIdpImport, err)
		}
	case "group":
		group, err := tmClient.GetGroupById(d.Id(), tenantContext)
		if err != nil {
			if govcd.ContainsNotFound(err) {
				d.SetId("")
				return nil
			}
			return diag.Errorf("error retrieving %s group: %s", labelOrgIdpImport, err)
		}
		if err := group.Delete(); err != nil {
			return diag.Errorf("error deleting %s group: %s", labelOrgIdpImport, err)
		}
	default:
		return diag.Errorf("unsupported principal type %q", d.Get("principal_type").(string))
	}

	d.SetId("")
	return nil
}

func resourceVcfaOrgIdpImportImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	tmClient := meta.(ClientContainer).tmClient

	idParts := strings.SplitN(d.Id(), "/", 3)
	if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
		return nil, fmt.Errorf("expected import ID to be <org_id>/<principal_type>/<name>")
	}
	if idParts[1] != "user" && idParts[1] != "group" {
		return nil, fmt.Errorf("principal type must be either user or group")
	}

	tenantContext, err := getTenantContextFromOrgId(tmClient, idParts[0])
	if err != nil {
		return nil, fmt.Errorf("error getting %s: %s", labelVcfaOrg, err)
	}

	var entityId string
	if idParts[1] == "user" {
		user, err := tmClient.GetUserByName(idParts[2], tenantContext)
		if err != nil {
			return nil, fmt.Errorf("error retrieving %s user: %s", labelOrgIdpImport, err)
		}
		entityId = user.User.ID
	} else {
		group, err := tmClient.GetGroupByName(idParts[2], tenantContext)
		if err != nil {
			return nil, fmt.Errorf("error retrieving %s group: %s", labelOrgIdpImport, err)
		}
		entityId = group.Group.ID
	}

	d.SetId(entityId)
	dSet(d, "org_id", idParts[0])
	dSet(d, "principal_type", idParts[1])
	dSet(d, "name", idParts[2])

	return []*schema.ResourceData{d}, nil
}

func setOrgIdpImportUserData(d *schema.ResourceData, user *govcd.OpenApiUser) error {
	if user == nil || user.User == nil {
		return fmt.Errorf("nil user structure")
	}

	d.SetId(user.User.ID)
	dSet(d, "name", user.User.Username)
	dSet(d, "provider_type", user.User.ProviderType)
	dSet(d, "inherit_group_roles", user.User.InheritGroupRoles)
	if user.User.OrgEntityRef != nil {
		dSet(d, "org_id", user.User.OrgEntityRef.ID)
	}

	roleIds := extractIdsFromOpenApiReferences(user.User.RoleEntityRefs)
	if err := d.Set("role_ids", roleIds); err != nil {
		return fmt.Errorf("error storing 'role_ids': %s", err)
	}
	return nil
}

func setOrgIdpImportGroupData(d *schema.ResourceData, group *govcd.OpenApiGroup) error {
	if group == nil || group.Group == nil {
		return fmt.Errorf("nil group structure")
	}

	d.SetId(group.Group.ID)
	dSet(d, "name", group.Group.Name)
	dSet(d, "provider_type", group.Group.ProviderType)
	if group.Group.OrgEntityRef != nil {
		dSet(d, "org_id", group.Group.OrgEntityRef.ID)
	}

	roleIds := extractIdsFromOpenApiReferences(group.Group.RoleEntityRefs)
	if err := d.Set("role_ids", roleIds); err != nil {
		return fmt.Errorf("error storing 'role_ids': %s", err)
	}
	return nil
}
