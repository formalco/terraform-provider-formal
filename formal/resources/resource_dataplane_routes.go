package resource

import (
	adminv1 "buf.build/gen/go/formal/admin/protocolbuffers/go/admin/v1"
	"context"
	"errors"
	"github.com/bufbuild/connect-go"
	"github.com/formalco/terraform-provider-formal/formal/clients"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceDataplaneRoutes() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description:   "Attaching Routes with Formal Dataplane.",
		CreateContext: resourceDataplaneRoutesCreate,
		ReadContext:   resourceDataplaneRoutesRead,
		// UpdateContext: resourceDataplaneRoutesUpdate,
		DeleteContext: resourceDataplaneRoutesDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(25 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"id": {
				// This description is used by the documentation generator and the language server.
				Description: "ID of this transit gateway attached to dataplane.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"dataplane_id": {
				// This description is used by the documentation generator and the language server.
				Description: "ID of the dataplane.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"transit_gateway_id": {
				// This description is used by the documentation generator and the language server.
				Description: "ID of the transit gateway.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"vpc_peering_connection_id": {
				// This description is used by the documentation generator and the language server.
				Description: "ID of the vpc peering connection.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"destination_cidr_block": {
				// This description is used by the documentation generator and the language server.
				Description: "CIDR block of the destination.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"deployed": {
				// This description is used by the documentation generator and the language server.
				Description: "Whether the dataplane is deployed.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
		},
	}
}

func resourceDataplaneRoutesCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	DataplaneId := d.Get("dataplane_id").(string)
	DestinationCidrBlock := d.Get("destination_cidr_block").(string)
	TransitGatewayId := d.Get("transit_gateway_id").(string)
	VpcPeeringConnectionId := d.Get("vpc_peering_connection_id").(string)

	res, err := c.Grpc.Sdk.CloudServiceClient.CreateDataplaneRoutes(ctx, connect.NewRequest(&adminv1.CreateDataplaneRoutesRequest{
		DataplaneId:            DataplaneId,
		DestinationCidrBlock:   DestinationCidrBlock,
		TransitGatewayId:       TransitGatewayId,
		VpcPeeringConnectionId: VpcPeeringConnectionId,
	}))
	if err != nil {
		return diag.FromErr(err)
	}
	newDataplaneRoutesId := res.Msg.DataplaneRoutes.Id
	tflog.Info(ctx, newDataplaneRoutesId)
	if newDataplaneRoutesId == "" {
		return diag.FromErr(errors.New("created dataplane routes ID is empty, please try again later"))
	}
	time.Sleep(60 * time.Second)

	const ERROR_TOLERANCE = 5
	currentErrors := 0
	for {
		// Retrieve status
		existingDp, err := c.Grpc.Sdk.CloudServiceClient.GetDataplaneRoutesById(ctx, connect.NewRequest(&adminv1.GetDataplaneRoutesByIdRequest{Id: newDataplaneRoutesId}))
		if err != nil {
			if currentErrors >= ERROR_TOLERANCE {
				return diag.FromErr(err)
			} else {
				tflog.Warn(ctx, "Experienced error #"+strconv.Itoa(currentErrors+1)+" checking on Dataplane Routes Status: ", map[string]interface{}{"err": err})
				currentErrors += 1
				time.Sleep(15 * time.Second)
				continue
			}
		}

		// Found

		if existingDp.Msg.DataplaneRoutes.Deployed {
			break
		} else {
			time.Sleep(15 * time.Second)
		}
	}

	// DsId is the UUID type id. See GetDataplaneInfraByDataplaneID in admin-api for more details
	d.SetId(newDataplaneRoutesId)

	resourceDataplaneRoutesRead(ctx, d, meta)

	return diags
}

func resourceDataplaneRoutesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	dataplaneId := d.Id()

	foundDataplaneRoutes, err := c.Grpc.Sdk.CloudServiceClient.GetDataplaneRoutesById(ctx, connect.NewRequest(&adminv1.GetDataplaneRoutesByIdRequest{Id: dataplaneId}))
	if err != nil || foundDataplaneRoutes == nil {
		if status.Code(err) == codes.NotFound {
			// Datplane was deleted
			tflog.Warn(ctx, "The dataplane routes was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	d.Set("id", foundDataplaneRoutes.Msg.DataplaneRoutes.Id)
	d.Set("dataplane_id", foundDataplaneRoutes.Msg.DataplaneRoutes.DataplaneId)
	d.Set("destination_cidr_block", foundDataplaneRoutes.Msg.DataplaneRoutes.DestinationCidrBlock)
	d.Set("transit_gateway_id", foundDataplaneRoutes.Msg.DataplaneRoutes.TransitGatewayId)
	d.Set("vpc_peering_connection_id", foundDataplaneRoutes.Msg.DataplaneRoutes.VpcPeeringConnectionId)
	d.Set("deployed", foundDataplaneRoutes.Msg.DataplaneRoutes.Deployed)

	// DsId is the UUID type id. See GetDataplaneInfraByDataplaneID in admin-api for more details
	d.SetId(foundDataplaneRoutes.Msg.DataplaneRoutes.Id)

	return diags
}

// func resourceDataplaneRoutesUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
// 	return diag.Errorf("Dataplanes are immutable. Please delete and recreate this Dataplane.")
// }

func resourceDataplaneRoutesDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	routeId := d.Id()

	_, err := c.Grpc.Sdk.CloudServiceClient.DeleteDataplaneRoutes(ctx, connect.NewRequest(&adminv1.DeleteDataplaneRoutesRequest{Id: routeId}))
	if err != nil {
		return diag.FromErr(err)
	}

	const ERROR_TOLERANCE = 5
	currentErrors := 0
	deleteTimeStart := time.Now()
	for {
		// Retrieve status
		_, err := c.Grpc.Sdk.CloudServiceClient.GetDataplaneRoutesById(ctx, connect.NewRequest(&adminv1.GetDataplaneRoutesByIdRequest{Id: routeId}))
		if err != nil {
			if status.Code(err) == codes.NotFound {
				// Dataplane was deleted
				break
			}

			// Handle other errors
			if currentErrors >= ERROR_TOLERANCE {
				return diag.FromErr(err)
			} else {
				tflog.Warn(ctx, "Experienced an error #"+strconv.Itoa(currentErrors)+" checking on Dataplane Routes Status: ", map[string]interface{}{"err": err})
				currentErrors += 1
			}
		}

		if time.Since(deleteTimeStart) > time.Minute*10 {
			tflog.Info(ctx, "Deletion of dataplane routes has taken more than 10m. The deletion process may be unhealthy and will be managed by Formal. Exiting and marking as deleted.")
			break
		}

		time.Sleep(15 * time.Second)
	}

	d.SetId("")
	return diags
}
