package digest

//
//import (
//	"github.com/gorilla/websocket"
//	"net/http"
//	"strings"
//)
//
//type Subscriber struct {
//	conn   *websocket.Conn
//	events []string
//}
//
//var clients = make(map[*websocket.Conn]bool)
//var subscribers = make(map[*websocket.Conn]*Subscriber)
//
//var upgrader = websocket.Upgrader{
//	CheckOrigin: func(r *http.Request) bool {
//		return true
//	},
//}
//
//func (hd *Handlers) handleConnections(w http.ResponseWriter, r *http.Request) {
//	ws, err := upgrader.Upgrade(w, r, nil)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer ws.Close()
//
//	clients[ws] = true
//
//	for {
//		_, msg, err := ws.ReadMessage()
//		if err != nil {
//			delete(clients, ws)
//			delete(subscribers, ws)
//			break
//		}
//
//		messageParts := strings.Split(string(msg), ":")
//		if len(messageParts) == 2 && messageParts[0] == "subscribe" {
//			eventName := messageParts[1]
//			if sub, ok := subscribers[ws]; ok {
//				sub.events = append(sub.events, eventName)
//			} else {
//				subscribers[ws] = &Subscriber{
//					conn:   ws,
//					events: []string{eventName},
//				}
//			}
//		}
//	}
//}
//
//func emitEvent(eventName string, message []byte) {
//	for _, sub := range subscribers {
//		for _, subEvent := range sub.events {
//			if subEvent == eventName {
//				err := sub.conn.WriteMessage(websocket.TextMessage, message)
//				if err != nil {
//					log.Printf("Websocket error: %v", err)
//					sub.conn.Close()
//					delete(clients, sub.conn)
//					delete(subscribers, sub.conn)
//				}
//			}
//		}
//	}
//}
