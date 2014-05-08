package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"code.google.com/p/go.net/websocket"
)

func root(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.Header().Add("Access-Control-Allow-Origin","*")
		return
	}
	broadCast("Update")
}

var clients map[*websocket.Conn]struct{} = make(map[*websocket.Conn]struct{})
var p *string
var ping string = "ping"

var script = `
	(function(){
		var ws = new WebSocket("ws://%v/ws");
		ws.onmessage = function(d) {
			if(JSON.parse(d.data) == "Update") {
				window.location.reload();
			}
		}
	})();
`

func keepAlive(c *websocket.Conn) {
	for {
		c.SetDeadline(time.Now().Add(10 * time.Second))
		err := websocket.JSON.Send(c, ping)
		if err != nil {
			delete(clients, c)
			break
		}
		time.Sleep(8 * time.Second)
	}
}

func broadCast(msg string) {
	for c := range clients {
		websocket.JSON.Send(c, msg)
	}
}

func getScript(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, script, server)
}

func ws(c *websocket.Conn) {
	clients[c] = struct{}{}
	keepAlive(c)
}

var server string

func main() {
	p = flag.String("port", ":8080", "-port=\":8080\"")
	flag.Parse()
	server = fmt.Sprintf("localhost%v", *p)
	_, err := http.Get("http://" + server)
	if err != nil {
		http.Handle("/ws", websocket.Handler(ws))
		http.HandleFunc("/script", getScript)
		http.HandleFunc("/", root)
		http.ListenAndServe(server, nil)
	}
}
