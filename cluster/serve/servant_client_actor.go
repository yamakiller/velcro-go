package serve

import "context"

type ServantClientActor interface {
	Closed(context.Context)
}
