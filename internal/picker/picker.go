package picker

type SharedType int

const (
	File SharedType = 1
	Text SharedType = 2
)

type PickResult struct {
	Path string
	Name string
	Err  error
}

type SharedEvent struct {
	Type SharedType
	Path string
	Name string
	Text string
}
