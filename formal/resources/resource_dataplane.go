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

func ResourceDataplane() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description:   "Creating a Dataplane with Formal.",
		CreateContext: resourceDataplaneCreate,
		ReadContext:   resourceDataplaneRead,
		// UpdateContext: resourceDataplaneUpdate,
		DeleteContext: resourceDataplaneDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
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
				ForceNew:    true,
			},
			"availability_zones": {
				// This description is used by the documentation generator and the language server.
				Description: "Number of availability zones.",
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
			},
			"cloud_account_id": {
				// This description is used by the documentation generator and the language server.
				Description: "Cloud account ID for deploying the dataplane.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"cloud_region": {
				// This description is used by the documentation generator and the language server.
				Description: "The cloud region the dataplane should be deployed in.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
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
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
			"formal_vpc_id": {
				// This description is used by the documentation generator and the language server.
				Description: "The VPC ID created with this dataplane.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"formal_vpc_cidr_block": {
				// This description is used by the documentation generator and the language server.
				Description: "The VPC CIDR block created with this dataplane.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"formal_private_subnets": {
				// This description is used by the documentation generator and the language server.
				Description: "The private subnet IDs created with this dataplane.",
				Type:        schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
			"formal_public_subnets": {
				// This description is used by the documentation generator and the language server.
				Description: "The public subnet IDs created with this dataplane.",
				Type:        schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
			"formal_r53_private_hosted_zone_id": {
				// This description is used by the documentation generator and the language server.
				Description: "The id of the AWS Route 53 Private Zone Formal creates in your account.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"vpc_peering": {
				// This description is used by the documentation generator and the language server.
				Description: "Set to true to enable VPC peering.",
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
			},
		},
	}
}

func resourceDataplaneCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	// Wait, just in case the apply that triggered this resource creation also included creating the cloud acc integration, but the dataplane resource does
	// not have a dependency on the aws_cloudformation resource, so we need to wait for that to be complete before doing dataplane elements.
	time.Sleep(60 * time.Second)

	StackName := d.Get("name").(string)
	CloudAccountId := d.Get("cloud_account_id").(string)
	Region := d.Get("cloud_region").(string)
	AvailabilityZone := d.Get("availability_zones").(int)
	VpcPeering := d.Get("vpc_peering").(bool)

	res, err := c.Grpc.Sdk.CloudServiceClient.CreateDataplane(ctx, connect.NewRequest(&adminv1.CreateDataplaneRequest{
		Name:             StackName,
		Region:           Region,
		CloudAccountId:   CloudAccountId,
		AvailabilityZone: int32(AvailabilityZone),
		VpcPeering:       VpcPeering,
	}))

	if err != nil {
		return diag.FromErr(err)
	}
	newDataPlaneId := res.Msg.Dataplane.Id
	tflog.Info(ctx, newDataPlaneId)
	if newDataPlaneId == "" {
		return diag.FromErr(errors.New("could not initiate a dataplane creation at this time. Please try again later"))
	}

	time.Sleep(30 * time.Second)

	const ERROR_TOLERANCE = 5
	currentErrors := 0
	for {
		// Retrieve status
		existingDp, err := c.Grpc.Sdk.CloudServiceClient.GetDataplaneById(ctx, connect.NewRequest(&adminv1.GetDataplaneByIdRequest{Id: newDataPlaneId}))
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
		if existingDp.Msg.Dataplane.Status == "healthy" {
			break
		} else {
			time.Sleep(15 * time.Second)
		}
	}

	// DsId is the UUID type id. See GetDataplaneInfraByDataplaneID in admin-api for more details
	d.SetId(newDataPlaneId)

	resourceDataplaneRead(ctx, d, meta)

	return diags
}

func resourceDataplaneRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	dataplaneId := d.Id()

	res, err := c.Grpc.Sdk.CloudServiceClient.GetDataplaneById(ctx, connect.NewRequest(&adminv1.GetDataplaneByIdRequest{Id: dataplaneId}))
	if err != nil {
		if status.Code(err) == codes.NotFound {
			tflog.Warn(ctx, "The dataplane was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	d.Set("name", res.Msg.Dataplane.Name)
	d.Set("cloud_account_id", res.Msg.Dataplane.CloudAccountId)
	d.Set("cloud_region", res.Msg.Dataplane.Region)
	d.Set("availability_zones", res.Msg.Dataplane.AvailabilityZone)
	d.Set("formal_public_route_table_id", res.Msg.Dataplane.FormalVpcPublicRouteTableId)
	d.Set("formal_private_route_table_ids", res.Msg.Dataplane.FormalVpcPrivateRouteTableRoutes)
	d.Set("formal_vpc_id", res.Msg.Dataplane.FormalVpcId)
	d.Set("formal_vpc_cidr_block", res.Msg.Dataplane.FormalVpcCidrBlock)
	d.Set("formal_private_subnets", res.Msg.Dataplane.FormalPrivateSubnets)
	d.Set("formal_public_subnets", res.Msg.Dataplane.FormalPublicSubnets)
	d.Set("formal_r53_private_hosted_zone_id", res.Msg.Dataplane.FormalR53PrivateHostedZoneId)
	d.Set("id", res.Msg.Dataplane.Id)

	// DsId is the UUID type id. See GetDataplaneInfraByDataplaneID in admin-api for more details
	d.SetId(res.Msg.Dataplane.Id)

	return diags
}

// func resourceDataplaneUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
// 	return diag.Errorf("Dataplanes are immutable. You can contact the Formal team for assistance.")
// }

func resourceDataplaneDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	dataplaneId := d.Id()
	_, err := c.Grpc.Sdk.CloudServiceClient.DeleteDataplane(ctx, connect.NewRequest(&adminv1.DeleteDataplaneRequest{Id: dataplaneId}))
	if err != nil {
		return diag.FromErr(err)
	}

	const ERROR_TOLERANCE = 5
	currentErrors := 0
	deleteTimeStart := time.Now()
	for {
		// Retrieve status
		_, err := c.Grpc.Sdk.CloudServiceClient.GetDataplaneById(ctx, connect.NewRequest(&adminv1.GetDataplaneByIdRequest{Id: dataplaneId}))
		if err != nil {
			if status.Code(err) == codes.NotFound {
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
