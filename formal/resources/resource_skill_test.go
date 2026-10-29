package resource

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/require"

	formal "github.com/formalco/go-sdk/v3"
	corev1 "github.com/formalco/go-sdk/v3/core/v1"
	"github.com/formalco/go-sdk/v3/core/v1/corev1connect"
	"github.com/formalco/terraform-provider-formal/formal/api"
	"github.com/formalco/terraform-provider-formal/formal/clients"
)

const testSkillMD = `---
name: release-notes
description: Write release notes from merged PRs.
---
Summarize every merged PR since the last tag.
`

type skillTestServer struct {
	corev1connect.UnimplementedSkillServiceHandler

	mu       sync.Mutex
	skills   []*corev1.SkillVariant
	pageSize int
	deleted  []string
}

func (s *skillTestServer) CreateFormalSkill(_ context.Context, req *connect.Request[corev1.CreateFormalSkillRequest]) (*connect.Response[corev1.CreateFormalSkillResponse], error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	parsed, err := parseSkillMD(req.Msg.GetSkillMd())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	for _, existing := range s.skills {
		if existing.Name == parsed.Name {
			return nil, connect.NewError(connect.CodeAlreadyExists, errors.New("skill already exists"))
		}
	}
	skill := &corev1.SkillVariant{Id: "skill_" + parsed.Name, Name: parsed.Name, Description: parsed.Description, Contents: parsed.Contents}
	s.skills = append(s.skills, skill)
	return connect.NewResponse(&corev1.CreateFormalSkillResponse{Skill: skill}), nil
}

func (s *skillTestServer) ListSkills(_ context.Context, req *connect.Request[corev1.ListSkillsRequest]) (*connect.Response[corev1.ListSkillsResponse], error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var matches []*corev1.SkillVariant
	for _, skill := range s.skills {
		if strings.Contains(strings.ToLower(skill.Name), strings.ToLower(req.Msg.NameFilter)) {
			matches = append(matches, skill)
		}
	}
	start := 0
	if req.Msg.Cursor != "" {
		for i, skill := range matches {
			if skill.Id == req.Msg.Cursor {
				start = i + 1
			}
		}
	}
	pageSize := s.pageSize
	if pageSize == 0 {
		pageSize = int(req.Msg.Limit)
	}
	end := min(start+pageSize, len(matches))
	res := &corev1.ListSkillsResponse{Skills: matches[start:end]}
	if end < len(matches) {
		res.NextCursor = matches[end-1].Id
	}
	return connect.NewResponse(res), nil
}

func (s *skillTestServer) DeleteFormalSkill(_ context.Context, req *connect.Request[corev1.DeleteFormalSkillRequest]) (*connect.Response[corev1.DeleteFormalSkillResponse], error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, skill := range s.skills {
		if skill.Name == req.Msg.Name {
			s.skills = append(s.skills[:i], s.skills[i+1:]...)
			s.deleted = append(s.deleted, req.Msg.Name)
			return connect.NewResponse(&corev1.DeleteFormalSkillResponse{}), nil
		}
	}
	return nil, connect.NewError(connect.CodeNotFound, errors.New("not found"))
}

func newSkillTestClients(t *testing.T, server *skillTestServer) *clients.Clients {
	t.Helper()

	mux := http.NewServeMux()
	path, handler := corev1connect.NewSkillServiceHandler(server)
	mux.Handle(path, handler)
	httpServer := httptest.NewServer(mux)
	t.Cleanup(httpServer.Close)

	sdkClient, err := formal.New(
		formal.WithAPIKey("test-api-key"),
		formal.WithBaseURL(httpServer.URL),
	)
	require.NoError(t, err)

	return &clients.Clients{Grpc: &api.GrpcClient{Sdk: sdkClient}}
}

func TestSkillCreateStoresParsedFields(t *testing.T) {
	t.Parallel()

	server := &skillTestServer{}
	meta := newSkillTestClients(t, server)
	data := schema.TestResourceDataRaw(t, ResourceSkill().Schema, map[string]any{"skill_md": testSkillMD})

	diags := ResourceSkill().CreateContext(t.Context(), data, meta)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	require.Equal(t, "skill_release-notes", data.Id())
	require.Equal(t, "release-notes", data.Get("name"))
	require.Equal(t, "Write release notes from merged PRs.", data.Get("description"))
	require.Equal(t, testSkillMD, data.Get("skill_md"), "configured SKILL.md must be kept when it matches the remote skill")
}

func TestSkillReadFindsExactNameAcrossPages(t *testing.T) {
	t.Parallel()

	server := &skillTestServer{
		pageSize: 1,
		skills: []*corev1.SkillVariant{
			{Id: "skill_1", Name: "release-notes-v2", Description: "Other skill.", Contents: "Other.\n"},
			{Id: "skill_2", Name: "release-notes", Description: "Write release notes from merged PRs.", Contents: "Summarize every merged PR since the last tag.\n"},
		},
	}
	meta := newSkillTestClients(t, server)
	data := schema.TestResourceDataRaw(t, ResourceSkill().Schema, map[string]any{"skill_md": testSkillMD})
	data.SetId("skill_2")
	require.NoError(t, data.Set("name", "release-notes"))

	diags := ResourceSkill().ReadContext(t.Context(), data, meta)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	require.Equal(t, "skill_2", data.Id())
	require.Equal(t, testSkillMD, data.Get("skill_md"))
}

