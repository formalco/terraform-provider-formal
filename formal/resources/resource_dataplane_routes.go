package resource

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/formalco/terraform-provider-formal/formal/api"
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
		UpdateContext: resourceDataplaneRoutesUpdate,
		DeleteContext: resourceDataplaneRoutesDelete,

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
			},
			"transit_gateway_id": {
				// This description is used by the documentation generator and the language server.
				Description: "ID of the transit gateway.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"vpc_peering_connection_id": {
				// This description is used by the documentation generator and the language server.
				Description: "ID of the vpc peering connection.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"destination_cidr_block": {
				// This description is used by the documentation generator and the language server.
				Description: "CIDR block of the destination.",
				Type:        schema.TypeString,
				Required:    true,
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
	client := meta.(*(api.Client))

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	newDataplaneRoutes := api.DataplaneRoutes{
		DataplaneId:            d.Get("dataplane_id").(string),
		DestinationCidrBlock:   d.Get("destination_cidr_block").(string),
		TransitGatewayId:       d.Get("transit_gateway_id").(string),
		VpcPeeringConnectionId: d.Get("vpc_peering_connection_id").(string),
	}
	res, err := client.CreateDataplaneRoutes(newDataplaneRoutes)
	if err != nil {
		return diag.FromErr(err)
	}
	newDataplaneRoutesId := res.Id
	tflog.Info(ctx, newDataplaneRoutesId)
	if newDataplaneRoutesId == "" {
		return diag.FromErr(errors.New("created dataplane routes ID is empty, please try again later"))
	}
	time.Sleep(60 * time.Second)

	const ERROR_TOLERANCE = 5
	currentErrors := 0
	for {
		// Retrieve status
		existingDp, err := client.GetDataplaneRoutes(newDataplaneRoutesId)
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

		if existingDp == nil {
			err = errors.New("dataplane Routes with the given ID not found. It may have been deleted")
			return diag.FromErr(err)
		}
		if existingDp.Deployed {
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
	client := meta.(*api.Client)
	var diags diag.Diagnostics

	dataplaneId := d.Id()

	foundDataplaneRoutes, err := client.GetDataplaneRoutes(dataplaneId)
	if err != nil || foundDataplaneRoutes == nil {
		if strings.Contains(fmt.Sprint(err), "status: 404") {
			// Datplane was deleted
			tflog.Warn(ctx, "The dataplane routes was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	d.Set("id", foundDataplaneRoutes.Id)
	d.Set("dataplane_id", foundDataplaneRoutes.DataplaneId)
	d.Set("destination_cidr_block", foundDataplaneRoutes.DestinationCidrBlock)
	d.Set("transit_gateway_id", foundDataplaneRoutes.TransitGatewayId)
	d.Set("vpc_peering_connection_id", foundDataplaneRoutes.VpcPeeringConnectionId)
	d.Set("deployed", foundDataplaneRoutes.Deployed)

	// DsId is the UUID type id. See GetDataplaneInfraByDataplaneID in admin-api for more details
	d.SetId(foundDataplaneRoutes.Id)

	return diags
}

func resourceDataplaneRoutesUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return diag.Errorf("Dataplanes are immutable. Please delete and recreate this Dataplane.")
}

func resourceDataplaneRoutesDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	client := meta.(*api.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	routeId := d.Id()

	err := client.DeleteDataplaneRoutes(routeId)
	if err != nil {
		return diag.FromErr(err)
	}

	const ERROR_TOLERANCE = 5
	currentErrors := 0
	deleteTimeStart := time.Now()
	for {
		// Retrieve status
		_, err := client.GetDataplaneRoutes(routeId)
		if err != nil {
			if strings.Contains(fmt.Sprint(err), "status: 404") {
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
			tflog.Info(ctx, "Deletion of dataplane routes has taken more than 10m. The deletion process may be unhealthy and will be managed by the Formal. Exiting and marking as deleted.")
			break
		}

		time.Sleep(15 * time.Second)
	}

	d.SetId("")
	return diags
}
