package transport

type Protocol int

// Predefined transport protocols.
const (
	PurePayload Protocol = 0

	VHeader Protocol = 1 << iota
	Framed
	//HTTP
	GRPC
	HESSIAN2

	VHeaderFramed = VHeader | Framed
)

// Unknown indicates the protocol is unknown.
const Unknown = "Unknown"

// String prints human readable information.
func (tp Protocol) String() string {
	switch tp {
	case PurePayload:
		return "PurePayload"
	case VHeader:
		return "VHeader"
	case Framed:
		return "Framed"
	//case HTTP:
	//	return "HTTP"
	case VHeaderFramed:
		return "VHeaderFramed"
	case GRPC:
		return "GRPC"
	case HESSIAN2:
		return "Hessian2"
	}
	return Unknown
}
