package protocol

// TODO No we want to use streams, or at least Write Read should be available
// but these names should be saved for ioReader ioWriter. Think http with
// middleware.
type Protocol interface {
	Write([]byte) (int, error) // NOTE Bytes written, error for a generic write
	Read([]byte) (int, error)
	Name() string
	Flow() string
	Pins() []int
}
