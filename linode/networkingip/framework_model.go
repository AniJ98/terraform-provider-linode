package networkingip

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/linode/linodego"
)

type NetworkingIPModel struct {
	ID         types.String `tfsdk:"id"`
	LinodeID   types.Int64  `tfsdk:"linode_id"`
	Reserved   types.Bool   `tfsdk:"reserved"`
	Region     types.String `tfsdk:"region"`
	Public     types.Bool   `tfsdk:"public"`
	Address    types.String `tfsdk:"address"`
	Gateway    types.String `tfsdk:"gateway"`
	SubnetMask types.String `tfsdk:"subnet_mask"`
	Prefix     types.Int64  `tfsdk:"prefix"`
	Type       types.String `tfsdk:"type"`
	RDNS       types.String `tfsdk:"rdns"`
}

func (m *NetworkingIPModel) FlattenIPAddress(ip *linodego.InstanceIP) {
	m.ID = types.StringValue(ip.Address)
	m.LinodeID = types.Int64Value(int64(ip.LinodeID))
	m.Reserved = types.BoolValue(ip.Type == "ipv4" && ip.LinodeID == 0)
	m.Region = types.StringValue(ip.Region)
	m.Public = types.BoolValue(ip.Public)
	m.Address = types.StringValue(ip.Address)
	m.Gateway = types.StringValue(ip.Gateway)
	m.SubnetMask = types.StringValue(ip.SubnetMask)
	m.Prefix = types.Int64Value(int64(ip.Prefix))
	m.Type = types.StringValue(string(ip.Type))
	m.RDNS = types.StringValue(ip.RDNS)
}
