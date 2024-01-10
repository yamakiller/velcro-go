package connectPool

import (
	"google.golang.org/protobuf/proto"
)


type IFuture interface{
	Error() error 
	Result() proto.Message
	Wait()
}