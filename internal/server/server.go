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
	mcp.AddTool(s, &mcp.Tool{
		Name:        "critique",
		Description: "Checks one UI source file against the design system's rules that have a Forbid: pattern. Returns one finding per match, naming the rule, the line, the matched text and the fix. Pass file, the path to the source file; relative paths resolve against the server root, so pass an absolute path for a file elsewhere. system is optional and defaults to the design system this server serves; leave it out.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in CritiqueInput) (*mcp.CallToolResult, any, error) {
		return critiqueTool(root, in), nil, nil
	})
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_tokens",
		Description: "Returns the design tokens of one category (for example color) and the valid rules linked to it. Pass the category name, which is the first part of a token's name.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in GetTokensInput) (*mcp.CallToolResult, any, error) {
		return deliverTokens(root, in.Category), nil, nil
	})
	for _, d := range []struct {
		name, desc string
		k          deliveryTarget
	}{
		{"get_component", "Returns the design component with the given id (its file name without .md) and the valid rules linked to it.", componentKind},
		{"get_template", "Returns the design template with the given id (its file name without .md) and the valid rules linked to it.", templateKind},
		{"get_flow", "Returns the design flow with the given id (its file name without .md) and the valid rules linked to it.", flowKind},
	} {
		mcp.AddTool(s, &mcp.Tool{Name: d.name, Description: d.desc}, func(_ context.Context, _ *mcp.CallToolRequest, in GetItemInput) (*mcp.CallToolResult, any, error) {
			return deliverItem(root, d.k, in.ID), nil, nil
		})
	}
	return s
}

// getPointOfView never returns a Go error: file problems become IsError results
// so that the client sees a readable tool failure, not a protocol failure.
func getPointOfView(root string) *mcp.CallToolResult {
	text, err := designsystem.ReadPointOfView(root)
	if err != nil {
		return errorResult(err, root)
	}
	return textResult(text)
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
	return toolError(msg)
}
