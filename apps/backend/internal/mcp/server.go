package mcp

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/polycloud/platform/apps/backend/internal/config"
	"github.com/polycloud/platform/apps/backend/internal/domain"
	"github.com/polycloud/platform/apps/backend/internal/index"
	"github.com/polycloud/platform/apps/backend/internal/storage"
)

type Server struct {
	cfg      *config.Config
	files    storage.Service
	accounts *storage.AccountService
	folders  *index.FolderRepo
	log      *slog.Logger
	userID   string
}

func NewServer(cfg *config.Config, files storage.Service, accounts *storage.AccountService, folders *index.FolderRepo, log *slog.Logger) *Server {
	userID := "usr_default"
	if cfg != nil && cfg.DefaultUserID != "" {
		userID = cfg.DefaultUserID
	}
	return &Server{
		cfg:      cfg,
		files:    files,
		accounts: accounts,
		folders:  folders,
		log:      log,
		userID:   userID,
	}
}

// HandleMessage memproses pesan tunggal JSON-RPC dan mengembalikan responsnya.
func (s *Server) HandleMessage(ctx context.Context, raw []byte) []byte {
	var req Request
	if err := json.Unmarshal(raw, &req); err != nil {
		resp := Response{
			JSONRPC: "2.0",
			Error:   &Error{Code: ParseError, Message: "Invalid JSON-RPC request: " + err.Error()},
		}
		out, _ := json.Marshal(resp)
		return out
	}

	// Notifikasi seperti notifications/initialized tidak memerlukan respon berulang jika ID kosong
	if req.ID == nil && strings.HasPrefix(req.Method, "notifications/") {
		return nil
	}

	resp := s.dispatch(ctx, req)
	out, err := json.Marshal(resp)
	if err != nil {
		errResp := Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &Error{Code: InternalError, Message: "Failed to serialize response: " + err.Error()},
		}
		out, _ = json.Marshal(errResp)
	}
	return out
}

func (s *Server) dispatch(ctx context.Context, req Request) Response {
	resp := Response{
		JSONRPC: "2.0",
		ID:      req.ID,
	}

	switch req.Method {
	case "initialize":
		resp.Result = InitializeResult{
			ProtocolVersion: "2024-11-05",
			Capabilities: ServerCapabilities{
				Tools:     &ToolsCapability{ListChanged: false},
				Resources: &ResourcesCapability{Subscribe: false, ListChanged: false},
			},
			ServerInfo: ServerInfo{
				Name:    "poly-cloud-mcp",
				Version: "1.0.0",
			},
		}

	case "ping":
		resp.Result = map[string]any{}

	case "tools/list":
		resp.Result = ToolsListResult{Tools: s.definedTools()}

	case "tools/call":
		var params ToolCallParams
		if len(req.Params) > 0 {
			if err := json.Unmarshal(req.Params, &params); err != nil {
				resp.Error = &Error{Code: InvalidParams, Message: "Invalid tool call parameters"}
				return resp
			}
		}
		res, err := s.callTool(ctx, params.Name, params.Arguments)
		if err != nil {
			resp.Result = ToolCallResult{
				Content: []TextContent{{Type: "text", Text: fmt.Sprintf("Error executing tool %s: %v", params.Name, err)}},
				IsError: true,
			}
		} else {
			resp.Result = res
		}

	case "resources/list":
		resp.Result = ResourcesListResult{
			Resources: []Resource{
				{
					URI:         "polycloud://quota",
					Name:        "Storage Quota Report",
					Description: "Live aggregated multi-cloud storage capacity across Google Drive, OneDrive, S3, Dropbox, etc.",
					MimeType:    "application/json",
				},
				{
					URI:         "polycloud://accounts",
					Name:        "Connected Cloud Accounts",
					Description: "List of all connected cloud storage provider remotes.",
					MimeType:    "application/json",
				},
			},
		}

	case "resources/read":
		var params ResourceReadParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			resp.Error = &Error{Code: InvalidParams, Message: "Invalid resource read params"}
			return resp
		}
		res, err := s.readResource(ctx, params.URI)
		if err != nil {
			resp.Error = &Error{Code: InternalError, Message: err.Error()}
			return resp
		}
		resp.Result = res

	default:
		resp.Error = &Error{
			Code:    MethodNotFound,
			Message: fmt.Sprintf("Method '%s' not found or unsupported", req.Method),
		}
	}

	return resp
}

