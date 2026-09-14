package mcp

import (
	"context"
	"encoding/json"
	"testing"
)

func TestMCPInitialize(t *testing.T) {
	srv := NewServer(nil, nil, nil, nil, nil)

	req := Request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
	}
	raw, _ := json.Marshal(req)

	out := srv.HandleMessage(context.Background(), raw)
	if len(out) == 0 {
		t.Fatal("expected non-empty response")
	}

	var resp Response
	if err := json.Unmarshal(out, &resp); err != nil {
		t.Fatalf("failed unmarshaling response: %v", err)
	}

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}

	resultMap, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatalf("expected map result, got %T", resp.Result)
	}

	if resultMap["protocolVersion"] != "2024-11-05" {
		t.Errorf("expected protocolVersion 2024-11-05, got %v", resultMap["protocolVersion"])
	}

	serverInfo := resultMap["serverInfo"].(map[string]any)
	if serverInfo["name"] != "poly-cloud-mcp" {
		t.Errorf("expected server name poly-cloud-mcp, got %v", serverInfo["name"])
	}
}

func TestMCPToolsList(t *testing.T) {
	srv := NewServer(nil, nil, nil, nil, nil)

	req := Request{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/list",
	}
	raw, _ := json.Marshal(req)

	out := srv.HandleMessage(context.Background(), raw)
	var resp Response
	if err := json.Unmarshal(out, &resp); err != nil {
		t.Fatalf("failed unmarshaling response: %v", err)
	}

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}

	toolsMap := resp.Result.(map[string]any)
	tools := toolsMap["tools"].([]any)
	if len(tools) < 5 {
		t.Errorf("expected at least 5 tools, got %d", len(tools))
	}
}

func TestMCPInvalidJSON(t *testing.T) {
	srv := NewServer(nil, nil, nil, nil, nil)

	out := srv.HandleMessage(context.Background(), []byte("not valid json"))
	var resp Response
	if err := json.Unmarshal(out, &resp); err != nil {
		t.Fatalf("failed unmarshaling error response: %v", err)
	}

	if resp.Error == nil {
		t.Fatal("expected error, got nil")
	}

	if resp.Error.Code != ParseError {
		t.Errorf("expected code %d, got %d", ParseError, resp.Error.Code)
	}
}
