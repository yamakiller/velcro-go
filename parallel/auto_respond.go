package parallel

type AutoRespond interface {
	GetAutoResponse(context Context) interface{}
}
