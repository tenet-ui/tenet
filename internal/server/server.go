package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tenet-ui/tenet/internal/designsystem"
)

// Version is the server version reported to MCP clients.
const Version = "0.1.0"

// GetPointOfViewInput is the input of get_point_of_view. The tool takes no arguments.
type GetPointOfViewInput struct{}

// New returns an MCP server rooted at root, with the get_point_of_view tool registered.
func New(root string) *mcp.Server {
	s := mcp.NewServer(&mcp.Implementation{Name: "tenet", Version: Version}, nil)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_point_of_view",
		Description: "Returns the brand point of view of this project's design system (design-system/point-of-view.md).",
	}, func(_ context.Context, _ *mcp.CallToolRequest, _ GetPointOfViewInput) (*mcp.CallToolResult, any, error) {
		return getPointOfView(root), nil, nil
	})
	return s
}

// getPointOfView never returns a Go error: file problems become IsError results
// so that the client sees a readable tool failure, not a protocol failure.
func getPointOfView(root string) *mcp.CallToolResult {
	text, err := designsystem.ReadPointOfView(root)
	if err != nil {
		return errorResult(err, root)
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: text}}}
}

func errorResult(err error, root string) *mcp.CallToolResult {
	var msg string
	switch {
	case errors.Is(err, designsystem.ErrPointOfViewMissing):
		msg = fmt.Sprintf("%s not found under %s. Create it to describe the brand point of view.", designsystem.PointOfViewPath, root)
	case errors.Is(err, designsystem.ErrPointOfViewEmpty):
		msg = fmt.Sprintf("%s under %s is empty. Write the point of view before using this tool.", designsystem.PointOfViewPath, root)
	default:
		msg = fmt.Sprintf("could not read %s: %v", designsystem.PointOfViewPath, err)
	}
	return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: msg}}}
}
