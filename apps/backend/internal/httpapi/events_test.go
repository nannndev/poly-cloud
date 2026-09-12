package httpapi

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Klien yang connect setelah upload selesai tetap harus menerima riwayat event —
// kalau tidak, progress bar menggantung selamanya.
func TestHubMengirimBacklogKeKlienTerlambat(t *testing.T) {
	hub := NewHub()
	hub.Publish("job-1", "progress", map[string]any{"bytes": 512, "total": 1024})
	hub.Publish("job-1", "done", map[string]any{"file_id": "f-1"})

	req := httptest.NewRequest("GET", "/api/v1/events/uploads/job-1", nil)
	req.SetPathValue("jobId", "job-1")
	w := httptest.NewRecorder()

	hub.HandleUploadEvents(w, req)

	body := w.Body.String()
	if !strings.Contains(body, "event: progress") || !strings.Contains(body, "event: done") {
		t.Fatalf("backlog tak terkirim: %q", body)
	}
	if ct := w.Header().Get("Content-Type"); ct != "text/event-stream" {
		t.Errorf("Content-Type = %q", ct)
	}
}

func TestHubMenstreamEventLangsung(t *testing.T) {
	hub := NewHub()
	req := httptest.NewRequest("GET", "/api/v1/events/uploads/job-2", nil)
	req.SetPathValue("jobId", "job-2")
	w := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		hub.HandleUploadEvents(w, req)
		close(done)
	}()

	// Beri kesempatan handler mendaftar sebagai listener sebelum event terbit.
	time.Sleep(50 * time.Millisecond)
	hub.Publish("job-2", "progress", map[string]any{"bytes": 100})
	hub.Publish("job-2", "done", map[string]any{"file_id": "f-2"})

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("handler tak berhenti setelah event done")
	}
	if !strings.Contains(w.Body.String(), `"bytes":100`) {
		t.Fatalf("event progres tak terkirim: %q", w.Body.String())
	}
}

// Event "error" juga harus menutup stream, bukan hanya "done".
func TestHubMenutupStreamSaatError(t *testing.T) {
	hub := NewHub()
	hub.Publish("job-3", "error", map[string]any{"message": "NO_ROOM"})

	req := httptest.NewRequest("GET", "/api/v1/events/uploads/job-3", nil)
	req.SetPathValue("jobId", "job-3")
	w := httptest.NewRecorder()

	hub.HandleUploadEvents(w, req)

	if !strings.Contains(w.Body.String(), "event: error") {
		t.Fatalf("event error tak terkirim: %q", w.Body.String())
	}
}
