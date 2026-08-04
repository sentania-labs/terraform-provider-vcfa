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

const labelVcfaOrgRegionalNetworkingAviSetting = "Regional Networking Avi Setting"

func resourceVcfaOrgRegionalNetworkingAviSetting() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcfaOrgRegionalNetworkingAviSettingCreateUpdate,
		ReadContext:   resourceVcfaOrgRegionalNetworkingAviSettingRead,
		UpdateContext: resourceVcfaOrgRegionalNetworkingAviSettingCreateUpdate,
		DeleteContext: resourceVcfaOrgRegionalNetworkingAviSettingDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcfaOrgRegionalNetworkingAviSettingImport,
		},

		Schema: map[string]*schema.Schema{
			"regional_networking_setting_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: fmt.Sprintf("ID of %s", labelVcfaRegionalNetworkingSetting),
			},
			"active": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: fmt.Sprintf("Whether %s is enabled", labelVcfaOrgRegionalNetworkingAviSetting),
			},
			"service_engine_group_mode": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      fmt.Sprintf("Service engine group management mode for %s", labelVcfaOrgRegionalNetworkingAviSetting),
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"PROVIDER_MANAGED", "TENANT_MANAGED"}, false)),
			},
			"service_engine_quota": {
				Type:             schema.TypeInt,
				Required:         true,
				Description:      fmt.Sprintf("Service engine quota for %s", labelVcfaOrgRegionalNetworkingAviSetting),
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(0)),
			},
			"service_engine_group_refs": {
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: fmt.Sprintf("Provider-managed service engine group IDs for %s", labelVcfaOrgRegionalNetworkingAviSetting),
			},
			"application_limit": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: fmt.Sprintf("Application limit reported for %s", labelVcfaOrgRegionalNetworkingAviSetting),
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Realization status of %s", labelVcfaOrgRegionalNetworkingAviSetting),
			},
		},
	}
}

func resourceVcfaOrgRegionalNetworkingAviSettingCreateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	rns, err := tmClient.GetTmRegionalNetworkingSettingById(d.Get("regional_networking_setting_id").(string))
	if err != nil {
		return diag.Errorf("error looking up %s by ID: %s", labelVcfaRegionalNetworkingSetting, err)
	}

	d.SetId(rns.TmRegionalNetworkingSetting.ID)
	cfg, err := getTmOrgRegionalNetworkingAviSettingType(d)
	if err != nil {
		return diag.Errorf("error getting %s configuration: %s", labelVcfaOrgRegionalNetworkingAviSetting, err)
	}

	if _, err := rns.UpdateAviSetting(cfg); err != nil {
		return diag.Errorf("error setting %s configuration: %s", labelVcfaOrgRegionalNetworkingAviSetting, err)
	}

	return resourceVcfaOrgRegionalNetworkingAviSettingRead(ctx, d, meta)
}

func resourceVcfaOrgRegionalNetworkingAviSettingRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	rns, err := tmClient.GetTmRegionalNetworkingSettingById(d.Id())
	if err != nil {
		if govcd.ContainsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("error looking up %s by ID: %s", labelVcfaRegionalNetworkingSetting, err)
	}

	aviSetting, err := rns.GetAviSetting()
	if err != nil {
		if govcd.ContainsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("error retrieving %s: %s", labelVcfaOrgRegionalNetworkingAviSetting, err)
	}

	d.SetId(rns.TmRegionalNetworkingSetting.ID)
	if err := setTmOrgRegionalNetworkingAviSettingData(d, aviSetting); err != nil {
		return diag.Errorf("error storing %s configuration to state: %s", labelVcfaOrgRegionalNetworkingAviSetting, err)
	}
	return nil
}

