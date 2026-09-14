// Command mcp adalah entrypoint MCP (Model Context Protocol) untuk Poly Cloud.
// Berjalan via stdio transport (stdin/stdout) untuk digunakan oleh Claude Desktop,
// Cursor, Antigravity, VS Code, dan AI agents lainnya.
package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/polycloud/platform/apps/backend/internal/mcp"
)

func main() {
	defaultURL := os.Getenv("POLYCLOUD_API_URL")
	if defaultURL == "" {
		defaultURL = "http://localhost:8080"
	}

	apiURLFlag := flag.String("api-url", defaultURL, "Poly Cloud backend API base URL")
	flag.Parse()

	apiURL := strings.TrimRight(*apiURLFlag, "/")
	mcpEndpoint := apiURL + "/api/v1/mcp"

	fmt.Fprintf(os.Stderr, "[poly-cloud-mcp] Starting MCP stdio bridge -> %s\n", mcpEndpoint)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	httpClient := &http.Client{
		Timeout: 2 * time.Minute, // Timeout cukup longgar untuk download/upload berkas
	}

	scanner := bufio.NewScanner(os.Stdin)
	// Buffer sampai 10MB untuk menangani upload berkas dan payload besar
	buf := make([]byte, 1024*1024)
	scanner.Buffer(buf, 10*1024*1024)

	localServer := mcp.NewServer(nil, nil, nil, nil, nil)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
		}

		line := scanner.Bytes()
		trimmed := bytes.TrimSpace(line)
		if len(trimmed) == 0 {
			continue
		}

		var rawReq map[string]any
		if err := json.Unmarshal(trimmed, &rawReq); err != nil {
			sendError(nil, -32700, "Parse error: "+err.Error())
			continue
		}

		reqID := rawReq["id"]
		method, _ := rawReq["method"].(string)

		// Notifikasi initialized dari MCP client tidak memerlukan balasan
		if reqID == nil && strings.HasPrefix(method, "notifications/") {
			continue
		}

		// Coba kirim ke Poly Cloud backend API jika aktif
		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, mcpEndpoint, bytes.NewReader(trimmed))
		var httpResp *http.Response
		if err == nil {
			httpReq.Header.Set("Content-Type", "application/json")
			httpResp, err = httpClient.Do(httpReq)
		}

		if err == nil && httpResp != nil && httpResp.StatusCode == http.StatusOK {
			body, readErr := io.ReadAll(httpResp.Body)
			_ = httpResp.Body.Close()
			if readErr == nil && len(bytes.TrimSpace(body)) > 0 {
				os.Stdout.Write(body)
				os.Stdout.Write([]byte("\n"))
				continue
			}
		}
		if httpResp != nil {
			_ = httpResp.Body.Close()
		}

		// Fallback: jika initialize, tools/list, atau ping, server lokal bisa langsung merespons
		// sehingga Claude Desktop / Cursor langsung mengenali tools tanpa gagal startup.
		if method == "initialize" || method == "tools/list" || method == "ping" {
			localResp := localServer.HandleMessage(ctx, trimmed)
			if len(localResp) > 0 {
				os.Stdout.Write(localResp)
				os.Stdout.Write([]byte("\n"))
				continue
			}
		}

		// Jika pemanggilan tools gagal karena backend mati, berikan pesan ramah
		sendError(reqID, -32603, fmt.Sprintf("Poly Cloud backend is not reachable at %s. Please start Poly Cloud with 'make dev' or 'docker compose up' to execute file operations.", apiURL))
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		fmt.Fprintf(os.Stderr, "[poly-cloud-mcp] Stdin read error: %v\n", err)
	}
}

func sendError(id any, code int, message string) {
	resp := map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"error": map[string]any{
			"code":    code,
			"message": message,
		},
	}
	out, _ := json.Marshal(resp)
	os.Stdout.Write(out)
	os.Stdout.Write([]byte("\n"))
}
