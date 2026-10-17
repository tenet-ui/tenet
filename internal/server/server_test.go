package server

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tenet-ui/tenet/internal/designsystem"
)

const fixtureRoot = "../../testdata/project"

func connect(t *testing.T, root string) *mcp.ClientSession {
	t.Helper()
	ctx := context.Background()
	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	if _, err := New(root).Connect(ctx, serverTransport, nil); err != nil {
		t.Fatalf("server connect: %v", err)
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0.0.0"}, nil)
	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() { _ = session.Close() })
	return session
}

func callPointOfView(t *testing.T, session *mcp.ClientSession) *mcp.CallToolResult {
	t.Helper()
	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "get_point_of_view",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("call get_point_of_view: %v", err)
	}
	return res
}

func textOf(t *testing.T, res *mcp.CallToolResult) string {
	t.Helper()
	if len(res.Content) != 1 {
		t.Fatalf("got %d content items, want 1", len(res.Content))
	}
	tc, ok := res.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("got content %T, want *mcp.TextContent", res.Content[0])
	}
	return tc.Text
}

func TestListsOnlyGetPointOfView(t *testing.T) {
	session := connect(t, fixtureRoot)

	res, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("list tools: %v", err)
	}
	if len(res.Tools) != 1 {
		t.Fatalf("got %d tools, want 1", len(res.Tools))
	}
	if res.Tools[0].Name != "get_point_of_view" {
		t.Fatalf("got tool %q, want get_point_of_view", res.Tools[0].Name)
	}
	raw, err := json.Marshal(res.Tools[0].InputSchema)
	if err != nil {
		t.Fatalf("marshal input schema: %v", err)
	}
	var schema map[string]any
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatalf("unmarshal input schema %s: %v", raw, err)
	}
	if schema["type"] != "object" {
		t.Fatalf("input schema type = %v, want \"object\" (schema: %s)", schema["type"], raw)
	}
}

func TestCallReturnsFixtureText(t *testing.T) {
	session := connect(t, fixtureRoot)
	want, err := os.ReadFile(filepath.Join(fixtureRoot, designsystem.PointOfViewPath))
	if err != nil {
		t.Fatal(err)
	}

	res := callPointOfView(t, session)
	if res.IsError {
		t.Fatalf("got IsError, text: %s", textOf(t, res))
	}
	if got := textOf(t, res); got != string(want) {
		t.Fatalf("got %q, want %q", got, string(want))
	}
}

func TestCallMissingFileIsToolError(t *testing.T) {
	session := connect(t, t.TempDir())

	res := callPointOfView(t, session)
	if !res.IsError {
		t.Fatal("want IsError true")
	}
	if got := textOf(t, res); !strings.Contains(got, "not found") {
		t.Fatalf("got %q, want text containing \"not found\"", got)
	}
}

func TestCallEmptyFileIsToolError(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "design-system")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "point-of-view.md"), []byte("\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	session := connect(t, root)

	res := callPointOfView(t, session)
	if !res.IsError {
		t.Fatal("want IsError true")
	}
	if got := textOf(t, res); !strings.Contains(got, "empty") {
		t.Fatalf("got %q, want text containing \"empty\"", got)
	}
}

func TestUnknownToolIsNotASuccess(t *testing.T) {
	session := connect(t, fixtureRoot)

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: "nope", Arguments: map[string]any{}})
	if err == nil && (res == nil || !res.IsError) {
		t.Fatal("want an error or IsError for an unknown tool")
	}
}
