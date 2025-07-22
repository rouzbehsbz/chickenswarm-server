package common

type Vector2 struct {
	X int
	Y int
}

func NewVector2(x, y int) *Vector2 {
	return &Vector2{
		X: x,
		Y: y,
	}
}

func NewZeroVector2() *Vector2 {
	return NewVector2(0, 0)
}
