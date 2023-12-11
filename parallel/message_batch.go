package parallel

type MessageBatch interface {
	GetMessages() []interface{}
}
