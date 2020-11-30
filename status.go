package main

type Status int

const (
	Nonexistent Status = iota
	Alive
	Dead
)

func (cs Status) String() string {
	var s string
	switch cs {
	case Nonexistent:
		s = "🌑"
	case Dead:
		s = "⭕️"
	case Alive:
		s = "🟢"
	default:
		panic("invalid status")
	}
	return s
}
