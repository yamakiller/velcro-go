package serve

type ServantClientActor interface {
	Closed(*ServantClientContext)
}
