// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/nats-io/nkeys"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &Nkey{}
var _ resource.ResourceWithImportState = &Nkey{}

func NewNkey() resource.Resource {
	return &Nkey{}
}

// Nkey defines the resource implementation.
type Nkey struct {
}

// NkeyModel describes the resource data model.
type NkeyModel struct {
	KeyType     types.String `tfsdk:"type"`
	Public_key  types.String `tfsdk:"public_key"`
	Private_key types.String `tfsdk:"private_key"`
}

func (r *Nkey) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nkey"
}

func (r *Nkey) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "An nkey is an ed25519 key pair formatted for use with NATS.",

		Attributes: map[string]schema.Attribute{
			"type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("account"),
				Description: "The type of nkey to generate. Must be one of user|account|server|cluster|operator|curve",
			},
			"public_key": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Public key of the nkey to be given in config to the nats server",
			},
			"private_key": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Private key of the nkey to be given to the client for authentication",
				Sensitive:           true,
			},
		},
	}
}

func (r *Nkey) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Nothing to do here as nkeys are simply generated
}

func (r *Nkey) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data NkeyModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var keys nkeys.KeyPair
	var err error
	switch strings.ToLower(data.KeyType.ValueString()) {
	case "user":
		keys, err = nkeys.CreateUser()
	case "account":
		keys, err = nkeys.CreateAccount()
	case "server":
		keys, err = nkeys.CreateServer()
	case "cluster":
		keys, err = nkeys.CreateCluster()
	case "operator":
		keys, err = nkeys.CreateOperator()
	case "curve":
		keys, err = nkeys.CreateCurveKeys()
	}

	if err != nil {
		resp.Diagnostics.AddError("generating nkey", err.Error())
		return
	}
	pubKey, err := keys.PublicKey()
	if err != nil {
		resp.Diagnostics.AddError("accessing public nkey", err.Error())
		return
	}

	data.Public_key = types.StringValue(pubKey)

	privKey, err := keys.PrivateKey()
	if err != nil {
		resp.Diagnostics.AddError("accessing private nkey", err.Error())
		return
	}

	data.Private_key = types.StringValue(string(privKey))

	tflog.Trace(ctx, "created nkey resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Nkey) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NkeyModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Nkey) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data NkeyModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Nkey) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data NkeyModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *Nkey) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
