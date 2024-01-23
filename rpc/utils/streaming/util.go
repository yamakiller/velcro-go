package streaming

import "github.com/yamakiller/velcro-go/rpc/utils/stats/serviceinfo"

// VelcroUnusedProtection may be anonymously referenced in another package to avoid build error
const VelcroUnusedProtection = 0

// UnaryCompatibleMiddleware returns whether to use compatible middleware for unary.
func UnaryCompatibleMiddleware(mode serviceinfo.StreamingMode, allow bool) bool {
	return allow && mode == serviceinfo.StreamingUnary
}