func (s *Server) definedTools() []Tool {
	return []Tool{
		{
			Name:        "search_files",
			Description: "Search for files across all connected clouds (Google Drive, OneDrive, S3, Dropbox, etc.) using unified PostgreSQL index.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query": map[string]any{
						"type":        "string",
						"description": "File name or keyword to search for",
					},
					"account_id": map[string]any{
						"type":        "string",
						"description": "Optional account ID to filter results",
					},
					"limit": map[string]any{
						"type":        "integer",
						"description": "Max number of items to return (default 25)",
					},
				},
				"required": []string{"query"},
			},
		},
		{
			Name:        "list_files",
			Description: "List files and folders in a virtual directory across all connected clouds.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"folder_id": map[string]any{
						"type":        "string",
						"description": "Virtual folder ID to list (empty or null for root directory)",
					},
					"account_id": map[string]any{
						"type":        "string",
						"description": "Optional cloud account ID to filter",
					},
				},
			},
		},
		{
			Name:        "read_file",
			Description: "Download and read the content of a file from any cloud drive by its File ID.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"file_id": map[string]any{
						"type":        "string",
						"description": "The unique ID of the file to read",
					},
					"max_bytes": map[string]any{
						"type":        "integer",
						"description": "Maximum bytes to read (default 1048576 = 1MB)",
					},
				},
				"required": []string{"file_id"},
			},
		},
		{
			Name:        "upload_file",
			Description: "Upload a new text/document file to Poly Cloud. The smart routing engine automatically stores it on whichever connected cloud has the most free space.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{
						"type":        "string",
						"description": "File name including extension (e.g. summary.md, report.json, notes.txt)",
					},
					"content": map[string]any{
						"type":        "string",
						"description": "Content of the file as text",
					},
					"folder_id": map[string]any{
						"type":        "string",
						"description": "Optional virtual folder ID to place the file in",
					},
				},
				"required": []string{"name", "content"},
			},
		},
		{
			Name:        "get_storage_quota",
			Description: "Get overall storage quota and breakdown per cloud provider (Google Drive, OneDrive, S3, Dropbox, etc.).",
			InputSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
		{
			Name:        "list_accounts",
			Description: "List all connected cloud accounts and their operational status.",
			InputSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
		{
			Name:        "create_folder",
			Description: "Create a virtual folder in Poly Cloud.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{
						"type":        "string",
						"description": "Name of the folder to create",
					},
					"parent_id": map[string]any{
						"type":        "string",
						"description": "Optional parent folder ID",
					},
				},
				"required": []string{"name"},
			},
		},
		{
			Name:        "delete_file",
			Description: "Delete a file from the cloud storage by its File ID.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"file_id": map[string]any{
						"type":        "string",
						"description": "ID of the file to delete",
					},
				},
				"required": []string{"file_id"},
			},
		},
	}
}

