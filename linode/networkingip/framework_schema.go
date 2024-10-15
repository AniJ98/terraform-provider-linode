package networkingip

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

var frameworkResourceSchema = schema.Schema{
	Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "The ID of the IPv4 address.",
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"linode_id": schema.Int64Attribute{
			Description: "The ID of the Linode to allocate an IPv4 address for. Required when reserved is false or not set.",
			Optional:    true,
		},
		"reserved": schema.BoolAttribute{
			Description: "Whether the IPv4 address should be reserved.",
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(false),
		},
		"region": schema.StringAttribute{
			Description: "The region for the reserved IPv4 address. Required when reserved is true and linode_id is not set.",
			Optional:    true,
		},
		"public": schema.BoolAttribute{
			Description: "Whether the IPv4 address is public or private.",
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(true),
		},
		"address": schema.StringAttribute{
			Description: "The allocated IPv4 address.",
			Computed:    true,
		},
		"gateway": schema.StringAttribute{
			Description: "The default gateway for this address.",
			Computed:    true,
		},
		"subnet_mask": schema.StringAttribute{
			Description: "The subnet mask for this address.",
			Computed:    true,
		},
		"prefix": schema.Int64Attribute{
			Description: "The number of bits set in the subnet mask.",
			Computed:    true,
		},
		"type": schema.StringAttribute{
			Description: "The type of IP address (ipv4).",
			Computed:    true,
		},
		"rdns": schema.StringAttribute{
			Description: "The reverse DNS assigned to this address.",
			Optional:    true,
			Computed:    true,
		},
	},
}
