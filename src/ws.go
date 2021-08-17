package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type Message struct {
	AuthorID  string    `json:"author_id"`
	AuthorName	  string    `json:"author_name"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type User struct {
	ID     string          `json:"id"`
	Token  string          `json:"token"`
	Name   string          `json:"name"`
	Socket *websocket.Conn `json:"socket"`
}

type Packet struct {
	Code string `json:"c"`
	Data Map    `json:"d"`
}

var Groups = make(map[string]*Group)
var packetQueue = make(chan *Packet)
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var (
	PacketPacket = "packet_error"
	PacketSendMessage         = "send_message"
	PacketAuthenticationError = "auth_error"

	PacketReceiveMessage  = "receive_message"
	PacketSetMessageError = "set_message_error"
	PacketClose = "close_connection"
	PacketSetUserData = "set_user_data"
)

func ManageWebsocketConnections(groupID string, setOwner bool, setOwnerID string, memberName string, w http.ResponseWriter, r *http.Request) interface{} {
	socket, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return false
	}

	if len(groupID) <= 0 {
		socket.WriteJSON(&Packet{
			Code: PacketAuthenticationError,
			Data: Map{
				"message": "Invalid group ID",
			},
		})

		return false
	}

	group := Groups[groupID]
	if group == nil {
		socket.WriteJSON(&Packet{
			Code: PacketAuthenticationError,
			Data: Map{
				"message": "The group that you try to connect doesn't exist",
			},
		})

		return false
	}

	var member *User
	if setOwner && len(setOwnerID) >= 1 {
		if strings.Compare(setOwnerID, group.OwnerID) != 0 {
			socket.WriteJSON(&Packet{
				Code: PacketAuthenticationError,
				Data: Map{
					"message": "The owner ID is invalid",
				},
			})

			return false
		}

		idAlreadyUsed := false
		for _, member := range group.Members {
			if member.ID == setOwnerID {
				idAlreadyUsed = true
				break
			}
		}

		if idAlreadyUsed {
			socket.WriteJSON(&Packet{
				Code: PacketAuthenticationError,
				Data: Map{
					"message": "The owner ID has been already claimed",
				},
			})

			return false
		}

		member = &User{
			ID:    setOwnerID,
			Name: memberName,
			Token: RandomToken(40),
			Socket: socket,
		}

		group.OwnerID = setOwnerID
		group.Members[setOwnerID] = member
	} else {
		if len(memberName) <= 0 {
			memberName = "A Cool Person!"
		}

		member = &User{
			ID:    fmt.Sprintf("%v", RandomID()),
			Token: RandomToken(40),
			Name:  memberName,
			Socket: socket,
		}

		group.Members[member.ID] = member
	}

	data := Packet{
		Code: PacketSetUserData,
		Data: Map{
			"userID": member.ID,
			"group": Map{
				"id": group.ID,
				"name": group.Name,
				"owner_id": group.OwnerID,
				"owner_name": group.Members[group.OwnerID].Name,
				"online_members": len(group.Members),
			},
		},
	}

	socket.WriteJSON(data)

	ticker := time.NewTicker(5 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
				case <- ticker.C:
					err := socket.WriteJSON(data)

					if err != nil {
						if group.OwnerID == member.ID {
							for _, member = range group.Members {
								member.Socket.WriteJSON(&Packet{
									Code: PacketClose,
									Data: Map{
										"closed": true,
									},
								})
							}
						}

						err := socket.Close()
						if err != nil {
							fmt.Println(err)
						}

						delete(Groups, group.ID)
						close(quit)
					}

				case <- quit:
					ticker.Stop()
					return
				}
		}
	}()

	for {
		var packet *Packet

		err := socket.ReadJSON(&packet)
		if err != nil {
			socket.WriteJSON(&Packet{
				Code: PacketPacket,
				Data: Map{
					"message": "An error has been ocurred while the packet content was tried to read",
				},
			})

			continue
		}
		
		if len(packet.Code) <= 0 || packet.Data == nil {
			continue
		}

		packet.Data["author_id"] = member.ID
		packet.Data["author_token"] = member.Token
		packet.Data["author_socket"] = socket
		packet.Data["group_id"] = group.ID

		packetQueue <- packet
	}
}

func ManagePackets() {
	for {
		packet := <-packetQueue

		switch packet.Code {
		case PacketSendMessage:
			SendMessage(packet)
		}
	}
}

func SendMessage(packet *Packet) {
	id := packet.Data["author_id"].(string)
	token := packet.Data["author_token"].(string)
	groupID := packet.Data["group_id"].(string)
	content := packet.Data["content"].(string)

	var socket *websocket.Conn
	socket = packet.Data["author_socket"].(*websocket.Conn)

	if len(content) <= 0 {
		socket.WriteJSON(&Packet{
			Code: PacketSetMessageError,
			Data: Map{
				"message": "The message content is invalid",
			},
		})

		return
	}

	if len(content) > 2000 {
		socket.WriteJSON(&Packet{
			Code: PacketSetMessageError,
			Data: Map{
				"message": "The message content cannot be higher than 2000 characters",
			},
		})

		return
	}

	group := Groups[groupID]
	if group == nil {
		socket.WriteJSON(&Packet{
			Code: PacketSetMessageError,
			Data: Map{
				"message": "Unknown group",
			},
		})

		return
	}

	var authorListed *User
	for _, member := range group.Members {
		if member.ID == id {
			authorListed = member
			break
		}
	}

	if authorListed == nil {
		socket.WriteJSON(&Packet{
			Code: PacketSetMessageError,
			Data: Map{
				"message": "You not are in this group",
			},
		})

		return
	}

	if strings.Compare(authorListed.Token, token) != 0 {
		socket.WriteJSON(&Packet{
			Code: PacketSetMessageError,
			Data: Map{
				"message": "Reload the page, your token has been expired or you never has one",
			},
		})

		return
	}

	message := Message{
		AuthorID:  id,
		AuthorName: authorListed.Name,
		Content:   content,
		CreatedAt: time.Now(),
	}

	for _, member := range group.Members {
		fmt.Println(member.Socket)
		if member.Socket != nil {
			err := member.Socket.WriteJSON(&Packet{
				Code: PacketReceiveMessage,
				Data: Map{
					"message": message,
				},
			})

			if err != nil {
				fmt.Println(err)
			}
		}
	}
}
