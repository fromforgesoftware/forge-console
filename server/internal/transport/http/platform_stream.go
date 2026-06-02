package http

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	gws "github.com/gorilla/websocket"

	kitrest "github.com/fromforgesoftware/go-kit/transport/rest"
	kitws "github.com/fromforgesoftware/go-kit/transport/websocket"

	"github.com/fromforgesoftware/forge/server/internal/api"
	"github.com/fromforgesoftware/forge/server/internal/app"
)

const (
	topicTopology kitws.TopicType = "topology"
	topicLogs     kitws.TopicType = "logs"
)

// PlatformStreamController serves the live topology + pod-log websocket at
// /api/platform/stream. It reuses the kit websocket envelope; topology pushes
// are informer-driven, logs are streamed from the kubernetes pods/log API.
type PlatformStreamController struct {
	topo     app.TopologyUsecase
	watcher  app.TopologyWatcher
	logs     app.LogStreamer
	auth     app.AuthUsecase
	authz    app.AuthzUsecase
	upgrader gws.Upgrader
}

func NewPlatformStreamController(topo app.TopologyUsecase, watcher app.TopologyWatcher, logs app.LogStreamer, auth app.AuthUsecase, authz app.AuthzUsecase) kitrest.Controller {
	return &PlatformStreamController{
		topo:     topo,
		watcher:  watcher,
		logs:     logs,
		auth:     auth,
		authz:    authz,
		upgrader: gws.Upgrader{CheckOrigin: func(*http.Request) bool { return true }},
	}
}

func (c *PlatformStreamController) Routes(r kitrest.Router) {
	r.Method(http.MethodGet, "/api/platform/stream", http.HandlerFunc(c.stream))
}

func (c *PlatformStreamController) stream(w http.ResponseWriter, r *http.Request) {
	u, ok := resolveUser(w, r, c.auth)
	if !ok {
		return
	}
	allowed, err := c.authz.Can(r.Context(), app.SubjectTypeUser, u.ID(), permPlatformRead)
	if err != nil {
		writeErr(w, err)
		return
	}
	if !allowed {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return
	}
	conn, err := c.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	newPlatformSession(conn, c.topo, c.watcher, c.logs).run(r.Context())
}

type platformSession struct {
	conn    *gws.Conn
	topo    app.TopologyUsecase
	watcher app.TopologyWatcher
	logs    app.LogStreamer
	out     chan kitws.Message
	seq     int64

	mu      sync.Mutex
	cancels map[string]func()
}

func newPlatformSession(conn *gws.Conn, topo app.TopologyUsecase, watcher app.TopologyWatcher, logs app.LogStreamer) *platformSession {
	return &platformSession{
		conn:    conn,
		topo:    topo,
		watcher: watcher,
		logs:    logs,
		out:     make(chan kitws.Message, 64),
		cancels: map[string]func(){},
	}
}

func (s *platformSession) run(parent context.Context) {
	ctx, cancel := context.WithCancel(parent)
	defer cancel()
	go s.writePump(ctx)

	s.conn.SetReadLimit(1 << 20)
	for {
		_, data, err := s.conn.ReadMessage()
		if err != nil {
			break
		}
		var m kitws.Message
		if json.Unmarshal(data, &m) != nil {
			continue
		}
		s.handle(ctx, m)
	}
	s.cancelAll()
}

func (s *platformSession) handle(ctx context.Context, m kitws.Message) {
	switch m.Type {
	case kitws.MessageTypeSubscribe:
		switch m.Topic {
		case topicTopology:
			s.startSub(ctx, m.Subject, func(c context.Context) { s.pushTopology(c, m.Subject) })
		case topicLogs:
			ns, pod, container := logTarget(m.Data)
			if pod == "" {
				return
			}
			s.startSub(ctx, m.Subject, func(c context.Context) { s.pushLogs(c, m.Subject, ns, pod, container) })
		}
		s.send(kitws.Message{Type: kitws.MessageTypeAck, Topic: m.Topic, Subject: m.Subject})
	default:
		s.stopSub(m.Subject)
	}
}

func (s *platformSession) startSub(ctx context.Context, id string, fn func(context.Context)) {
	subCtx, cancel := context.WithCancel(ctx)
	s.mu.Lock()
	if old, ok := s.cancels[id]; ok {
		old()
	}
	s.cancels[id] = cancel
	s.mu.Unlock()
	go func() {
		defer cancel()
		fn(subCtx)
	}()
}

func (s *platformSession) stopSub(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if cancel, ok := s.cancels[id]; ok {
		cancel()
		delete(s.cancels, id)
	}
}

func (s *platformSession) cancelAll() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, cancel := range s.cancels {
		cancel()
		delete(s.cancels, id)
	}
}

func (s *platformSession) pushTopology(ctx context.Context, id string) {
	ticks, unsub := s.watcher.Subscribe()
	defer unsub()
	s.sendTopology(ctx, id)
	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-ticks:
			if !ok {
				return
			}
			s.sendTopology(ctx, id)
		}
	}
}

func (s *platformSession) sendTopology(ctx context.Context, id string) {
	t, err := s.topo.Get(ctx)
	if err != nil {
		return
	}
	s.send(kitws.Message{Type: kitws.MessageTypeMessage, Topic: topicTopology, Subject: id, Data: api.TopologyToPayload(t)})
}

func (s *platformSession) pushLogs(ctx context.Context, id, ns, pod, container string) {
	rc, err := s.logs.OpenPodLogs(ctx, ns, pod, container)
	if err != nil {
		s.send(kitws.Message{Type: kitws.MessageTypeError, Topic: topicLogs, Subject: id, Data: map[string]string{"error": err.Error()}})
		return
	}
	defer func() { _ = rc.Close() }()
	go func() {
		<-ctx.Done()
		_ = rc.Close()
	}()
	scanner := bufio.NewScanner(rc)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		if ctx.Err() != nil {
			return
		}
		s.send(kitws.Message{Type: kitws.MessageTypeMessage, Topic: topicLogs, Subject: id, Data: map[string]string{"line": scanner.Text()}})
	}
}

func (s *platformSession) send(m kitws.Message) {
	m.SequenceNumber = atomic.AddInt64(&s.seq, 1)
	select {
	case s.out <- m:
	default:
	}
}

func (s *platformSession) writePump(ctx context.Context) {
	ping := time.NewTicker(30 * time.Second)
	defer ping.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case m := <-s.out:
			_ = s.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := s.conn.WriteJSON(m); err != nil {
				return
			}
		case <-ping.C:
			_ = s.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := s.conn.WriteMessage(gws.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func logTarget(data any) (ns, pod, container string) {
	m, ok := data.(map[string]any)
	if !ok {
		return "", "", ""
	}
	str := func(k string) string {
		if v, ok := m[k].(string); ok {
			return v
		}
		return ""
	}
	return str("namespace"), str("pod"), str("container")
}
