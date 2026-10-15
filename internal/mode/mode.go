package mode

type Mode int

const (
	Normal Mode = iota
	Insert
)

func (m Mode) String() string {
	if m == Insert {
		return "INSERT"
	}
	return "NORMAL"
}
