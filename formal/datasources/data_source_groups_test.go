package datasources

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/require"

	corev1 "github.com/formalco/go-sdk/v3/core/v1"
)

type fakeGroupLister struct {
	responses map[string]*corev1.ListGroupsResponse
	err       error
	requests  []*corev1.ListGroupsRequest
}

func (f *fakeGroupLister) ListGroups(_ context.Context, req *corev1.ListGroupsRequest) (*corev1.ListGroupsResponse, error) {
	f.requests = append(f.requests, req)
	if f.err != nil {
		return nil, f.err
	}
	return f.responses[req.Cursor], nil
}

func TestListAllGroupsPaginates(t *testing.T) {
	first := &corev1.Group{Id: "group_1"}
	second := &corev1.Group{Id: "group_2"}
	client := &fakeGroupLister{
		responses: map[string]*corev1.ListGroupsResponse{
			"": {
				Groups:       []*corev1.Group{first},
				ListMetadata: &corev1.ListMetadata{NextCursor: "group_1"},
			},
			"group_1": {
				Groups:       []*corev1.Group{second},
				ListMetadata: &corev1.ListMetadata{},
			},
		},
	}

	groups, err := listAllGroups(t.Context(), client)

	require.NoError(t, err)
	require.Equal(t, []*corev1.Group{first, second}, groups)
	require.Equal(t, []*corev1.ListGroupsRequest{
		{Limit: groupsPageSize},
		{Limit: groupsPageSize, Cursor: "group_1"},
	}, client.requests)
}

func TestListAllGroupsReturnsAPIError(t *testing.T) {
	client := &fakeGroupLister{err: errors.New("list failed")}

	_, err := listAllGroups(t.Context(), client)

	require.ErrorContains(t, err, "list failed")
}

func TestListAllGroupsRejectsRepeatedCursor(t *testing.T) {
	client := &fakeGroupLister{
		responses: map[string]*corev1.ListGroupsResponse{
			"": {
				ListMetadata: &corev1.ListMetadata{NextCursor: "group_1"},
			},
			"group_1": {
				ListMetadata: &corev1.ListMetadata{NextCursor: "group_1"},
			},
		},
	}

	_, err := listAllGroups(t.Context(), client)

	require.ErrorContains(t, err, `same cursor "group_1" twice`)
}

func TestFlattenGroups(t *testing.T) {
	groups := []*corev1.Group{
		{
			Id:                    "group_eng",
			Name:                  "Engineering",
			Description:           "Engineers with database access",
			DsyncGroupId:          "directory_group_eng",
			TerminationProtection: true,
		},
		{
			Id:   "group_ops",
			Name: "Operations",
		},
	}

	require.Equal(t, []map[string]any{
		{
			"id":                     "group_eng",
			"name":                   "Engineering",
			"description":            "Engineers with database access",
			"termination_protection": true,
		},
		{
			"id":                     "group_ops",
			"name":                   "Operations",
			"description":            "",
			"termination_protection": false,
		},
	}, flattenGroups(groups))
}

func TestGroupSchemasDefineLookupFieldsIndependently(t *testing.T) {
	common := commonGroupSchema()
	lookup := Group().Schema
	listed := Groups().Schema["groups"].Elem.(*schema.Resource).Schema

	require.NotContains(t, common, "id")
	require.NotContains(t, common, "name")
	require.False(t, lookup["id"].Computed)
	require.True(t, lookup["id"].Optional)
	require.Equal(t, []string{"id", "name"}, lookup["id"].ExactlyOneOf)
	require.False(t, lookup["name"].Computed)
	require.True(t, lookup["name"].Optional)
	require.Equal(t, []string{"id", "name"}, lookup["name"].ExactlyOneOf)
	require.True(t, listed["id"].Computed)
	require.False(t, listed["id"].Optional)
	require.True(t, listed["name"].Computed)
	require.False(t, listed["name"].Optional)
}