func resourceVcfaOrgRegionalNetworkingAviSettingDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	rns, err := tmClient.GetTmRegionalNetworkingSettingById(d.Get("regional_networking_setting_id").(string))
	if err != nil {
		if govcd.ContainsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("error looking up %s by ID: %s", labelVcfaRegionalNetworkingSetting, err)
	}

	if _, err := rns.UpdateAviSetting(&types.TmRegionalNetworkingAviSetting{Active: false}); err != nil {
		return diag.Errorf("error disabling %s: %s", labelVcfaOrgRegionalNetworkingAviSetting, err)
	}

	d.SetId("")
	return nil
}

func resourceVcfaOrgRegionalNetworkingAviSettingImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	tmClient := meta.(ClientContainer).tmClient

	idParts := strings.Split(d.Id(), ImportSeparator)
	if len(idParts) != 2 {
		return nil, fmt.Errorf("ID syntax should be <%s name>%s<%s name>", labelVcfaOrg, ImportSeparator, labelVcfaRegionalNetworkingSetting)
	}

	org, err := tmClient.GetTmOrgByName(idParts[0])
	if err != nil {
		return nil, fmt.Errorf("error retrieving %s '%s': %s", labelVcfaOrg, idParts[0], err)
	}

	rns, err := tmClient.GetTmRegionalNetworkingSettingByNameAndOrgId(idParts[1], org.TmOrg.ID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving %s '%s' within %s '%s': %s",
			labelVcfaRegionalNetworkingSetting, idParts[1], labelVcfaOrg, idParts[0], err)
	}

	d.SetId(rns.TmRegionalNetworkingSetting.ID)
	dSet(d, "regional_networking_setting_id", rns.TmRegionalNetworkingSetting.ID)
	return []*schema.ResourceData{d}, nil
}

func getTmOrgRegionalNetworkingAviSettingType(d *schema.ResourceData) (*types.TmRegionalNetworkingAviSetting, error) {
	active := d.Get("active").(bool)
	config := &types.TmRegionalNetworkingAviSetting{Active: active}
	if !active {
		return config, nil
	}

	mode := d.Get("service_engine_group_mode").(string)
	quota := d.Get("service_engine_quota").(int)
	groupIds := convertSchemaSetToSliceOfStrings(d.Get("service_engine_group_refs").(*schema.Set))

	if mode == "PROVIDER_MANAGED" && len(groupIds) == 0 {
		return nil, fmt.Errorf("'service_engine_group_refs' must contain at least one ID when 'service_engine_group_mode' is PROVIDER_MANAGED")
	}
	if mode == "TENANT_MANAGED" && len(groupIds) > 0 {
		return nil, fmt.Errorf("'service_engine_group_refs' must be empty when 'service_engine_group_mode' is TENANT_MANAGED")
	}

	config.ServiceEngineGroupMode = mode
	config.ServiceEngineQuota = quota
	if len(groupIds) > 0 {
		config.ServiceEngineGroupRefs = convertSliceOfStringsToOpenApiReferenceIds(groupIds)
	}
	return config, nil
}

func setTmOrgRegionalNetworkingAviSettingData(d *schema.ResourceData, setting *types.TmRegionalNetworkingAviSetting) error {
	if setting == nil {
		return fmt.Errorf("nil Avi setting structure")
	}

	dSet(d, "active", setting.Active)
	dSet(d, "service_engine_group_mode", setting.ServiceEngineGroupMode)
	dSet(d, "service_engine_quota", setting.ServiceEngineQuota)
	if setting.ApplicationLimit != nil {
		dSet(d, "application_limit", *setting.ApplicationLimit)
	} else {
		dSet(d, "application_limit", 0)
	}
	if setting.Status != nil {
		dSet(d, "status", *setting.Status)
	} else {
		dSet(d, "status", "")
	}

	groupIds := extractIdsFromOpenApiReferences(setting.ServiceEngineGroupRefs)
	if err := d.Set("service_engine_group_refs", groupIds); err != nil {
		return fmt.Errorf("error storing 'service_engine_group_refs': %s", err)
	}
	return nil
}