func TestSkillReadRecordsRemoteDrift(t *testing.T) {
	t.Parallel()

	server := &skillTestServer{
		skills: []*corev1.SkillVariant{
			{Id: "skill_2", Name: "release-notes", Description: "Write release notes from merged PRs.", Contents: "Edited in the console.\n"},
		},
	}
	meta := newSkillTestClients(t, server)
	data := schema.TestResourceDataRaw(t, ResourceSkill().Schema, map[string]any{"skill_md": testSkillMD})
	data.SetId("skill_2")
	require.NoError(t, data.Set("name", "release-notes"))

	diags := ResourceSkill().ReadContext(t.Context(), data, meta)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)

	remote, err := parseSkillMD(data.Get("skill_md").(string))
	require.NoError(t, err)
	require.Equal(t, "Edited in the console.\n", remote.Contents)
	require.False(t, suppressEquivalentSkillMD("", data.Get("skill_md").(string), testSkillMD, nil), "drift must produce a diff")
}

func TestSkillReadClearsMissingSkill(t *testing.T) {
	t.Parallel()

	meta := newSkillTestClients(t, &skillTestServer{})
	data := schema.TestResourceDataRaw(t, ResourceSkill().Schema, map[string]any{"skill_md": testSkillMD})
	data.SetId("skill_gone")
	require.NoError(t, data.Set("name", "release-notes"))

	diags := ResourceSkill().ReadContext(t.Context(), data, meta)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	require.Empty(t, data.Id())
}

func TestSkillDeleteByNameToleratesMissingSkill(t *testing.T) {
	t.Parallel()

	server := &skillTestServer{
		skills: []*corev1.SkillVariant{{Id: "skill_2", Name: "release-notes", Description: "d", Contents: "c\n"}},
	}
	meta := newSkillTestClients(t, server)
	data := schema.TestResourceDataRaw(t, ResourceSkill().Schema, map[string]any{"skill_md": testSkillMD})
	data.SetId("skill_2")
	require.NoError(t, data.Set("name", "release-notes"))

	diags := ResourceSkill().DeleteContext(t.Context(), data, meta)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	require.Equal(t, []string{"release-notes"}, server.deleted)

	data.SetId("skill_2")
	diags = ResourceSkill().DeleteContext(t.Context(), data, meta)
	require.False(t, diags.HasError(), "deleting an already deleted skill must succeed: %v", diags)
}

func TestSkillImportByName(t *testing.T) {
	t.Parallel()

	server := &skillTestServer{
		skills: []*corev1.SkillVariant{
			{Id: "skill_2", Name: "release-notes", Description: "Write release notes: from merged PRs.", Contents: "Summarize.\n"},
		},
	}
	meta := newSkillTestClients(t, server)
	data := schema.TestResourceDataRaw(t, ResourceSkill().Schema, nil)
	data.SetId("release-notes")

	imported, err := ResourceSkill().Importer.StateContext(t.Context(), data, meta)
	require.NoError(t, err)
	require.Len(t, imported, 1)
	require.Equal(t, "skill_2", imported[0].Id())

	parsed, err := parseSkillMD(imported[0].Get("skill_md").(string))
	require.NoError(t, err)
	require.Equal(t, parsedSkillMD{Name: "release-notes", Description: "Write release notes: from merged PRs.", Contents: "Summarize.\n"}, parsed)
}

func TestSkillImportFailsForUnknownName(t *testing.T) {
	t.Parallel()

	meta := newSkillTestClients(t, &skillTestServer{})
	data := schema.TestResourceDataRaw(t, ResourceSkill().Schema, nil)
	data.SetId("missing")

	_, err := ResourceSkill().Importer.StateContext(t.Context(), data, meta)
	require.ErrorContains(t, err, "not found")
}

func TestSuppressEquivalentSkillMD(t *testing.T) {
	t.Parallel()

	reformatted := "---\r\ndescription: \"Write release notes from merged PRs.\"\r\nname:   release-notes\r\n---\r\nSummarize every merged PR since the last tag.\r\n"
	require.True(t, suppressEquivalentSkillMD("", testSkillMD, reformatted, nil))

	renamed := strings.Replace(testSkillMD, "name: release-notes", "name: changelog", 1)
	require.False(t, suppressEquivalentSkillMD("", testSkillMD, renamed, nil))

	edited := strings.Replace(testSkillMD, "Summarize", "List", 1)
	require.False(t, suppressEquivalentSkillMD("", testSkillMD, edited, nil))

	require.False(t, suppressEquivalentSkillMD("", "", testSkillMD, nil))
}
