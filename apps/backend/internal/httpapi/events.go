package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// UploadEvent = satu pesan progres upload yang dikirim lewat SSE (doc 06 §Events).
type UploadEvent struct {
	Name string `json:"-"` // nama event SSE: progress | done | error
	Data any    `json:"-"`
}

// Hub menyiarkan progres job upload ke subscriber SSE. Job berumur pendek;
// buffer per-job menahan event yang terbit sebelum klien sempat connect.
type Hub struct {
	mu   sync.Mutex
	jobs map[string]*job
}

type job struct {
	mu        sync.Mutex
	buffered  []UploadEvent
	listeners []chan UploadEvent
	done      bool
	expires   time.Time
}

func NewHub() *Hub {
	h := &Hub{jobs: make(map[string]*job)}
	go h.reap()
	return h
}

// reap membersihkan job yang sudah selesai dan lewat masa simpan.
func (h *Hub) reap() {
	for range time.Tick(time.Minute) {
		now := time.Now()
		h.mu.Lock()
		for id, j := range h.jobs {
			j.mu.Lock()
			expired := j.done && now.After(j.expires)
			j.mu.Unlock()
			if expired {
				delete(h.jobs, id)
			}
		}
		h.mu.Unlock()
	}
}

func (h *Hub) get(jobID string) *job {
	h.mu.Lock()
	defer h.mu.Unlock()
	j, ok := h.jobs[jobID]
	if !ok {
		j = &job{expires: time.Now().Add(10 * time.Minute)}
		h.jobs[jobID] = j
	}
	return j
}

// Publish menyiarkan satu event ke semua listener job.
func (h *Hub) Publish(jobID, name string, data any) {
	j := h.get(jobID)
	ev := UploadEvent{Name: name, Data: data}

	j.mu.Lock()
	defer j.mu.Unlock()
	if len(j.buffered) < 256 {
		j.buffered = append(j.buffered, ev)
	}
	for _, ch := range j.listeners {
		select {
		case ch <- ev:
		default: // listener lambat: lewati, progres berikutnya tetap menyusul
		}
	}
	if name == "done" || name == "error" {
		j.done = true
		j.expires = time.Now().Add(2 * time.Minute)
		for _, ch := range j.listeners {
			close(ch)
		}
		j.listeners = nil
	}
}

// subscribe mengembalikan channel event + snapshot event yang sudah lewat.
func (h *Hub) subscribe(jobID string) (<-chan UploadEvent, []UploadEvent, bool) {
	j := h.get(jobID)
	j.mu.Lock()
	defer j.mu.Unlock()

	backlog := append([]UploadEvent(nil), j.buffered...)
	if j.done {
		return nil, backlog, false
	}
	ch := make(chan UploadEvent, 32)
	j.listeners = append(j.listeners, ch)
	return ch, backlog, true
}

// HandleUploadEvents menstream progres job ke frontend.
func (h *Hub) HandleUploadEvents(w http.ResponseWriter, r *http.Request) {
	jobID := r.PathValue("jobId")
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming tidak didukung", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	ch, backlog, live := h.subscribe(jobID)
	for _, ev := range backlog {
		writeSSE(w, ev)
	}
	flusher.Flush()
	if !live {
		return
	}

	// Heartbeat mencegah proxy memutus koneksi yang lama sunyi.
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case ev, open := <-ch:
			if !open {
				return
			}
			writeSSE(w, ev)
			flusher.Flush()
		case <-ticker.C:
			fmt.Fprint(w, ": keepalive\n\n")
			flusher.Flush()
		}
	}
}

func writeSSE(w http.ResponseWriter, ev UploadEvent) {
	payload, err := json.Marshal(ev.Data)
	if err != nil {
		return
	}
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", ev.Name, payload)
}
