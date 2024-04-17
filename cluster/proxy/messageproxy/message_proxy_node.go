package messageproxy

func NewMessageProxyNode()IMessageProxyNode{
	return &MessageProxyNode{}
}

type IMessageProxyNode interface{
	WithNext(IMessageProxyNode)
	Next()IMessageProxyNode
	JumpNext()
}

type MessageProxyNode struct{
	next IMessageProxyNode
}
func (m *MessageProxyNode) WithNext(n IMessageProxyNode){
	m.next = n
}
func (m *MessageProxyNode) Next()IMessageProxyNode{
	return m.next
}
func (m *MessageProxyNode) JumpNext(){
	if m.next != nil && m.next.Next() != nil{
		m.WithNext(m.next.Next())
	}
}