func (s *Server) callTool(ctx context.Context, name string, args map[string]any) (ToolCallResult, error) {
	if args == nil {
		args = map[string]any{}
	}

	switch name {
	case "search_files":
		query, _ := args["query"].(string)
		if strings.TrimSpace(query) == "" {
			return ToolCallResult{}, errors.New("parameter 'query' is required")
		}
		accountID, _ := args["account_id"].(string)
		limit := 25
		if l, ok := args["limit"].(float64); ok && l > 0 {
			limit = int(l)
		}

		res, err := s.files.Search(ctx, s.userID, domain.SearchQuery{
			Q:         query,
			AccountID: accountID,
			PerPage:   limit,
			Page:      1,
		})
		if err != nil {
			return ToolCallResult{}, err
		}

		formatted, _ := json.MarshalIndent(map[string]any{
			"total_found": len(res.Items),
			"files":       res.Items,
		}, "", "  ")
		return ToolCallResult{Content: []TextContent{{Type: "text", Text: string(formatted)}}}, nil

	case "list_files":
		var folderID *string
		if fID, ok := args["folder_id"].(string); ok && strings.TrimSpace(fID) != "" {
			folderID = &fID
		}
		accountID, _ := args["account_id"].(string)

		fileRes, err := s.files.List(ctx, s.userID, folderID, domain.SearchQuery{
			AccountID: accountID,
			PerPage:   100,
			Page:      1,
		})
		if err != nil {
			return ToolCallResult{}, err
		}

		folders, err := s.folders.List(ctx, s.userID, folderID)
		if err != nil {
			return ToolCallResult{}, err
		}

		formatted, _ := json.MarshalIndent(map[string]any{
			"folder_id": folderID,
			"folders":   folders,
			"files":     fileRes.Items,
		}, "", "  ")
		return ToolCallResult{Content: []TextContent{{Type: "text", Text: string(formatted)}}}, nil

	case "read_file":
		fileID, _ := args["file_id"].(string)
		if strings.TrimSpace(fileID) == "" {
			return ToolCallResult{}, errors.New("parameter 'file_id' is required")
		}
		maxBytes := int64(1024 * 1024) // 1MB default
		if mb, ok := args["max_bytes"].(float64); ok && mb > 0 {
			maxBytes = int64(mb)
		}

		dl, err := s.files.Download(ctx, s.userID, fileID)
		if err != nil {
			return ToolCallResult{}, err
		}
		defer dl.Body.Close()

		lr := io.LimitReader(dl.Body, maxBytes+1)
		data, err := io.ReadAll(lr)
		if err != nil {
			return ToolCallResult{}, fmt.Errorf("failed reading file stream: %w", err)
		}

		truncated := false
		if int64(len(data)) > maxBytes {
			data = data[:maxBytes]
			truncated = true
		}

		if utf8.Valid(data) {
			text := string(data)
			if truncated {
				text += fmt.Sprintf("\n\n[... Truncated at %d bytes ...]", maxBytes)
			}
			return ToolCallResult{
				Content: []TextContent{{
					Type: "text",
					Text: fmt.Sprintf("File: %s\nMime: %s\nSize: %d bytes\n\n%s", dl.Name, dl.Mime, dl.SizeBytes, text),
				}},
			}, nil
		}

		// Binary content -> encode base64
		encoded := base64.StdEncoding.EncodeToString(data)
		return ToolCallResult{
			Content: []TextContent{{
				Type: "text",
				Text: fmt.Sprintf("File: %s\nMime: %s\nSize: %d bytes (binary data base64 encoded)\n\n%s", dl.Name, dl.Mime, dl.SizeBytes, encoded),
			}},
		}, nil

	case "upload_file":
		name, _ := args["name"].(string)
		content, _ := args["content"].(string)
		if strings.TrimSpace(name) == "" {
			return ToolCallResult{}, errors.New("parameter 'name' is required")
		}
		var folderID *string
		if fID, ok := args["folder_id"].(string); ok && strings.TrimSpace(fID) != "" {
			folderID = &fID
		}

		size := int64(len(content))
		res, err := s.files.Upload(ctx, s.userID, folderID, name, strings.NewReader(content), size, nil)
		if err != nil {
			return ToolCallResult{}, fmt.Errorf("upload failed: %w", err)
		}

		formatted, _ := json.MarshalIndent(map[string]any{
			"message":       "File successfully uploaded and indexed",
			"file_id":       res.File.ID,
			"name":          res.File.Name,
			"size_bytes":    res.File.SizeBytes,
			"account_id":    res.AccountID,
			"account_label": res.AccountLabel,
			"routed_by":     res.RoutedBy,
		}, "", "  ")
		return ToolCallResult{Content: []TextContent{{Type: "text", Text: string(formatted)}}}, nil

	case "get_storage_quota":
		rep, err := s.files.Quota(ctx, s.userID)
		if err != nil {
			return ToolCallResult{}, err
		}
		formatted, _ := json.MarshalIndent(rep, "", "  ")
		return ToolCallResult{Content: []TextContent{{Type: "text", Text: string(formatted)}}}, nil

	case "list_accounts":
		accs, err := s.accounts.List(ctx, s.userID)
		if err != nil {
			return ToolCallResult{}, err
		}
		formatted, _ := json.MarshalIndent(accs, "", "  ")
		return ToolCallResult{Content: []TextContent{{Type: "text", Text: string(formatted)}}}, nil

	case "create_folder":
		name, _ := args["name"].(string)
		if strings.TrimSpace(name) == "" {
			return ToolCallResult{}, errors.New("parameter 'name' is required")
		}
		var parentID *string
		if pID, ok := args["parent_id"].(string); ok && strings.TrimSpace(pID) != "" {
			parentID = &pID
		}

		f, err := s.folders.Create(ctx, s.userID, name, parentID)
		if err != nil {
			return ToolCallResult{}, err
		}
		formatted, _ := json.MarshalIndent(f, "", "  ")
		return ToolCallResult{Content: []TextContent{{Type: "text", Text: string(formatted)}}}, nil

	case "delete_file":
		fileID, _ := args["file_id"].(string)
		if strings.TrimSpace(fileID) == "" {
			return ToolCallResult{}, errors.New("parameter 'file_id' is required")
		}
		if err := s.files.Delete(ctx, s.userID, fileID); err != nil {
			return ToolCallResult{}, err
		}
		return ToolCallResult{Content: []TextContent{{Type: "text", Text: fmt.Sprintf("File %s successfully deleted.", fileID)}}}, nil

	default:
		return ToolCallResult{}, fmt.Errorf("unknown tool: %s", name)
	}
}

