package bad

type Bare struct{}

// wrong start.
func Exported() {}

func (Bare) Method() {}

var Loose = 1

// Documented is fine.
type Documented struct{}

const (
	Ungrouped = 1
)
