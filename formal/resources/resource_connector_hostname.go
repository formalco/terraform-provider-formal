package resource

import (
	"context"
	"fmt"
	"strings"
	"time"

	corev1 "buf.build/gen/go/formal/core/protocolbuffers/go/core/v1"
	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/formalco/terraform-provider-formal/formal/clients"
)

func ResourceConnectorHostname() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description:   "Registering a Connector Hostname with Formal.",
		CreateContext: resourceConnectorHostnameCreate,
		ReadContext:   resourceConnectorHostnameRead,
		UpdateContext: resourceConnectorHostnameUpdate,
		DeleteContext: resourceConnectorHostnameDelete,
		SchemaVersion: 1,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(25 * time.Minute),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				// This description is used by the documentation generator and the language server.
				Description: "The ID of this Connector Hostname.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"connector_id": {
				// This description is used by the documentation generator and the language server.
				Description: "The ID of the Connector this hostname is linked to.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"hostname": {
				// This description is used by the documentation generator and the language server.
				Description: "The hostname for this Connector hostname.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"managed_tls": {
				// This description is used by the documentation generator and the language server.
				Description: "If set to true, Formal will manage the TLS certificate for this hostname.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Deprecated:  "This field is deprecated and has no effect. It will be removed in a future release.",
			},
			"termination_protection": {
				// This description is used by the documentation generator and the language server.
				Description: "If set to true, this connector hostname cannot be deleted.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"dns_record": {
				// This description is used by the documentation generator and the language server.
				Description: "The DNS record for this hostname.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"tls_certificate_status": {
				// This description is used by the documentation generator and the language server.
				Description: "The status of the TLS certificate for this hostname. Accepted values are `none`, `issuing`, and `issued`.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"dns_record_status": {
				// This description is used by the documentation generator and the language server.
				Description: "The status of the DNS record for this hostname. Accepted values are `none`, `pending`, `success` and `failed`.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceConnectorHostnameCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	req := &corev1.CreateConnectorHostnameRequest{
		ConnectorId:           d.Get("connector_id").(string),
		Hostname:              d.Get("hostname").(string),
		TerminationProtection: d.Get("termination_protection").(bool),
		DnsRecord:             d.Get("dns_record").(string),
	}

	res, err := c.Grpc.Sdk.ConnectorServiceClient.CreateConnectorHostname(ctx, connect.NewRequest(req))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Msg.ConnectorHostname.Id)
	resourceConnectorHostnameRead(ctx, d, meta)
	return diags
}

func resourceConnectorHostnameRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	connectorHostnameId := d.Id()

	stateConf := &retry.StateChangeConf{
		Pending: []string{"tls_issuing", "dns_record_pending"}, // States we wait on
		Target:  []string{"ready"},                             // Custom target state
		Refresh: func() (interface{}, string, error) {
			req := connect.NewRequest(&corev1.GetConnectorHostnameRequest{
				Id: &corev1.GetConnectorHostnameRequest_HostnameId{
					HostnameId: connectorHostnameId,
				},
			})

			res, err := c.Grpc.Sdk.ConnectorServiceClient.GetConnectorHostname(ctx, req)
			if err != nil {
				if connect.CodeOf(err) == connect.CodeNotFound {
					tflog.Warn(ctx, "The Connector Hostname was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
					d.SetId("")
					return nil, "", fmt.Errorf("resource not found")
				}
				return nil, "", err
			}

			// Check TlsCertificateStatus and DnsRecordStatus
			tlsStatus := res.Msg.ConnectorHostname.TlsCertificateStatus
			dnsStatus := res.Msg.ConnectorHostname.DnsRecordStatus

			tflog.Info(ctx, "Polling for TLS and DNS readiness", map[string]interface{}{
				"tls_certificate_status": tlsStatus,
				"dns_record_status":      dnsStatus,
			})

			if tlsStatus == "issuing" {
				return res, "tls_issuing", nil
			}
			if dnsStatus == "pending" {
				return res, "dns_record_pending", nil
			}

			// Both statuses are ready
			return res, "ready", nil
		},
		Timeout:                   10 * time.Minute, // Set a timeout for the wait
		Delay:                     10 * time.Second, // Poll every 10 seconds
		MinTimeout:                5 * time.Second,
		ContinuousTargetOccurence: 1,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for TLS certificate and DNS record to become active: %s", err)
	}

	req := connect.NewRequest(&corev1.GetConnectorHostnameRequest{
		Id: &corev1.GetConnectorHostnameRequest_HostnameId{
			HostnameId: connectorHostnameId,
		},
	})

	res, err := c.Grpc.Sdk.ConnectorServiceClient.GetConnectorHostname(ctx, req)
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			tflog.Warn(ctx, "The Connector Hostname was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	d.Set("id", res.Msg.ConnectorHostname.Id)
	d.Set("connector_id", res.Msg.ConnectorHostname.Connector.Id)
	d.Set("hostname", res.Msg.ConnectorHostname.Hostname)
	d.Set("termination_protection", res.Msg.ConnectorHostname.TerminationProtection)
	d.Set("tls_certificate_status", res.Msg.ConnectorHostname.TlsCertificateStatus)
	d.Set("dns_record", res.Msg.ConnectorHostname.DnsRecord)
	d.Set("dns_record_status", res.Msg.ConnectorHostname.DnsRecordStatus)
	d.SetId(res.Msg.ConnectorHostname.Id)

	return diags
}

func resourceConnectorHostnameUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	connectorHostnameId := d.Id()

	fieldsThatCanChange := []string{"termination_protection", "dns_record"}
	if d.HasChangesExcept(fieldsThatCanChange...) {
		err := fmt.Sprintf("At the moment you can only update the following fields: %s. If you'd like to update other fields, please message the Formal team and we're happy to help.", strings.Join(fieldsThatCanChange, ", "))
		return diag.Errorf(err)
	}

	terminationProtection := d.Get("termination_protection").(bool)
	dnsRecord := d.Get("dns_record").(string)
	req := connect.NewRequest(&corev1.UpdateConnectorHostnameRequest{
		Id:                    connectorHostnameId,
		TerminationProtection: &terminationProtection,
		DnsRecord:             &dnsRecord,
	})

	_, err := c.Grpc.Sdk.ConnectorServiceClient.UpdateConnectorHostname(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}

	resourceConnectorHostnameRead(ctx, d, meta)
	return diags
}

func resourceConnectorHostnameDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	connectorHostnameId := d.Id()

	terminationProtection := d.Get("termination_protection").(bool)
	if terminationProtection {
		return diag.Errorf("Connector Hostname cannot be deleted because termination_protection is set to true")
	}

	req := connect.NewRequest(&corev1.DeleteConnectorHostnameRequest{
		Id: connectorHostnameId,
	})

	_, err := c.Grpc.Sdk.ConnectorServiceClient.DeleteConnectorHostname(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId("")
	return diags
}
