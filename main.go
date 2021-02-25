// Copyright 2015 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type AuthenticationInput struct {
	Bearer   string `json:"bearer"`
	UserUUID string `json:"user_uuid"`
}

type EventMessage struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

type UserConnection struct {
	UserID        string
	Authenticated bool
	Conn          *websocket.Conn
	ChallengeSent bool
}

var addr = flag.String("addr", "localhost:8080", "http service address")

var upgrader = websocket.Upgrader{} // use default options

var shared *websocket.Conn
var sharedUsers map[string]UserConnection

func verifyTokenWithLearnalist(bearer string, userUUID string) ([]byte, error) {

	// https://github.com/freshteapot/learnalist-api/blob/master/docs/api.user.info.md
	url := fmt.Sprintf("https://learnalist.net/api/v1/user/info/%s", userUUID)

	spaceClient := http.Client{
		Timeout: time.Millisecond * 500,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", bearer))

	res, getErr := spaceClient.Do(req)
	if getErr != nil {
		return nil, getErr
	}

	if res.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("%d", res.StatusCode))
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		return nil, readErr
	}

	return body, nil
}

func echo(w http.ResponseWriter, r *http.Request) {
	userConnection := UserConnection{}
	c, err := upgrader.Upgrade(w, r, nil)
	userConnection.Conn = c

	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	shared = c
	defer c.Close()
	for {
		if !userConnection.ChallengeSent {
			clientEvent := EventMessage{
				Type: "authenticate",
				Data: "123",
			}
			wsResponse, _ := json.Marshal(clientEvent)
			shared.WriteMessage(1, wsResponse)
			// TODO could trigger a timer to kill the connection if the user hasn't responded (it should be almost instant)
			userConnection.ChallengeSent = true
			continue
		}

		mt, message, err := userConnection.Conn.ReadMessage()
		fmt.Println(mt)
		if err != nil {
			log.Println("read readmessage:", err)
			break
		}

		// 1) Add a handshake
		// 2) Ask for the user to authenticate
		// 3) Lookup user session from Bearer token
		// 4) if next message is not bearer token, close connection

		log.Printf("recv: %s", message)
		var clientEvent EventMessage
		err = json.Unmarshal(message, &clientEvent)

		if !userConnection.Authenticated {
			if clientEvent.Type != "authenticate" {
				userConnection.Conn.Close()
				break
			}

			var input AuthenticationInput
			err := json.Unmarshal([]byte(clientEvent.Data), &input)
			if err != nil {
				fmt.Println("error getting AuthenticationInput", err)
				userConnection.Conn.Close()
				break
			}

			data, err := verifyTokenWithLearnalist(input.Bearer, input.UserUUID)
			if err != nil {
				fmt.Println("error verifyToken", err)
				userConnection.Conn.Close()
				break
			}
			fmt.Println(string(data))

			//
			userConnection.Authenticated = true
			clientEvent := EventMessage{
				Type: "authenticated",
				Data: string(data),
			}
			wsResponse, _ := json.Marshal(clientEvent)
			userConnection.Conn.WriteMessage(1, wsResponse)
			continue
		}

		if err != nil {
			log.Println("read json:", err)
			userConnection.Conn.Close()
			break
		}
		clientEvent.Type = "chat"

		wsResponse, _ := json.Marshal(clientEvent)

		log.Printf("recv: %s", wsResponse)
		err = userConnection.Conn.WriteMessage(mt, wsResponse)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	homeTemplate.Execute(w, "ws://"+r.Host+"/echo")
}

func update(w http.ResponseWriter, r *http.Request) {
	clientEvent := EventMessage{
		Type: "update",
		Data: "I am an update",
	}
	wsResponse, _ := json.Marshal(clientEvent)
	shared.WriteMessage(1, wsResponse)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/echo", echo)
	http.HandleFunc("/update", update)
	http.Handle("/", http.FileServer(http.Dir("./public")))

	log.Fatal(http.ListenAndServe(*addr, nil))
}

var homeTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<script>
window.addEventListener("load", function(evt) {

    var output = document.getElementById("output");
    var input = document.getElementById("input");
    var ws;

    var print = function(message) {
        var d = document.createElement("div");
        d.textContent = message;
        output.appendChild(d);
    };

    document.getElementById("open").onclick = function(evt) {
        if (ws) {
            return false;
        }
        ws = new WebSocket("{{.}}");
        ws.onopen = function(evt) {
            print("OPEN");
        }
        ws.onclose = function(evt) {
            print("CLOSE");
            ws = null;
        }
        ws.onmessage = function(evt) {
            print("RESPONSE: " + evt.data);
        }
        ws.onerror = function(evt) {
            print("ERROR: " + evt.data);
        }
        return false;
    };

    document.getElementById("send").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        print("SEND: " + input.value);
        ws.send(input.value);
        return false;
    };

    document.getElementById("close").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        ws.close();
        return false;
    };

});
</script>
</head>
<body>
<table>
<tr><td valign="top" width="50%">
<p>Click "Open" to create a connection to the server,
"Send" to send a message to the server and "Close" to close the connection.
You can change the message and send multiple times.
<p>
<form>
<button id="open">Open</button>
<button id="close">Close</button>
<p><input id="input" type="text" value="Hello world!">
<button id="send">Send</button>
</form>
</td><td valign="top" width="50%">
<div id="output"></div>
</td></tr></table>
</body>
</html>
`))
