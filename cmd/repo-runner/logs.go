package main

import (
	"bytes"
	"encoding/json"
	"html/template"
	"io"
	"log"
	"net/http"
	"path"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/hpcloud/tail"
)

const (
	logTypeDefault      = "log"
	logTypeMetaStart    = "meta_start"
	logTypeMetaFinished = "meta_finished"
	logTypeMetaRepoName = "meta_repo-name"
	logTypeMetaRepoURL  = "meta_repo-url"
)

type logLine struct {
	Type    string    `json:"type"`
	Channel string    `json:"channel"`
	Time    time.Time `json:"time"`
	Message string    `json:"message"`
}

type logWriter struct {
	channel string
	file    io.Writer
	buf     *bytes.Buffer
}

func newLogWriter(channel string, persister io.Writer) *logWriter {
	return &logWriter{
		channel: channel,
		file:    persister,
		buf:     bytes.NewBuffer([]byte{}),
	}
}

func (l *logWriter) MetaMessage(messageType string, message string) error {
	ll := logLine{
		Type:    messageType,
		Channel: l.channel,
		Time:    time.Now(),
		Message: message,
	}

	return json.NewEncoder(l.file).Encode(ll)
}

func (l *logWriter) Write(p []byte) (n int, err error) {
	n, err = l.buf.Write(p)
	if err != nil {
		return
	}

	for bytes.Contains(l.buf.Bytes(), []byte{'\n'}) {
		parts := bytes.SplitN(l.buf.Bytes(), []byte{'\n'}, 2)

		if err = l.MetaMessage(logTypeDefault, string(parts[0])); err != nil {
			return
		}

		l.buf.Reset()
		if _, err = l.buf.Write(parts[1]); err != nil {
			return
		}
	}

	return
}

func registerLogHandlers(r *mux.Router) {
	r.HandleFunc("/{logID}", handleLogInterface)
	r.HandleFunc("/stream/{logID}", handleLogStream)
}

func handleLogInterface(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	logID := vars["logID"]

	tplRaw, err := Asset("assets/loginterface.html")
	if err != nil {
		log.Printf("[ERRO] loginterface.html not found in binary.")
		return
	}

	tpl, err := template.New("loginterface").Parse(string(tplRaw))
	if err != nil {
		log.Printf("[ERRO] Could not parse loginterface.html: %s", err)
		return
	}

	tpl.Execute(w, map[string]interface{}{
		"LogID": logID,
	})
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Maximum message size allowed from peer.
	maxMessageSize = 8192

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Number of concurrent calculations per socket
	calculationPoolSize = 3
)

var upgrader = websocket.Upgrader{}

func handleLogStream(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	logID := vars["logID"]

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Could not open socket", http.StatusInternalServerError)
		return
	}
	defer ws.Close()

	doneChan := make(chan struct{})
	defer close(doneChan)
	go pingSocket(ws, doneChan)

	ws.SetReadLimit(maxMessageSize)
	ws.SetReadDeadline(time.Now().Add(pongWait))
	ws.SetPongHandler(func(string) error { ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	logSource, err := tail.TailFile(path.Join(cfg.LogDir, logID+".jsonl"), tail.Config{MustExist: true, Follow: true})
	if err != nil {
		ws.WriteJSON(logLine{
			Channel: "stderr",
			Message: "Log with ID " + logID + " not found.",
		})
		return
	}

	for line := range logSource.Lines {
		msg := []byte(line.Text)
		ws.WriteMessage(websocket.TextMessage, msg)

		ll := logLine{}
		if err := json.Unmarshal(msg, &ll); err != nil {
			return
		}
		if ll.Type == logTypeMetaFinished {
			break
		}
	}
}

func pingSocket(ws *websocket.Conn, done chan struct{}) {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ws.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(writeWait))
		case <-done:
			return
		}
	}
}
