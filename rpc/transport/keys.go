package transport

type Protocol int

// Predefined transport protocols.
const (
	Framed Protocol = 1 << iota
	HESSIAN2
)

// Unknown indicates the protocol is unknown.
const Unknown = "Unknown"

// String prints human readable information.
func (tp Protocol) String() string {
	switch tp {
	case Framed:
		return "Framed"
	case HESSIAN2:
		return "Hessian2"
	}
	return Unknown
}
