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

	if len(plan.Assignments) > 0 {
		// Handle IP assignment
		tflog.Info(ctx, "Assigning IP addresses", map[string]interface{}{
			"assignments": plan.Assignments,
		})

		apiAssignments := make([]linodego.LinodeIPAssignment, len(plan.Assignments))
		for i, assignment := range plan.Assignments {
			apiAssignments[i] = linodego.LinodeIPAssignment{
				Address:  assignment.Address.ValueString(),
				LinodeID: int(assignment.LinodeID.ValueInt64()),
			}
		}

		assignOpts := linodego.LinodesAssignIPsOptions{
			Region:      plan.Region.ValueString(),
			Assignments: apiAssignments,
		}

		err := client.InstancesAssignIPs(ctx, assignOpts)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error assigning IP Addresses",
				fmt.Sprintf("Could not assign IP addresses: %s", err),
			)
			return
		}

		// Only set the necessary fields for IP assignment
		plan.ID = types.StringValue(plan.Assignments[0].Address.ValueString())
		plan.Address = plan.Assignments[0].Address
		plan.LinodeID = plan.Assignments[0].LinodeID
		plan.Region = types.StringValue(assignOpts.Region)

	} else if !plan.Public.IsNull() {
		// Handle IP creation
		createOpts := linodego.LinodeReserveIPOptions{
			Type:   "ipv4",
			Public: plan.Public.ValueBool(),
		}

		if !plan.LinodeID.IsNull() {
			createOpts.LinodeID = int(plan.LinodeID.ValueInt64())
		}

		if !plan.Reserved.IsNull() {
			createOpts.Reserved = plan.Reserved.ValueBool()
		}

		if !plan.Region.IsNull() {
			createOpts.Region = plan.Region.ValueString()
		}

		ip, err := client.AllocateReserveIP(ctx, createOpts)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error creating IP Address",
				fmt.Sprintf("Could not create IP address: %s", err),
			)
			return
		}

		// Only set the necessary fields after creation
		plan.ID = types.StringValue(ip.Address)
		plan.Address = types.StringValue(ip.Address)
		plan.LinodeID = types.Int64Value(int64(ip.LinodeID))
		plan.Public = types.BoolValue(ip.Public)
		plan.Reserved = types.BoolValue(ip.Reserved)
		plan.Region = types.StringValue(ip.Region)
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

	// Check if this is an assigned IP
	if len(state.Assignments) > 0 {
		// For assigned IPs, we just need to verify they still exist
		for _, assignment := range state.Assignments {
			ip, err := client.GetIPAddress(ctx, assignment.Address.ValueString())
			if err != nil {
				if lerr, ok := err.(*linodego.Error); ok && lerr.Code == 404 {
					resp.State.RemoveResource(ctx)
					return
				}
				resp.Diagnostics.AddError(
					"Error reading IP Address",
					fmt.Sprintf("Could not read IP address %s: %s", assignment.Address.ValueString(), err),
				)
				return
			}
			if ip.LinodeID != int(assignment.LinodeID.ValueInt64()) {
				resp.Diagnostics.AddWarning(
					"IP Assignment Changed",
					fmt.Sprintf("IP address %s is no longer assigned to Linode %d", ip.Address, assignment.LinodeID.ValueInt64()),
				)
			}
		}
	} else {
		// For created/reserved IPs
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

		tflog.Debug(ctx, fmt.Sprintf("API Response for IP read: %+v", ip))
		tflog.Debug(ctx, fmt.Sprintf("API Response Reserved status: %t", ip.Reserved))

		state.FlattenIPAddress(ip)
	}

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

	if len(plan.Assignments) > 0 {
		// Handle IP assignment updates
		apiAssignments := make([]linodego.LinodeIPAssignment, len(plan.Assignments))
		for i, assignment := range plan.Assignments {
			apiAssignments[i] = linodego.LinodeIPAssignment{
				Address:  assignment.Address.ValueString(),
				LinodeID: int(assignment.LinodeID.ValueInt64()),
			}
		}

		assignOpts := linodego.LinodesAssignIPsOptions{
			Region:      plan.Region.ValueString(),
			Assignments: apiAssignments,
		}

		err := client.InstancesAssignIPs(ctx, assignOpts)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating IP Assignments",
				fmt.Sprintf("Could not update IP assignments: %s", err),
			)
			return
		}

		// Update plan with new assignment details
		plan.ID = types.StringValue(plan.Assignments[0].Address.ValueString())
		plan.Address = plan.Assignments[0].Address
	}

	// Re-read the IP address to get the latest state
	ip, err := client.GetIPAddress(ctx, plan.Address.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading updated IP Address",
			fmt.Sprintf("Could not read updated IP address %s: %s", plan.Address.ValueString(), err),
		)
		return
	}
	plan.FlattenIPAddress(ip)

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

	// Check if this is an assigned IP
	if len(state.Assignments) > 0 {
		// For assigned IPs, we need to unassign them
		for _, assignment := range state.Assignments {
			linodeID := int(assignment.LinodeID.ValueInt64())
			ipAddress := assignment.Address.ValueString()

			err := client.DeleteInstanceIPAddress(ctx, linodeID, ipAddress)
			if err != nil {
				if lErr, ok := err.(*linodego.Error); (ok && lErr.Code != 404) || !ok {
					resp.Diagnostics.AddError(
						"Failed to Unassign IP",
						fmt.Sprintf(
							"failed to unassign ip (%s) from instance (%d): %s",
							ipAddress, linodeID, err.Error(),
						),
					)
				}
			}
		}
	} else {
		// For created/reserved IPs
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
}
