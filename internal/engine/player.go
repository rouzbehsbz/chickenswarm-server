package engine

import (
	"chickenswarm-server/internal/common"
	"regexp"
	"strings"

	"github.com/gorilla/websocket"
)

const (
	PlayerDirectionUp    = 0
	PlayerDirectionDown  = 1
	PlayerDirectionLeft  = 2
	PlayerDirectionRight = 3
)

const (
	PlayerSpeed = 5
)

type Player struct {
	Id             int
	Nickname       string
	Position       *common.Vector2
	SequenceNumber int

	Conn *websocket.Conn
}

func NewPlayer(id, x, y int, conn *websocket.Conn) *Player {
	return &Player{
		Id:             id,
		Position:       common.NewVector2(x, y),
		Conn:           conn,
		SequenceNumber: 0,
	}
}

func (p *Player) Move(direction, sequenceNumber int) {
	p.SequenceNumber = sequenceNumber
	d := common.NewZeroVector2()

	switch direction {
	case PlayerDirectionUp:
		d.Y -= PlayerSpeed
	case PlayerDirectionDown:
		d.Y += PlayerSpeed
	case PlayerDirectionRight:
		d.X += PlayerSpeed
	case PlayerDirectionLeft:
		d.X -= PlayerSpeed
	}

	p.Position.X += d.X
	p.Position.Y += d.Y
}

func (p *Player) IsNicknameValid(nickname string) bool {
	const minLength = 3
	const maxLength = 16

	if nickname == "" {
		return false
	}

	nickname = strings.TrimSpace(nickname)

	if len(nickname) < minLength || len(nickname) > maxLength {
		return false
	}

	if !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(nickname) {
		return false
	}

	if strings.Contains(nickname, "__") {
		return false
	}

	return true
}
