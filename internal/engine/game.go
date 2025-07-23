package engine

import (
	"chickenswarm-server/internal/presentation"
	"chickenswarm-server/internal/presentation/protobuf"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	WORLD_WIDTH   = 700
	WORLD_HEIGHT  = 700
	FPS           = 60
	PING_INTERVAL = 3 * time.Second
)

type Game struct {
	Server *presentation.Server

	playersIndexById   map[int]*Player
	playersIndexByConn map[*websocket.Conn]*Player
	playerIdCounter    int
	mup                sync.RWMutex

	broadcastBuffer []presentation.Message
	mub             sync.RWMutex
}

func NewGame() *Game {
	game := &Game{
		Server: presentation.NewServer(),

		playersIndexById:   make(map[int]*Player),
		playersIndexByConn: make(map[*websocket.Conn]*Player),
		playerIdCounter:    0,
		broadcastBuffer:    []presentation.Message{},
		mup:                sync.RWMutex{},
		mub:                sync.RWMutex{},
	}

	go game.startLoop()
	go game.handleIncomingPlayers()
	go game.handleOutgoingPlayers()
	go game.handleIncomingMessages()

	return game
}

func (g *Game) startLoop() {
	lastFrameTime := time.Now().UnixMilli()

	for {
		now := time.Now().UnixMilli()
		delta := now - lastFrameTime

		lastFrameTime = now

		g.updateWorldState()

		packet, _ := g.Server.CreatePacket(g.broadcastBuffer)
		g.broadcastPacket(packet)
		g.broadcastBuffer = []presentation.Message{}

		sleepTime := math.Max(0, float64(1000/FPS)-float64(delta))

		time.Sleep(time.Duration(sleepTime * float64(time.Millisecond)))
	}
}

func (g *Game) AddPlayer(conn *websocket.Conn) *Player {
	g.mup.Lock()
	defer g.mup.Unlock()

	g.playerIdCounter += 1

	x := rand.Intn(WORLD_WIDTH)
	y := rand.Intn(WORLD_HEIGHT)

	player := NewPlayer(g.playerIdCounter, x, y, conn)

	g.playersIndexById[g.playerIdCounter] = player
	g.playersIndexByConn[conn] = player

	return player
}

func (g *Game) RemovePlayer(conn *websocket.Conn) {
	g.mup.Lock()
	defer g.mup.Unlock()

	if player, exists := g.playersIndexByConn[conn]; exists {
		g.AddToBroadcastBuffer(presentation.Message{
			TypeId: presentation.DisconnectMessageTypeId,
			Message: &protobuf.Disconnect{
				PlayerId: int32(player.Id),
			},
		})

		delete(g.playersIndexById, player.Id)
		delete(g.playersIndexByConn, conn)

		conn.Close()
	}
}

func (g *Game) handleIncomingPlayers() {
	for conn := range g.Server.IncomingPlayers {
		_ = g.AddPlayer(conn)

		players := []*protobuf.Nicknames_Player{}
		for _, player := range g.playersIndexById {
			if player.Nickname == "" {
				continue
			}

			players = append(players, &protobuf.Nicknames_Player{
				PlayerId: int32(player.Id),
				Nickname: player.Nickname,
			})
		}

		buffer := []presentation.Message{
			{
				TypeId: presentation.NicknamesMessageTypeId,
				Message: &protobuf.Nicknames{
					Players: players,
				},
			},
		}

		packet, _ := g.Server.CreatePacket(buffer)
		g.sendPacket(packet, conn)
	}
}

func (g *Game) handleOutgoingPlayers() {
	for conn := range g.Server.OutgoingPlayers {
		g.RemovePlayer(conn)
	}
}

func (g *Game) handleIncomingMessages() {
	for msg := range g.Server.IncomingMessages {
		player, exists := g.playersIndexByConn[msg.Conn]
		if !exists {
			g.RemovePlayer(msg.Conn)
			continue
		}

		switch m := msg.Message.(type) {
		case *protobuf.Move:
			player, exists := g.playersIndexById[player.Id]

			if exists {
				player.Move(int(m.Direction), int(m.SequenceNumber))
			}
		case *protobuf.Ping:
			buffer := []presentation.Message{
				{
					TypeId: presentation.PongMesssageTypeId,
					Message: &protobuf.Pong{
						Timestamp: time.Now().UnixMilli(),
					},
				},
			}

			packet, _ := g.Server.CreatePacket(buffer)
			g.sendPacket(packet, msg.Conn)

		case *protobuf.Join:
			if !player.IsNicknameValid(m.Nickname) {
				continue
			}

			player.Nickname = m.Nickname

			connectBuffer := []presentation.Message{
				{
					TypeId: presentation.ConnectMessageTypeId,
					Message: &protobuf.Connect{
						PlayerId: int32(player.Id),
						Nickname: player.Nickname,
					},
				},
			}

			welcomeBuffer := []presentation.Message{
				{
					TypeId: presentation.WelcomeMessageTypeId,
					Message: &protobuf.Welcome{
						PlayerId: int32(player.Id),
						X:        int32(player.Position.X),
						Y:        int32(player.Position.Y),
					},
				},
			}

			g.AddToBroadcastBuffer(connectBuffer...)

			welcomePacket, _ := g.Server.CreatePacket(welcomeBuffer)
			g.sendPacket(welcomePacket, msg.Conn)
		}
	}
}

func (g *Game) updateWorldState() {
	buffer := []presentation.Message{}

	players := []*protobuf.Players_Player{}
	for _, player := range g.playersIndexById {
		if player.Nickname == "" {
			continue
		}

		players = append(players, &protobuf.Players_Player{
			PlayerId:       int32(player.Id),
			X:              int32(player.Position.X),
			Y:              int32(player.Position.Y),
			SequenceNumber: int32(player.SequenceNumber),
		})
	}

	buffer = append(buffer, presentation.Message{
		TypeId: presentation.PlayersMessageTypeId,
		Message: &protobuf.Players{
			Players:   players,
			Timestamp: time.Now().UnixMilli(),
		},
	})

	g.AddToBroadcastBuffer(buffer...)
}

func (g *Game) AddToBroadcastBuffer(message ...presentation.Message) {
	g.mub.Lock()
	defer g.mub.Unlock()

	g.broadcastBuffer = append(g.broadcastBuffer, message...)
}

func (g *Game) broadcastPacket(packet []byte) {
	g.mup.Lock()
	defer g.mup.Unlock()

	for _, player := range g.playersIndexById {
		_ = player.Conn.WriteMessage(websocket.BinaryMessage, packet)
	}
}

func (g *Game) sendPacket(packet []byte, conn *websocket.Conn) {
	g.mup.Lock()
	defer g.mup.Unlock()

	_ = conn.WriteMessage(websocket.BinaryMessage, packet)
}
