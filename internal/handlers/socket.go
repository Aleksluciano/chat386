package handlers

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sort"
)

var wsChan = make(chan WsPayload)
var clients = make(map[WebSocketConnection]string)

var upgradeConnection = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebSocketConnection struct {
	*websocket.Conn
}

//Defines the response sent back from webscoket
type WsJsonResponse struct {
	Action         string   `json:"action"`
	Username       string   `json:"username"`
	Message        string   `json:"message"`
	MessageType    string   `json:"message_type"`
	Removed        string `json:"removed"`
	ConnectedUsers []string `json:"connected_users"`
}

type WsPayload struct {
	Action   string              `json:"action"`
	Username string              `json:"username"`
	Message  string              `json:"message"`
	Conn     WebSocketConnection `json:"-"`
}

//Upgrades connection to websocket
func WsEndpoint(w http.ResponseWriter, r *http.Request) {
	ws, err := upgradeConnection.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}

	log.Println("Client connect to endpoint")

	var response WsJsonResponse
	response.Message = `<em><small>Connect to server</small></em>`

	conn := WebSocketConnection{Conn: ws}
	clients[conn] = ""

	err = ws.WriteJSON(response)
	if err != nil {
		log.Println(err)
	}

	go ListenForWs(&conn)

}

func ListenForWs(conn *WebSocketConnection) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Error", fmt.Sprintf("%v", r))
		}
	}()

	var payload WsPayload
	for {
		err := conn.ReadJSON(&payload)
		if err != nil {
			//do nothing
		} else {
			payload.Conn = *conn
			wsChan <- payload
		}
	}
}

func ListenToWsChannel() {
	var response WsJsonResponse
	for {
		e := <-wsChan
		response = WsJsonResponse{}
		switch e.Action {
		case "username":
			//get list of user users and send it back via broadcast
			clients[e.Conn] = e.Username
			users := getUserList()
			response.Action = "list_users"
			response.ConnectedUsers = users
			broadcastToAll(response, e.Conn)
		case "left":
			response.Action = "list_users"
			delete(clients, e.Conn)
			users := getUserList()
			response.ConnectedUsers = users
			response.Removed = e.Username
			broadcastToAll(response, e.Conn)
			fmt.Printf("Client %v saiu", e.Conn)
		case "broadcast":
			response.Action = "broadcast"
			response.Username = e.Username
			//response.Message = fmt.Sprintf("<strong>%s<strong>: %s", e.Username, e.Message)
			response.Message = fmt.Sprintf("%s", e.Message)
			broadcastToAll(response, e.Conn)

		}
		fmt.Printf("%v falando com server", e.Conn)
		fmt.Printf("%v falando com server", response)
		//response.Action = "Got here"
		//response.Message = fmt.Sprintf("Some message, and action was %s", e.Action)
		//broadcastToAll(response)
	}
}

func getUserList() []string {
	var userList []string
	for _, x := range clients {
		if x != "" {
			userList = append(userList, x)
		}
	}
	sort.Strings(userList)
	return userList

}

func broadcastToAll(response WsJsonResponse, user WebSocketConnection) {
	for client := range clients {
		//if ( client.Conn != user.Conn ) || ( client.Conn == user.Conn && response.Action == "list_users" ) {
			err := client.WriteJSON(response)
			if err != nil {
				log.Println("websocket err")
				_ = client.Close()
				delete(clients, client)
			}
		//}
	}
}
