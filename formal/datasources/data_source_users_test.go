package datasources

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/require"

	corev1 "github.com/formalco/go-sdk/v3/core/v1"
)

type fakeUserLister struct {
	responses map[string]*corev1.ListUsersResponse
	err       error
	requests  []*corev1.ListUsersRequest
}

func (f *fakeUserLister) ListUsers(_ context.Context, req *corev1.ListUsersRequest) (*corev1.ListUsersResponse, error) {
	f.requests = append(f.requests, req)
	if f.err != nil {
		return nil, f.err
	}
	return f.responses[req.Cursor], nil
}

func TestListAllUsersPaginates(t *testing.T) {
	first := &corev1.User{Id: "usr_1"}
	second := &corev1.User{Id: "usr_2"}
	client := &fakeUserLister{
		responses: map[string]*corev1.ListUsersResponse{
			"": {
				Users:        []*corev1.User{first},
				ListMetadata: &corev1.ListMetadata{NextCursor: "usr_1"},
			},
			"usr_1": {
				Users:        []*corev1.User{second},
				ListMetadata: &corev1.ListMetadata{},
			},
		},
	}

	users, err := listAllUsers(t.Context(), client)

	require.NoError(t, err)
	require.Equal(t, []*corev1.User{first, second}, users)
	require.Equal(t, []*corev1.ListUsersRequest{
		{Limit: usersPageSize},
		{Limit: usersPageSize, Cursor: "usr_1"},
	}, client.requests)
}

func TestListAllUsersReturnsAPIError(t *testing.T) {
	client := &fakeUserLister{err: errors.New("list failed")}

	_, err := listAllUsers(t.Context(), client)

	require.ErrorContains(t, err, "list failed")
}

func TestListAllUsersRejectsRepeatedCursor(t *testing.T) {
	client := &fakeUserLister{
		responses: map[string]*corev1.ListUsersResponse{
			"": {
				ListMetadata: &corev1.ListMetadata{NextCursor: "usr_1"},
			},
			"usr_1": {
				ListMetadata: &corev1.ListMetadata{NextCursor: "usr_1"},
			},
		},
	}

	_, err := listAllUsers(t.Context(), client)

	require.ErrorContains(t, err, `same cursor "usr_1" twice`)
}

func TestFlattenUsers(t *testing.T) {
	users := []*corev1.User{
		{
			Id:                    "usr_human",
			DbUsername:            "idp:formal:human:jane@example.com",
			Type:                  "human",
			FullName:              "Jane Doe",
			GroupIds:              []string{"group_1"},
			TerminationProtection: true,
			Info: &corev1.User_Human_{
				Human: &corev1.User_Human{
					FirstName: "Jane",
					LastName:  "Doe",
					Email:     "jane@example.com",
				},
			},
		},
		{
			Id:         "usr_machine",
			DbUsername: "idp:formal:machine:example",
			Type:       "machine",
			FullName:   "Example",
		},
	}

	require.Equal(t, []map[string]any{
		{
			"id":                     "usr_human",
			"db_username":            "idp:formal:human:jane@example.com",
			"type":                   "human",
			"full_name":              "Jane Doe",
			"group_ids":              []string{"group_1"},
			"termination_protection": true,
			"first_name":             "Jane",
			"last_name":              "Doe",
			"email":                  "jane@example.com",
		},
		{
			"id":                     "usr_machine",
			"db_username":            "idp:formal:machine:example",
			"type":                   "machine",
			"full_name":              "Example",
			"group_ids":              []string(nil),
			"termination_protection": false,
		},
	}, flattenUsers(users))
}

func TestUserSchemasDefineLookupFieldsIndependently(t *testing.T) {
	common := commonUserSchema()
	lookup := User().Schema
	listed := Users().Schema["users"].Elem.(*schema.Resource).Schema

	require.NotContains(t, common, "id")
	require.NotContains(t, common, "db_username")
	require.False(t, lookup["id"].Computed)
	require.True(t, lookup["id"].Optional)
	require.Equal(t, []string{"id", "db_username"}, lookup["id"].ExactlyOneOf)
	require.False(t, lookup["db_username"].Computed)
	require.True(t, lookup["db_username"].Optional)
	require.Equal(t, []string{"id", "db_username"}, lookup["db_username"].ExactlyOneOf)
	require.True(t, listed["id"].Computed)
	require.False(t, listed["id"].Optional)
	require.True(t, listed["db_username"].Computed)
	require.False(t, listed["db_username"].Optional)
}
