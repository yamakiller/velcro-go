package xds

import (
	"github.com/yamakiller/velcro-go/rpc2/pkg/discovery"
	"github.com/yamakiller/velcro-go/utils/endpoint"
)

// ClientSuite includes all extensions used by the xDS-enabled client.
type ClientSuite struct {
	RouterMiddleware endpoint.Middleware
	Resolver         discovery.Resolver
}

// CheckClientSuite checks if the client suite is valid.
// RouteMiddleware and Resolver should not be nil.
func CheckClientSuite(cs ClientSuite) bool {
	return cs.RouterMiddleware != nil && cs.Resolver != nil
}
