package wsserver

import (
	"context"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const (
	templateDir = "./internal/web/templates/html"
)

type WSServer interface {
	Start() error
	Stop(ctx context.Context) error
}

type wsSrv struct {
	mux       *http.ServeMux
	srv       *http.Server
	wsUpg     *websocket.Upgrader
	wsClients map[*websocket.Conn]struct{}
	mu        sync.RWMutex
	broadcast chan *wsMwssage
	log       *zap.SugaredLogger
	shutdownCh chan struct{}
	stopOnce sync.Once
	readersWG sync.WaitGroup
	broadcastWG sync.WaitGroup
}

func NewWsServer(addr string, log *zap.SugaredLogger) WSServer {
	m := &http.ServeMux{}
	upg := &websocket.Upgrader{}

	return &wsSrv{
		mux: m,
		srv: &http.Server{
			Addr:    addr,
			Handler: m,
		},
		wsUpg:     upg,
		wsClients: map[*websocket.Conn]struct{}{},
		mu:        sync.RWMutex{},
		broadcast: make(chan *wsMwssage),
		log:       log,
		shutdownCh: make(chan struct{}),
	}
}

func (ws *wsSrv) Start() error {
	ws.mux.Handle("/", http.FileServer(http.Dir(templateDir)))
	ws.mux.HandleFunc("/ws", ws.wsHandler)
	ws.mux.HandleFunc("/test", ws.testHandler)

	ws.broadcastWG.Add(1)
	go ws.WriteToClientsBroadcast()

	return ws.srv.ListenAndServe()
}

func (ws *wsSrv) Stop(ctx context.Context) error {
	var stopErr error

	ws.stopOnce.Do(func() {
		close(ws.shutdownCh)

		if err := ws.srv.Shutdown(ctx); err != nil {
			stopErr = err
		}
		for _, conn := range ws.snapshotClients() {
			if err := conn.Close(); err != nil {
				ws.log.Errorf("close websocket conn: %v", err)
			}
		}

		ws.readersWG.Wait()
		close(ws.broadcast)

		ws.broadcastWG.Wait()
	})

	return stopErr
}

func (ws *wsSrv) testHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Test is successful"))
}

func (ws *wsSrv) wsHandler(w http.ResponseWriter, r *http.Request) {
	select {
	case <-ws.shutdownCh:
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	default:
	}

	conn, err := ws.wsUpg.Upgrade(w, r, nil)
	if err != nil {
		ws.log.Error("upgared ws conn: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	select {
	case <-ws.shutdownCh:
		conn.Close()
		return
	default:
	}

	ws.AddConn(conn)
	ws.log.Infof("Client with address %s connected", conn.RemoteAddr().String())

	ws.readersWG.Add(1)
	go ws.ReadFromClient(conn)
}

func (ws *wsSrv) AddConn(conn *websocket.Conn) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	ws.wsClients[conn] = struct{}{}
}

func (ws *wsSrv) RemoveConn(conn *websocket.Conn) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	delete(ws.wsClients, conn)
}

func (ws *wsSrv) ReadFromClient(conn *websocket.Conn) {
	defer ws.readersWG.Done()
	defer conn.Close()
	defer ws.RemoveConn(conn)

	for {
		msg := new(wsMwssage)
		if err := conn.ReadJSON(msg); err != nil {
			select {
			case <-ws.shutdownCh:
			default:
				ws.log.Errorf("reading from websocket: %v", err)
			}
			break
		}

		host, _, err := net.SplitHostPort(conn.RemoteAddr().String())
		if err != nil {
			ws.log.Errorf("conn address split: %v", err)
		}
		msg.IPAddress = host
		msg.Time = time.Now().Format("15:04")

		select {
		case <-ws.shutdownCh:
			return
		case ws.broadcast <- msg:
		}
	}
}

func (ws *wsSrv) WriteToClientsBroadcast() {
	defer ws.broadcastWG.Done()

	for msg := range ws.broadcast {
		for _, client := range ws.snapshotClients() {
			if err := client.WriteJSON(msg); err != nil {
				ws.log.Errorf("writing message: %v", err)
				client.Close()
				ws.RemoveConn(client)
			}
		}
	}
}

func (ws *wsSrv) snapshotClients() []*websocket.Conn {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	clients := make([]*websocket.Conn, 0, len(ws.wsClients))
	for conn := range ws.wsClients {
		clients = append(clients, conn)
	}

	return clients
}
