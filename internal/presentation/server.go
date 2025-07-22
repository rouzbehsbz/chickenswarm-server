package presentation

import (
	"bytes"
	"encoding/binary"
	"net/http"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

type Message struct {
	TypeId  uint8
	Message proto.Message
}

type IncomingMessage struct {
	Conn    *websocket.Conn
	Message proto.Message
}

type Server struct {
	upgrader websocket.Upgrader

	IncomingPlayers chan *websocket.Conn
	OutgoingPlayers chan *websocket.Conn

	IncomingMessages chan IncomingMessage
}

func NewServer() *Server {
	return &Server{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		IncomingPlayers:  make(chan *websocket.Conn),
		OutgoingPlayers:  make(chan *websocket.Conn),
		IncomingMessages: make(chan IncomingMessage),
	}
}

func (s *Server) UpgradeConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	s.IncomingPlayers <- conn

	go s.handleIncomingMessages(conn)
}

func (s *Server) handleIncomingMessages(conn *websocket.Conn) {
	defer func() {
		s.OutgoingPlayers <- conn
	}()

	for {
		_, msg, err := conn.ReadMessage()

		if err != nil {
			println(err.Error())
			return
		}

		reader := bytes.NewReader(msg)

		for reader.Len() > 0 {
			var typeId uint8

			if err := binary.Read(reader, binary.BigEndian, &typeId); err != nil {
				return
			}

			var length uint32
			if err := binary.Read(reader, binary.BigEndian, &length); err != nil {
				return
			}

			data := make([]byte, length)
			if _, err := reader.Read(data); err != nil {
				return
			}

			creator, ok := messageTypeRegistry[typeId]
			if !ok {
				println(ok)
				continue
			}

			msg := creator()
			if msg == nil {
				continue
			}

			if err := proto.Unmarshal(data, msg); err != nil {
				continue
			}

			s.IncomingMessages <- IncomingMessage{Conn: conn, Message: msg}
		}
	}
}

func (s *Server) CreatePacket(messages []Message) ([]byte, error) {
	var payload bytes.Buffer

	for _, msg := range messages {
		data, err := proto.Marshal(msg.Message)
		if err != nil {
			return nil, err
		}

		if err := binary.Write(&payload, binary.BigEndian, msg.TypeId); err != nil {
			return nil, err
		}

		if err := binary.Write(&payload, binary.BigEndian, uint32(len(data))); err != nil {
			return nil, err
		}

		if _, err := payload.Write(data); err != nil {
			return nil, err
		}
	}

	return payload.Bytes(), nil
}
