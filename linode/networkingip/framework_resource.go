package networkingip

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/linode/linodego"
	"github.com/linode/terraform-provider-linode/v2/linode/helper"
)

func NewResource() resource.Resource {
	return &Resource{
		BaseResource: helper.NewBaseResource(
			helper.BaseResourceConfig{
				Name:   "linode_networking_ip",
				IDType: types.StringType,
				Schema: &frameworkResourceSchema,
			},
		),
	}
}

type Resource struct {
	helper.BaseResource
}

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "Create linode_networking_ip")
	var plan NetworkingIPModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client := r.Meta.Client

	createOpts := linodego.LinodeReserveIPOptions{
		Type:   "ipv4",
		Public: plan.Public.ValueBool(),
	}

	if plan.Reserved.ValueBool() {
		createOpts.Reserved = true
		if !plan.Region.IsNull() {
			createOpts.Region = plan.Region.ValueString()
		} else if !plan.LinodeID.IsNull() {
			createOpts.LinodeID = helper.FrameworkSafeInt64ToInt(plan.LinodeID.ValueInt64(), &resp.Diagnostics)
		} else {
			resp.Diagnostics.AddError(
				"Invalid Configuration",
				"When reserved is true, either region or linode_id must be set.",
			)
			return
		}
	} else {
		if plan.LinodeID.IsNull() {
			resp.Diagnostics.AddError(
				"Invalid Configuration",
				"When reserved is false or not set, linode_id is required.",
			)
			return
		}
		createOpts.LinodeID = helper.FrameworkSafeInt64ToInt(plan.LinodeID.ValueInt64(), &resp.Diagnostics)
	}

	tflog.Debug(ctx, "client.AllocateReserveIP(...)", map[string]interface{}{
		"options": createOpts,
	})

	ip, err := client.AllocateReserveIP(ctx, createOpts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating IP Address",
			fmt.Sprintf("Could not create IP address: %s", err),
		)
		return
	}

	plan.FlattenIPAddress(ip)

	if !plan.RDNS.IsNull() {
		updateOpts := linodego.IPAddressUpdateOptions{
			RDNS: plan.RDNS.ValueStringPointer(),
		}
		_, err := client.UpdateIPAddress(ctx, ip.Address, updateOpts)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error setting RDNS",
				fmt.Sprintf("Could not set RDNS for IP address: %s", err),
			)
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Read linode_networking_ip")
	var state NetworkingIPModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client := r.Meta.Client
	ip, err := client.GetIPAddress(ctx, state.ID.ValueString())
	if err != nil {
		if lerr, ok := err.(*linodego.Error); ok && lerr.Code == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading IP Address",
			fmt.Sprintf("Could not read IP address %s: %s", state.ID.ValueString(), err),
		)
		return
	}

	state.FlattenIPAddress(ip)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Debug(ctx, "Update linode_networking_ip")
	var plan, state NetworkingIPModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client := r.Meta.Client

	if !plan.RDNS.Equal(state.RDNS) {
		updateOpts := linodego.IPAddressUpdateOptions{
			RDNS: plan.RDNS.ValueStringPointer(),
		}
		_, err := client.UpdateIPAddress(ctx, state.ID.ValueString(), updateOpts)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating IP Address",
				fmt.Sprintf("Could not update IP address %s: %s", state.ID.ValueString(), err),
			)
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "Delete linode_networking_ip")
	var state NetworkingIPModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client := r.Meta.Client

	if !state.Reserved.ValueBool() {
		// This is a regular IP address
		linodeID := helper.FrameworkSafeInt64ToInt(state.LinodeID.ValueInt64(), &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		err := client.DeleteInstanceIPAddress(ctx, linodeID, state.Address.ValueString())
		if err != nil {
			if lErr, ok := err.(*linodego.Error); (ok && lErr.Code != 404) || !ok {
				resp.Diagnostics.AddError(
					"Failed to Delete IP",
					fmt.Sprintf(
						"failed to delete instance (%d) ip (%s): %s",
						linodeID, state.Address.ValueString(), err.Error(),
					),
				)
			}
		}
	} else {
		// This is a reserved IP address
		err := client.DeleteReservedIPAddress(ctx, state.Address.ValueString())
		if err != nil {
			if lErr, ok := err.(*linodego.Error); (ok && lErr.Code != 404) || !ok {
				resp.Diagnostics.AddError(
					"Failed to Delete Reserved IP",
					fmt.Sprintf(
						"failed to delete reserved ip (%s): %s",
						state.Address.ValueString(), err.Error(),
					),
				)
			}
		}
	}
}
