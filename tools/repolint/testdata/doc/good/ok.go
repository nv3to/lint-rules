package good

// Thing is documented.
type Thing struct{}

// A Widget starts with an article.
type Widget int

// Do is documented.
func (Thing) Do() {}

func (Thing) String() string { return "" }

// Limits are grouped.
const (
	One = 1
	Two = 2
)

// Max, Min are declared together.
var Max, Min = 1, 0

func hidden() {}

type unexported struct{}

func (unexported) Exported() {}
