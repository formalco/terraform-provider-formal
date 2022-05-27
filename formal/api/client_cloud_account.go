package api

import (
	"encoding/json"
	"net/http"
	"strings"
)

type CloudIntegration struct {
	Id                   string `json:"id,omitempty"`
	AwsFormalId          string `json:"aws_formal_id,omitempty"`
	CloudAccountName     string `json:"cloud_account_name,omitempty"`
	CloudProvider        string `json:"cloud_provider,omitempty"`
	AwsFormalIamRole     string `json:"aws_formal_iam_role,omitempty"`
	AwsFormalHandshakeID string `json:"aws_formal_handshake_id,omitempty"`
	GCPProjectID         string `json:"gcp_project_id,omitempty"`
	TemplateBody         string `json:"aws_template_body,omitempty"`
	AwsFormalPingbackArn string `json:"aws_formal_pingback_arn,omitempty"`
	AwsFormalStackName   string `json:"aws_formal_stack_name,omitempty"`
	AwsCloudRegion       string `json:"aws_cloud_region,omitempty"`
}

func (c *Client) CreateCloudAccount(cloudAccountName, awsCloudRegion string) (*CloudIntegration, error) {
	// Compile
	p := CloudIntegration{
		CloudAccountName: cloudAccountName,
		AwsCloudRegion: awsCloudRegion,
	}

	// Send
	rb, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", c.HostURL+"/admin/integrations/cloud/aws/new", strings.NewReader(string(rb)))
	if err != nil {
		return nil, err
	}
	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	var ret CloudIntegration
	err = json.Unmarshal(body, &ret)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

type GetIntegrationsCloudAccountByIDRes struct {
	Integration CloudIntegration `json:"integration"`
}

func (c *Client) GetCloudAccount(cloudAccountFormalId string) (*CloudIntegration, error) {
	req, err := http.NewRequest("GET", c.HostURL+"/admin/integrations/cloud/aws/"+cloudAccountFormalId, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	cloudIntegrationRes := GetIntegrationsCloudAccountByIDRes{}
	err = json.Unmarshal(body, &cloudIntegrationRes)
	if err != nil {
		return nil, err
	}

	return &cloudIntegrationRes.Integration, nil
}
