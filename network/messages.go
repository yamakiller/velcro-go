package network

type AutoReceiveMessage interface {
	AutoReceiveMessage()
}

type SystemMessage interface {
	SystemMessage()
}

type Closing struct{}

type Closed struct{}

func (*Closing) AutoReceiveMessage() {}
func (*Closed) AutoReceiveMessage()  {}

func (*Activation) SystemMessage() {}
func (*Close) SystemMessage()      {}

var (
	closingMessage AutoReceiveMessage = &Closing{}
	closedMessage  AutoReceiveMessage = &Closed{}

	activationMessage SystemMessage = &Activation{}
	closeMessage      SystemMessage = &Close{}
)