func (s *Server) readResource(ctx context.Context, uri string) (ResourceReadResult, error) {
	switch uri {
	case "polycloud://quota":
		rep, err := s.files.Quota(ctx, s.userID)
		if err != nil {
			return ResourceReadResult{}, err
		}
		data, _ := json.MarshalIndent(rep, "", "  ")
		return ResourceReadResult{
			Contents: []ResourceContents{{
				URI:      uri,
				MimeType: "application/json",
				Text:     string(data),
			}},
		}, nil

	case "polycloud://accounts":
		accs, err := s.accounts.List(ctx, s.userID)
		if err != nil {
			return ResourceReadResult{}, err
		}
		data, _ := json.MarshalIndent(accs, "", "  ")
		return ResourceReadResult{
			Contents: []ResourceContents{{
				URI:      uri,
				MimeType: "application/json",
				Text:     string(data),
			}},
		}, nil

	default:
		return ResourceReadResult{}, fmt.Errorf("unsupported resource URI: %s", uri)
	}
}

// ServeHTTP menangani request MCP lewat HTTP POST (/api/v1/mcp).
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed - Use POST for MCP JSON-RPC", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	resp := s.HandleMessage(r.Context(), body)
	if resp == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(resp)
}

// ServeStdio menjalankan transport loop stdio untuk Claude Desktop, Cursor, dll.
func (s *Server) ServeStdio(ctx context.Context, in io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(in)
	// Buffer besar untuk file upload / base64 payloads
	buf := make([]byte, 1024*1024)
	scanner.Buffer(buf, 10*1024*1024)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line := scanner.Bytes()
		if len(strings.TrimSpace(string(line))) == 0 {
			continue
		}

		resp := s.HandleMessage(ctx, line)
		if len(resp) > 0 {
			if _, err := out.Write(resp); err != nil {
				return err
			}
			if _, err := out.Write([]byte("\n")); err != nil {
				return err
			}
		}
	}
	return scanner.Err()
}
