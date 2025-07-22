package presentation

import (
	"chickenswarm-server/internal/presentation/protobuf"

	"google.golang.org/protobuf/proto"
)

const (
	WelcomeMessageTypeId    = 1
	MoveMessageTypeId       = 2
	PlayersMessageTypeId    = 3
	DisconnectMessageTypeId = 4
	PingMessageTypeId       = 5
	PongMesssageTypeId      = 6
	JoinMessageTypeId       = 7
)

type MessageTypeCreator func() proto.Message

var messageTypeRegistry = map[uint8]MessageTypeCreator{
	1: func() proto.Message { return &protobuf.Welcome{} },
	2: func() proto.Message { return &protobuf.Move{} },
	3: func() proto.Message { return &protobuf.Players{} },
	4: func() proto.Message { return &protobuf.Disconnect{} },
	5: func() proto.Message { return &protobuf.Ping{} },
	6: func() proto.Message { return &protobuf.Pong{} },
	7: func() proto.Message { return &protobuf.Join{} },
}
