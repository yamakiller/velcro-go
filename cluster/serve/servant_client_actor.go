package serve

type ServantClientActor interface {
	onClosed(*ServantClientContext)
}
