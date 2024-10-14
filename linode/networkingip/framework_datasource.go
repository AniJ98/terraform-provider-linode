package networkingip

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/linode/linodego"
	"github.com/linode/terraform-provider-linode/v2/linode/helper"
)

func NewDataSource() datasource.DataSource {
	return &DataSource{
		BaseDataSource: helper.NewBaseDataSource(
			helper.BaseDataSourceConfig{
				Name:   "linode_networking_ip",
				Schema: &frameworkDatasourceSchema,
			},
		),
	}
}

type DataSource struct {
	helper.BaseDataSource
}

func (data *DataSourceModel) parseIP(ip *linodego.InstanceIP) {
	data.Address = types.StringValue(ip.Address)
	data.Gateway = types.StringValue(ip.Gateway)
	data.SubnetMask = types.StringValue(ip.SubnetMask)
	data.Prefix = types.Int64Value(int64(ip.Prefix))
	data.Type = types.StringValue(string(ip.Type))
	data.Public = types.BoolValue(ip.Public)
	data.RDNS = types.StringValue(ip.RDNS)
	data.LinodeID = types.Int64Value(int64(ip.LinodeID))
	data.Region = types.StringValue(ip.Region)

	id, _ := json.Marshal(ip)
	data.Reserved = types.BoolValue(ip.Reserved)
	data.ID = types.StringValue(string(id))
}

type DataSourceModel struct {
	Address     types.String `tfsdk:"address"`
	Gateway     types.String `tfsdk:"gateway"`
	SubnetMask  types.String `tfsdk:"subnet_mask"`
	Prefix      types.Int64  `tfsdk:"prefix"`
	Type        types.String `tfsdk:"type"`
	Public      types.Bool   `tfsdk:"public"`
	RDNS        types.String `tfsdk:"rdns"`
	LinodeID    types.Int64  `tfsdk:"linode_id"`
	Region      types.String `tfsdk:"region"`
	ID          types.String `tfsdk:"id"`
	Reserved    types.Bool   `tfsdk:"reserved"`
	IPAddresses types.List   `tfsdk:"ip_addresses"`
}

func (d *DataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	tflog.Debug(ctx, "Read data.linode_networking_ip")

	var data DataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.Address.IsNull() {
		// Fetch a specific IP address
		ctx = tflog.SetField(ctx, "address", data.Address.ValueString())

		ip, err := d.Meta.Client.GetIPAddress(ctx, data.Address.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to get IP Address",
				err.Error(),
			)
			return
		}

		data.parseIP(ip)
	} else {
		// List all IP addresses
		filter := ""
		if !data.Region.IsNull() {
			filter = fmt.Sprintf("{\"region\":\"%s\"}", data.Region.ValueString())
		}

		opts := &linodego.ListOptions{Filter: filter}
		ips, err := d.Meta.Client.ListIPAddresses(ctx, opts)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to list IP Addresses",
				err.Error(),
			)
			return
		}

		ipList := make([]DataSourceModel, len(ips))
		for i, ip := range ips {
			var ipModel DataSourceModel
			ipModel.parseIP(&ip)
			ipList[i] = ipModel
		}

		data.IPAddresses, _ = types.ListValueFrom(ctx, types.ObjectType{AttrTypes: DataSourceModel{}.AttrTypes()}, ipList)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (m DataSourceModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"address":     types.StringType,
		"region":      types.StringType,
		"gateway":     types.StringType,
		"subnet_mask": types.StringType,
		"prefix":      types.Int64Type,
		"type":        types.StringType,
		"public":      types.BoolType,
		"rdns":        types.StringType,
		"linode_id":   types.Int64Type,
		"reserved":    types.BoolType,
	}
}
