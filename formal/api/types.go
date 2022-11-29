package api

type FieldEncryptionStruct struct {
	OrgId      string `json:"org_id"`
	DsId       string `json:"datastore_id"`
	FieldName  string `json:"name"`
	Path       string `json:"path"`
	KeyStorage string `json:"key_storage"`
	KeyId      string `json:"key_id"`
	Alg        string `json:"alg"`
}

// Used for datastore creation status
type DataStore struct {
	ID               string          `json:"id,omitempty"`
	Name             string          `json:"name,omitempty"`
	OriginalHostname string          `json:"original_hostname,omitempty"`
	FormalHostname   string          `json:"formal_hostname,omitempty"`
	Username         string          `json:"username"`
	CloudRegion      string          `json:"cloud_region"`
	CreatedAt        int64           `json:"created_at,omitempty"`
	ProxyStatus      string          `json:"proxy_status,omitempty"`
	DeploymentType   string          `json:"deployment_type,omitempty"`
	Policies         []PolicyOrgItem `json:"linked_policies"`
	CloudAccountId   string          `json:"cloud_account_id"`
	Deployed         bool            `json:"deployed"`
}

type DataStoreInfra struct {
	Id                    string `json:"id"`
	DsId                  string `json:"datastore_id"`
	OrganisationID        string `json:"org_id"`
	StackName             string `json:"stack_name"`
	Name                  string `json:"name"`
	Hostname              string `json:"hostname"`
	Port                  int    `json:"port"`
	Username              string `json:"username"`
	Password              string `json:"password"`
	HealthCheckDbName     string `json:"health_check_db_name"`
	FormalHostname        string `json:"formal_hostname"`
	Technology            string `json:"technology"`
	CloudProvider         string `json:"cloud_provider"`
	CloudRegion           string `json:"cloud_region"`
	DeploymentType        string `json:"deployment_type"`
	CloudAccountID        string `json:"cloud_account_id"`
	DataplaneID           string `json:"dataplane_id"`
	NetStackId            string `json:"net_stack_id"`
	FailOpen              bool   `json:"fail_open"`
	NetworkType           string `json:"network_type"`
	CreatedAt             int    `json:"created_at"`
	FullKMSDecryption     bool   `json:"global_kms_decrypt"`
	DefaultAccessBehavior string `json:"default_access_behavior"`
}

type CreatePolicyPayload struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	CreatedBy      string `json:"created_by"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
	Module         string `json:"module"`
	Active         bool   `json:"active"`
	OrganisationID string `json:"org_id"`
	SourceType     string `json:"source_type"`
}

type PolicyOrgItem struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	CreatedBy      string `json:"created_by"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
	Module         string `json:"module"`
	Active         bool   `json:"active"`
	OrganisationID string `json:"org_id"`
	SourceType     string `json:"source_type"`

	// below is diff vs payload
	ExpireAt string `json:"expire_at"`
	Status   string `json:"status"`
}

type Message struct {
	Message string `json:"message"`
}

type PolicyLinkStruct struct {
	// ID of this link
	ID string `json:"id"`
	// ID of the policy
	PolicyID string `json:"policy_id"`
	// OrganisationID string `json:"org_id"`
	// ID of the item it's linked to
	ItemID string `json:"item_id"`
	Type   string `json:"type"`
	// Active         bool   `json:"active"`
	ExpireAt string `json:"expire_at"`
}

type KeyStruct struct {
	Id          string `json:"id"`
	OrgId       string `json:"org_id"`
	KeyName     string `json:"name"`
	KeyId       string `json:"key_id"`
	CloudRegion string `json:"cloud_region"`
	KeyArn      string `json:"arn"`
	Active      bool   `json:"active"`
	KeyType     string `json:"key_type"`
	// ^ kms, gcp, hashicorp

	ManagedBy string `json:"managed_by"`
	// ^ managed_by, onprem, saas

	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
	CloudAccountID string `json:"cloud_account_id"`
}

