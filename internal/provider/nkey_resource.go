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
	KeyType    types.String `tfsdk:"type"`
	PublicKey  types.String `tfsdk:"public_key"`
	PrivateKey types.String `tfsdk:"private_key"`
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

	if err := data.generateKeys(); err != nil {
		resp.Diagnostics.AddError("generating nkey", err.Error())
		return
	}
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
	var plan NkeyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state NkeyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.KeyType.Equal(state.KeyType) {
		tflog.Debug(ctx, "key type changed. generating new key")
		if err := plan.generateKeys(); err != nil {
			resp.Diagnostics.AddError("generating nkey", err.Error())
			return
		}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
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

func (m *NkeyModel) generateKeys() (err error) {
	var keys nkeys.KeyPair

	switch strings.ToLower(m.KeyType.ValueString()) {
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
		return err
	}

	pubKey, err := keys.PublicKey()
	if err != nil {
		return err
	}
	privKey, err := keys.PrivateKey()
	if err != nil {
		return err
	}

	m.PublicKey = types.StringValue(pubKey)
	m.PrivateKey = types.StringValue(string(privKey))

	return nil
}
