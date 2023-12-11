package parallel

type header map[string]string

func (m header) Get(key string) string {
	return m[key]
}

func (m header) Set(key string, value string) {
	m[key] = value
}

func (m header) Keys() []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	return keys
}

func (m header) Length() int {
	return len(m)
}

func (m header) ToMap() map[string]string {
	mp := make(map[string]string)
	for k, v := range m {
		mp[k] = v
	}
	return mp
}

// ReadOnlyMessageHeader : 只读消息头对象
type ReadOnlyMessageHeader interface {
	Get(key string) string
	Keys() []string
	Length() int
	ToMap() map[string]string
}

// Pack : 消息报对象
type MessageReport struct {
	Header  header
	Message interface{}
	Sender  *PID
}

// GetHeader : 获取包的对象的包头信息 key -> string
func (mp *MessageReport) GetHeader(key string) string {
	if mp.Header == nil {
		return ""
	}

	return mp.Header.Get(key)
}

// SetHeader :  设置消息头信息 key -> value
func (mp *MessageReport) SetHeader(key string, value string) {
	if mp.Header == nil {
		mp.Header = make(map[string]string)
	}
	mp.Header.Set(key, value)
}

// DefaultMessageHeader : 默认消息头
var DefaultMessageHeader = make(header)

// WrapReport 消息打包
func WrapReport(message interface{}) *MessageReport {
	if e, ok := message.(*MessageReport); ok {
		return e
	}

	return &MessageReport{nil, message, nil}
}

// UnWrapPack : 消息包拆分返回 [消息头 | 消息 | 发送者]
func UnWrapReport(message interface{}) (ReadOnlyMessageHeader, interface{}, *PID) {
	if e, ok := message.(*MessageReport); ok {
		return e.Header, e.Message, e.Sender
	}
	return nil, message, nil
}

// UnWrapReportHeader : 消息包拆分返回 [消息头]
func UnWrapReportHeader(message interface{}) ReadOnlyMessageHeader {
	if e, ok := message.(*MessageReport); ok {
		return e.Header
	}

	return nil
}

// UnWrapReportMessage : 消息包拆分返回[消息]
func UnWrapReportMessage(message interface{}) interface{} {
	if e, ok := message.(*MessageReport); ok {
		return e.Message
	}

	return message
}

// UnWrapReportSender : 消息包拆分返回[发送者]
func UnWrapReportSender(message interface{}) *PID {
	if e, ok := message.(*MessageReport); ok {
		return e.Sender
	}

	return nil
}