type Role struct {
	ID string `json:"id"`
	// OrganisationID *string         `json:"organization_id"`
	// Formal/idp etc
	DBUsername string `json:"db_username"`
	Type       string `json:"type"`

	// Human
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`

	// Other
	// Role           *string         `json:"role"`
	// Idp        string          `json:"idp"`
	// IdpUserID  string          `json:"idp_user_id"`

	// Machine
	Name                   string `json:"name"`
	AppType                string `json:"app_type"`
	MachineRoleAccessToken string `json:"machine_role_access_token"` // returned in CREATE and GET routes. added for terraform

	// Status     string          `json:"status"`
	// Expire     int64           `json:"expire_at"`
	// Created    int64           `json:"created_at"`
	// Policies   []PolicyOrgItem `json:"linked_policies"`
}

type CreateRolePayload struct {
	// ID string `json:"id"`
	// OrganisationID *string `json:"organization_id"`
	// Formal/idp etc
	// DBUsername string `json:"db_username"`

	// Human
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Type      string `json:"type"`
	Email     string `json:"email"`

	// Other
	// Role       *string         `json:"role"`
	// Idp        string          `json:"idp"`
	// IdpUserID  string          `json:"idp_user_id"`

	// Machine
	Name    string `json:"name"`
	AppType string `json:"app_type"`

	// Status     string          `json:"status"`
	// Expire     int64           `json:"expire_at"`
	// Created    int64           `json:"created_at"`
	// Policies   []PolicyOrgItem `json:"linked_policies"`
}

type GroupStruct struct {
	ID string `json:"id"`
	// OrganisationID *string         `json:"org_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	// Active         bool            `json:"active"`
	// Status         string          `json:"status"`
	// Roles          []DbUser        `json:"roles"`
	// Policies       []PolicyOrgItem `json:"policies"`
	// Created string `json:"created_at"`
	RolesIDs []string `json:"user_ids"`
}

/*
8/18 Sync with database.shared
- Changed terraform `formal_public_route_table_id` param from being assigned FormalVpcPublicRouteTables. It is now assigned FormalVpcPublicRouteTableId
*/
type FlatDataplane struct {
	Id                            string      `json:"id"`
	OrgId                         string      `json:"org_id"`
	StackName                     string      `json:"name"`
	Region                        string      `json:"region"`
	CloudAccountId                string      `json:"cloud_account_id"`
	AvailabilityZone              int         `json:"availability_zone"`
	VpcPeeringConnectionId        string      `json:"vpc_peering_connection_id"`
	FormalR53PrivateHostedZoneId  string      `json:"formal_r53_private_hosted_zone_id"`
	FormalVpcFlowLogsGroupArn     string      `json:"formal_vpc_flow_logs_group_arn"`
	FormalVpcFlowLogGroupName     string      `json:"formal_vpc_flow_logs_group_name"`
	FormalVpcFlowLogsIamRoleArn   string      `json:"formal_vpc_flow_logs_iam_role_arn"`
	FormalVpcFlowLogsIamPolicyArn string      `json:"formal_vpc_flow_logs_iam_policy_arn"`
	InternetGateway               string      `json:"formal_vpc_igw_id"`
	EgressOnlyInternetGateway     string      `json:"egress_only_igw"`
	FormalVpcPrivateSubnetsIds    interface{} `json:"formal_vpc_private_subnets_ids"`
	FormalVpcPublicSubnetsIds     interface{} `json:"formal_vpc_public_subnets_ids"`
	FormalPublicSubnets           []string    `json:"formal_public_subnets"`
	FormalPrivateSubnets          []string    `json:"formal_private_subnets"`
	CustomerVpcRouteTables        interface{} `json:"customer_vpc_route_tables"`
	FormalNatGatewayIds           interface{} `json:"formal_vpc_natg_ids"`
	FormalVpcNatGatewayEips       interface{} `json:"formal_vpc_natg_eips"`
	FormalVpcPublicRouteTableId   string      `json:"formal_vpc_public_route_table_id"`
	FormalVpcPublicRouteTables    string      `json:"formal_vpc_public_route_tables"`
	FormalVpcPrivateRouteTables   []string    `json:"formal_vpc_private_route_table_routes"`
	FormalVpcId                   string      `json:"formal_vpc_id"`
	FormalVpcCidrBlock            string      `json:"formal_vpc_cidr_block"`
	EcsClusterName                string      `json:"ecs_cluster_name"`
	EcsClusterArn                 string      `json:"ecs_cluster_arn"`
	Status                        string      `json:"status"`
	VpcPeering                    bool        `json:"vpc_peering"`
}

type DataplaneRoutes struct {
	Id                     string `json:"id"`
	OrgId                  string `json:"org_id"`
	DataplaneId            string `json:"dataplane_id"`
	DestinationCidrBlock   string `json:"destination_cidr_block"`
	TransitGatewayId       string `json:"transit_gateway_id"`
	VpcPeeringConnectionId string `json:"vpc_peering_connection_id"`
	Deployed               bool   `json:"deployed"`
}
