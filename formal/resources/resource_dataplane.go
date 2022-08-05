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

func ResourceDataplane() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description:   "Creating a Dataplane with Formal.",
		CreateContext: resourceDataplaneCreate,
		ReadContext:   resourceDataplaneRead,
		UpdateContext: resourceDataplaneUpdate,
		DeleteContext: resourceDataplaneDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(25 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"id": {
				// This description is used by the documentation generator and the language server.
				Description: "ID of this dataplane.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				// This description is used by the documentation generator and the language server.
				Description: "Friendly name for this dataplane.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"customer_vpc_id": {
				// This description is used by the documentation generator and the language server.
				Description: "The VPC ID that this dataplane should be deployed in.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"availability_zones": {
				// This description is used by the documentation generator and the language server.
				Description: "Number of availability zones.",
				Type:        schema.TypeInt,
				Required:    true,
			},
			"cloud_account_id": {
				// This description is used by the documentation generator and the language server.
				Description: "Cloud account ID for deploying the dataplane.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"cloud_region": {
				// This description is used by the documentation generator and the language server.
				Description: "The cloud region the dataplane should be deployed in.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"formal_public_route_table_id": {
				// This description is used by the documentation generator and the language server.
				Description: "The public route table ID for the dataplane.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"formal_private_route_table_ids": {
				// This description is used by the documentation generator and the language server.
				Description: "The private route table IDs created with this dataplane.",
				Type:        schema.TypeList,
				Computed:    true,
			},
			"formal_vpc_id": {
				// This description is used by the documentation generator and the language server.
				Description: "The VPC ID created with this dataplane.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceDataplaneCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	client := meta.(*(api.Client))

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	newDataplane := api.FlatDataplane{
		StackName:             d.Get("name").(string),
		CustomerVpcId:         d.Get("customer_vpc_id").(string),
		OriginalCustomerVpcId: d.Get("customer_vpc_id").(string),
		CloudAccountId:        d.Get("cloud_account_id").(string),
		Region:                d.Get("cloud_region").(string),
		AvailabilityZone:      d.Get("availability_zones").(int),
	}

	res, err := client.CreateDataplane(newDataplane)
	if err != nil {
		return diag.FromErr(err)
	}
	newDataPlaneId := res.Id
	tflog.Info(ctx, newDataPlaneId)
	if newDataPlaneId == "" {
		return diag.FromErr(errors.New("created dataplane ID is empty, please try again later"))
	}
	time.Sleep(60 * time.Second)

	var formalPublicRouteTableId string
	var formalPrivateRouteTableIds []string
	var formalVpcId string

	const ERROR_TOLERANCE = 5
	currentErrors := 0
	for {
		// Retrieve status
		existingDp, err := client.GetDataplane(newDataPlaneId)
		if err != nil {
			if currentErrors >= ERROR_TOLERANCE {
				return diag.FromErr(err)
			} else {
				tflog.Warn(ctx, "Experienced error #"+strconv.Itoa(currentErrors+1)+" checking on DataplaneStatus: ", map[string]interface{}{"err": err})
				currentErrors += 1
				time.Sleep(15 * time.Second)
				continue
			}
		}

		// Found

		if existingDp == nil {
			err = errors.New("dataplane with the given ID not found. It may have been deleted")
			return diag.FromErr(err)
		}
		if existingDp.Status == "healthy" {
			formalPublicRouteTableId = existingDp.FormalPublicRouteTableId
			formalPrivateRouteTableIds = existingDp.FormalVpcPrivateRouteTables
			formalVpcId = existingDp.FormalVpcId
			break
		} else {
			time.Sleep(15 * time.Second)
		}
	}

	// DsId is the UUID type id. See GetDataplaneInfraByDataplaneID in admin-api for more details
	d.SetId(newDataPlaneId)
	d.Set("formal_public_route_table_id", formalPublicRouteTableId)
	d.Set("formal_private_route_table_ids", formalPrivateRouteTableIds)
	d.Set("formal_vpc_id", formalVpcId)

	resourceDataplaneRead(ctx, d, meta)

	return diags
}

func resourceDataplaneRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	client := meta.(*api.Client)
	var diags diag.Diagnostics

	dataplaneId := d.Id()

	foundDataplane, err := client.GetDataplane(dataplaneId)
	if err != nil || foundDataplane == nil {
		if strings.Contains(fmt.Sprint(err), "status: 404") {
			// Datplane was deleted
			tflog.Warn(ctx, "The dataplane was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	d.Set("name", foundDataplane.StackName)
	d.Set("customer_vpc_id", foundDataplane.CustomerVpcId)
	d.Set("cloud_account_id", foundDataplane.CloudAccountId)
	d.Set("cloud_region", foundDataplane.Region)
	d.Set("availability_zones", foundDataplane.AvailabilityZone)
	d.Set("formal_public_route_table_id", foundDataplane.FormalPublicRouteTableId)
	d.Set("formal_private_route_table_ids", foundDataplane.FormalVpcPrivateRouteTables)
	d.Set("formal_vpc_id", foundDataplane.FormalVpcId)
	d.Set("id", foundDataplane.Id)

	// DsId is the UUID type id. See GetDataplaneInfraByDataplaneID in admin-api for more details
	d.SetId(foundDataplane.Id)

	return diags
}

func resourceDataplaneUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return diag.Errorf("Dataplanes are immutable. Please delete and recreate this Dataplane.")
}

func resourceDataplaneDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	client := meta.(*api.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	dataplaneId := d.Id()

	err := client.DeleteDataplane(dataplaneId)
	if err != nil {
		return diag.FromErr(err)
	}

	const ERROR_TOLERANCE = 2
	currentErrors := 0
	deleteTimeStart := time.Now()
	for {
		// Retrieve status
		_, err := client.GetDataplane(dataplaneId)
		if err != nil {
			if strings.Contains(fmt.Sprint(err), "status: 404") {
				// Dataplane was deleted
				break
			}

			// Handle other errors
			if currentErrors >= ERROR_TOLERANCE {
				return diag.FromErr(err)
			} else {
				tflog.Warn(ctx, "Experienced an error #"+strconv.Itoa(currentErrors)+" checking on DataplaneStatus: ", map[string]interface{}{"err": err})
				currentErrors += 1
			}
		}

		if time.Since(deleteTimeStart) > time.Minute*10 {
			tflog.Info(ctx, "Deletion of dataplane has taken more than 10m. The deletion process may be unhealthy and will be managed by the Formal. Exiting and marking as deleted.")
			break
		}

		time.Sleep(15 * time.Second)
	}

	d.SetId("")
	return diags
}
