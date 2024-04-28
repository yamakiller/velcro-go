// Code generated by Thrift Compiler (0.19.0). DO NOT EDIT.

package prvs

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"
	thrift "github.com/apache/thrift/lib/go/thrift"
	"strings"
	"regexp"
	"github.com/yamakiller/velcro-go/network"

)

// (needed to ensure safety because of naive import list construction.)
var _ = thrift.ZERO
var _ = fmt.Printf
var _ = errors.New
var _ = context.Background
var _ = time.Now
var _ = bytes.Equal
// (needed by validator.)
var _ = strings.Contains
var _ = regexp.MatchString

// Attributes:
//  - Sender
//  - Body
type ForwardBundle struct {
  Sender *network.ClientID `thrift:"sender,1" db:"sender" json:"sender"`
  Body []byte `thrift:"body,2" db:"body" json:"body"`
}

func NewForwardBundle() *ForwardBundle {
  return &ForwardBundle{}
}

var ForwardBundle_Sender_DEFAULT *network.ClientID
func (p *ForwardBundle) GetSender() *network.ClientID {
  if !p.IsSetSender() {
    return ForwardBundle_Sender_DEFAULT
  }
return p.Sender
}

func (p *ForwardBundle) GetBody() []byte {
  return p.Body
}
func (p *ForwardBundle) IsSetSender() bool {
  return p.Sender != nil
}

func (p *ForwardBundle) Read(ctx context.Context, iprot thrift.TProtocol) error {
  if _, err := iprot.ReadStructBegin(ctx); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T read error: ", p), err)
  }


  for {
    _, fieldTypeId, fieldId, err := iprot.ReadFieldBegin(ctx)
    if err != nil {
      return thrift.PrependError(fmt.Sprintf("%T field %d read error: ", p, fieldId), err)
    }
    if fieldTypeId == thrift.STOP { break; }
    switch fieldId {
    case 1:
      if fieldTypeId == thrift.STRUCT {
        if err := p.ReadField1(ctx, iprot); err != nil {
          return err
        }
      } else {
        if err := iprot.Skip(ctx, fieldTypeId); err != nil {
          return err
        }
      }
    case 2:
      if fieldTypeId == thrift.STRING {
        if err := p.ReadField2(ctx, iprot); err != nil {
          return err
        }
      } else {
        if err := iprot.Skip(ctx, fieldTypeId); err != nil {
          return err
        }
      }
    default:
      if err := iprot.Skip(ctx, fieldTypeId); err != nil {
        return err
      }
    }
    if err := iprot.ReadFieldEnd(ctx); err != nil {
      return err
    }
  }
  if err := iprot.ReadStructEnd(ctx); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T read struct end error: ", p), err)
  }
  return nil
}

func (p *ForwardBundle)  ReadField1(ctx context.Context, iprot thrift.TProtocol) error {
  p.Sender = &network.ClientID{}
  if err := p.Sender.Read(ctx, iprot); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T error reading struct: ", p.Sender), err)
  }
  return nil
}

func (p *ForwardBundle)  ReadField2(ctx context.Context, iprot thrift.TProtocol) error {
  if v, err := iprot.ReadBinary(ctx); err != nil {
  return thrift.PrependError("error reading field 2: ", err)
} else {
  p.Body = v
}
  return nil
}

func (p *ForwardBundle) Write(ctx context.Context, oprot thrift.TProtocol) error {
  if err := oprot.WriteStructBegin(ctx, "ForwardBundle"); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write struct begin error: ", p), err) }
  if p != nil {
    if err := p.writeField1(ctx, oprot); err != nil { return err }
    if err := p.writeField2(ctx, oprot); err != nil { return err }
  }
  if err := oprot.WriteFieldStop(ctx); err != nil {
    return thrift.PrependError("write field stop error: ", err) }
  if err := oprot.WriteStructEnd(ctx); err != nil {
    return thrift.PrependError("write struct stop error: ", err) }
  return nil
}

func (p *ForwardBundle) writeField1(ctx context.Context, oprot thrift.TProtocol) (err error) {
  if err := oprot.WriteFieldBegin(ctx, "sender", thrift.STRUCT, 1); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write field begin error 1:sender: ", p), err) }
  if err := p.Sender.Write(ctx, oprot); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T error writing struct: ", p.Sender), err)
  }
  if err := oprot.WriteFieldEnd(ctx); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write field end error 1:sender: ", p), err) }
  return err
}

func (p *ForwardBundle) writeField2(ctx context.Context, oprot thrift.TProtocol) (err error) {
  if err := oprot.WriteFieldBegin(ctx, "body", thrift.STRING, 2); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write field begin error 2:body: ", p), err) }
  if err := oprot.WriteBinary(ctx, p.Body); err != nil {
  return thrift.PrependError(fmt.Sprintf("%T.body (2) field write error: ", p), err) }
  if err := oprot.WriteFieldEnd(ctx); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write field end error 2:body: ", p), err) }
  return err
}

func (p *ForwardBundle) Equals(other *ForwardBundle) bool {
  if p == other {
    return true
  } else if p == nil || other == nil {
    return false
  }
  if !p.Sender.Equals(other.Sender) { return false }
  if bytes.Compare(p.Body, other.Body) != 0 { return false }
  return true
}

func (p *ForwardBundle) String() string {
  if p == nil {
    return "<nil>"
  }
  return fmt.Sprintf("ForwardBundle(%+v)", *p)
}

func (p *ForwardBundle) Validate() error {
  return nil
}
// Attributes:
//  - Target
//  - Body
type RequestGatewayPush struct {
  Target *network.ClientID `thrift:"target,1" db:"target" json:"target"`
  Body []byte `thrift:"body,2" db:"body" json:"body"`
}

func NewRequestGatewayPush() *RequestGatewayPush {
  return &RequestGatewayPush{}
}

var RequestGatewayPush_Target_DEFAULT *network.ClientID
func (p *RequestGatewayPush) GetTarget() *network.ClientID {
  if !p.IsSetTarget() {
    return RequestGatewayPush_Target_DEFAULT
  }
return p.Target
}

func (p *RequestGatewayPush) GetBody() []byte {
  return p.Body
}
func (p *RequestGatewayPush) IsSetTarget() bool {
  return p.Target != nil
}

func (p *RequestGatewayPush) Read(ctx context.Context, iprot thrift.TProtocol) error {
  if _, err := iprot.ReadStructBegin(ctx); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T read error: ", p), err)
  }


  for {
    _, fieldTypeId, fieldId, err := iprot.ReadFieldBegin(ctx)
    if err != nil {
      return thrift.PrependError(fmt.Sprintf("%T field %d read error: ", p, fieldId), err)
    }
    if fieldTypeId == thrift.STOP { break; }
    switch fieldId {
    case 1:
      if fieldTypeId == thrift.STRUCT {
        if err := p.ReadField1(ctx, iprot); err != nil {
          return err
        }
      } else {
        if err := iprot.Skip(ctx, fieldTypeId); err != nil {
          return err
        }
      }
    case 2:
      if fieldTypeId == thrift.STRING {
        if err := p.ReadField2(ctx, iprot); err != nil {
          return err
        }
      } else {
        if err := iprot.Skip(ctx, fieldTypeId); err != nil {
          return err
        }
      }
    default:
      if err := iprot.Skip(ctx, fieldTypeId); err != nil {
        return err
      }
    }
    if err := iprot.ReadFieldEnd(ctx); err != nil {
      return err
    }
  }
  if err := iprot.ReadStructEnd(ctx); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T read struct end error: ", p), err)
  }
  return nil
}

func (p *RequestGatewayPush)  ReadField1(ctx context.Context, iprot thrift.TProtocol) error {
  p.Target = &network.ClientID{}
  if err := p.Target.Read(ctx, iprot); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T error reading struct: ", p.Target), err)
  }
  return nil
}

func (p *RequestGatewayPush)  ReadField2(ctx context.Context, iprot thrift.TProtocol) error {
  if v, err := iprot.ReadBinary(ctx); err != nil {
  return thrift.PrependError("error reading field 2: ", err)
} else {
  p.Body = v
}
  return nil
}

func (p *RequestGatewayPush) Write(ctx context.Context, oprot thrift.TProtocol) error {
  if err := oprot.WriteStructBegin(ctx, "RequestGatewayPush"); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write struct begin error: ", p), err) }
  if p != nil {
    if err := p.writeField1(ctx, oprot); err != nil { return err }
    if err := p.writeField2(ctx, oprot); err != nil { return err }
  }
  if err := oprot.WriteFieldStop(ctx); err != nil {
    return thrift.PrependError("write field stop error: ", err) }
  if err := oprot.WriteStructEnd(ctx); err != nil {
    return thrift.PrependError("write struct stop error: ", err) }
  return nil
}

func (p *RequestGatewayPush) writeField1(ctx context.Context, oprot thrift.TProtocol) (err error) {
  if err := oprot.WriteFieldBegin(ctx, "target", thrift.STRUCT, 1); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write field begin error 1:target: ", p), err) }
  if err := p.Target.Write(ctx, oprot); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T error writing struct: ", p.Target), err)
  }
  if err := oprot.WriteFieldEnd(ctx); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write field end error 1:target: ", p), err) }
  return err
}

func (p *RequestGatewayPush) writeField2(ctx context.Context, oprot thrift.TProtocol) (err error) {
  if err := oprot.WriteFieldBegin(ctx, "body", thrift.STRING, 2); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write field begin error 2:body: ", p), err) }
  if err := oprot.WriteBinary(ctx, p.Body); err != nil {
  return thrift.PrependError(fmt.Sprintf("%T.body (2) field write error: ", p), err) }
  if err := oprot.WriteFieldEnd(ctx); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write field end error 2:body: ", p), err) }
  return err
}

func (p *RequestGatewayPush) Equals(other *RequestGatewayPush) bool {
  if p == other {
    return true
  } else if p == nil || other == nil {
    return false
  }
  if !p.Target.Equals(other.Target) { return false }
  if bytes.Compare(p.Body, other.Body) != 0 { return false }
  return true
}

func (p *RequestGatewayPush) String() string {
  if p == nil {
    return "<nil>"
  }
  return fmt.Sprintf("RequestGatewayPush(%+v)", *p)
}

func (p *RequestGatewayPush) Validate() error {
  return nil
}
type RequestGatewayPushService interface {
  // Parameters:
  //  - Req
  OnRequestGatewayPush(ctx context.Context, req *RequestGatewayPush) (_err error)
}

type RequestGatewayPushServiceClient struct {
  c thrift.TClient
  meta thrift.ResponseMeta
}

func NewRequestGatewayPushServiceClientFactory(t thrift.TTransport, f thrift.TProtocolFactory) *RequestGatewayPushServiceClient {
  return &RequestGatewayPushServiceClient{
    c: thrift.NewTStandardClient(f.GetProtocol(t), f.GetProtocol(t)),
  }
}

func NewRequestGatewayPushServiceClientProtocol(t thrift.TTransport, iprot thrift.TProtocol, oprot thrift.TProtocol) *RequestGatewayPushServiceClient {
  return &RequestGatewayPushServiceClient{
    c: thrift.NewTStandardClient(iprot, oprot),
  }
}

func NewRequestGatewayPushServiceClient(c thrift.TClient) *RequestGatewayPushServiceClient {
  return &RequestGatewayPushServiceClient{
    c: c,
  }
}

func (p *RequestGatewayPushServiceClient) Client_() thrift.TClient {
  return p.c
}

func (p *RequestGatewayPushServiceClient) LastResponseMeta_() thrift.ResponseMeta {
  return p.meta
}

func (p *RequestGatewayPushServiceClient) SetLastResponseMeta_(meta thrift.ResponseMeta) {
  p.meta = meta
}

// Parameters:
//  - Req
func (p *RequestGatewayPushServiceClient) OnRequestGatewayPush(ctx context.Context, req *RequestGatewayPush) (_err error) {
  var _args0 RequestGatewayPushServiceOnRequestGatewayPushArgs
  _args0.Req = req
  var _result2 RequestGatewayPushServiceOnRequestGatewayPushResult
  var _meta1 thrift.ResponseMeta
  _meta1, _err = p.Client_().Call(ctx, "OnRequestGatewayPush", &_args0, &_result2)
  p.SetLastResponseMeta_(_meta1)
  if _err != nil {
    return
  }
  return nil
}

type RequestGatewayPushServiceProcessor struct {
  processorMap map[string]thrift.TProcessorFunction
  handler RequestGatewayPushService
}

func (p *RequestGatewayPushServiceProcessor) AddToProcessorMap(key string, processor thrift.TProcessorFunction) {
  p.processorMap[key] = processor
}

func (p *RequestGatewayPushServiceProcessor) GetProcessorFunction(key string) (processor thrift.TProcessorFunction, ok bool) {
  processor, ok = p.processorMap[key]
  return processor, ok
}

func (p *RequestGatewayPushServiceProcessor) ProcessorMap() map[string]thrift.TProcessorFunction {
  return p.processorMap
}

func NewRequestGatewayPushServiceProcessor(handler RequestGatewayPushService) *RequestGatewayPushServiceProcessor {

  self3 := &RequestGatewayPushServiceProcessor{handler:handler, processorMap:make(map[string]thrift.TProcessorFunction)}
  self3.processorMap["OnRequestGatewayPush"] = &requestGatewayPushServiceProcessorOnRequestGatewayPush{handler:handler}
return self3
}

func (p *RequestGatewayPushServiceProcessor) Process(ctx context.Context, iprot, oprot thrift.TProtocol) (success bool, err thrift.TException) {
  name, _, seqId, err2 := iprot.ReadMessageBegin(ctx)
  if err2 != nil { return false, thrift.WrapTException(err2) }
  if processor, ok := p.GetProcessorFunction(name); ok {
    return processor.Process(ctx, seqId, iprot, oprot)
  }
  iprot.Skip(ctx, thrift.STRUCT)
  iprot.ReadMessageEnd(ctx)
  x4 := thrift.NewTApplicationException(thrift.UNKNOWN_METHOD, "Unknown function " + name)
  oprot.WriteMessageBegin(ctx, name, thrift.EXCEPTION, seqId)
  x4.Write(ctx, oprot)
  oprot.WriteMessageEnd(ctx)
  oprot.Flush(ctx)
  return false, x4

}

type requestGatewayPushServiceProcessorOnRequestGatewayPush struct {
  handler RequestGatewayPushService
}

func (p *requestGatewayPushServiceProcessorOnRequestGatewayPush) Process(ctx context.Context, seqId int32, iprot, oprot thrift.TProtocol) (success bool, err thrift.TException) {
  var _write_err5 error
  args := RequestGatewayPushServiceOnRequestGatewayPushArgs{}
  if err2 := args.Read(ctx, iprot); err2 != nil {
    iprot.ReadMessageEnd(ctx)
    x := thrift.NewTApplicationException(thrift.PROTOCOL_ERROR, err2.Error())
    oprot.WriteMessageBegin(ctx, "OnRequestGatewayPush", thrift.EXCEPTION, seqId)
    x.Write(ctx, oprot)
    oprot.WriteMessageEnd(ctx)
    oprot.Flush(ctx)
    return false, thrift.WrapTException(err2)
  }
  iprot.ReadMessageEnd(ctx)

  tickerCancel := func() {}
  // Start a goroutine to do server side connectivity check.
  if thrift.ServerConnectivityCheckInterval > 0 {
    var cancel context.CancelCauseFunc
    ctx, cancel = context.WithCancelCause(ctx)
    defer cancel(nil)
    var tickerCtx context.Context
    tickerCtx, tickerCancel = context.WithCancel(context.Background())
    defer tickerCancel()
    go func(ctx context.Context, cancel context.CancelCauseFunc) {
      ticker := time.NewTicker(thrift.ServerConnectivityCheckInterval)
      defer ticker.Stop()
      for {
        select {
        case <-ctx.Done():
          return
        case <-ticker.C:
          if !iprot.Transport().IsOpen() {
            cancel(thrift.ErrAbandonRequest)
            return
          }
        }
      }
    }(tickerCtx, cancel)
  }

  result := RequestGatewayPushServiceOnRequestGatewayPushResult{}
  if err2 := p.handler.OnRequestGatewayPush(ctx, args.Req); err2 != nil {
    tickerCancel()
    err = thrift.WrapTException(err2)
    if errors.Is(err2, thrift.ErrAbandonRequest) {
      return false, thrift.WrapTException(err2)
    }
    if errors.Is(err2, context.Canceled) {
      if err := context.Cause(ctx); errors.Is(err, thrift.ErrAbandonRequest) {
        return false, thrift.WrapTException(err)
      }
    }
    _exc6 := thrift.NewTApplicationException(thrift.INTERNAL_ERROR, "Internal error processing OnRequestGatewayPush: " + err2.Error())
    if err2 := oprot.WriteMessageBegin(ctx, "OnRequestGatewayPush", thrift.EXCEPTION, seqId); err2 != nil {
      _write_err5 = thrift.WrapTException(err2)
    }
    if err2 := _exc6.Write(ctx, oprot); _write_err5 == nil && err2 != nil {
      _write_err5 = thrift.WrapTException(err2)
    }
    if err2 := oprot.WriteMessageEnd(ctx); _write_err5 == nil && err2 != nil {
      _write_err5 = thrift.WrapTException(err2)
    }
    if err2 := oprot.Flush(ctx); _write_err5 == nil && err2 != nil {
      _write_err5 = thrift.WrapTException(err2)
    }
    if _write_err5 != nil {
      return false, thrift.WrapTException(_write_err5)
    }
    return true, err
  }
  tickerCancel()
  if err2 := oprot.WriteMessageBegin(ctx, "OnRequestGatewayPush", thrift.REPLY, seqId); err2 != nil {
    _write_err5 = thrift.WrapTException(err2)
  }
  if err2 := result.Write(ctx, oprot); _write_err5 == nil && err2 != nil {
    _write_err5 = thrift.WrapTException(err2)
  }
  if err2 := oprot.WriteMessageEnd(ctx); _write_err5 == nil && err2 != nil {
    _write_err5 = thrift.WrapTException(err2)
  }
  if err2 := oprot.Flush(ctx); _write_err5 == nil && err2 != nil {
    _write_err5 = thrift.WrapTException(err2)
  }
  if _write_err5 != nil {
    return false, thrift.WrapTException(_write_err5)
  }
  return true, err
}


// HELPER FUNCTIONS AND STRUCTURES

// Attributes:
//  - Req
type RequestGatewayPushServiceOnRequestGatewayPushArgs struct {
  Req *RequestGatewayPush `thrift:"req,1" db:"req" json:"req"`
}

func NewRequestGatewayPushServiceOnRequestGatewayPushArgs() *RequestGatewayPushServiceOnRequestGatewayPushArgs {
  return &RequestGatewayPushServiceOnRequestGatewayPushArgs{}
}

var RequestGatewayPushServiceOnRequestGatewayPushArgs_Req_DEFAULT *RequestGatewayPush
func (p *RequestGatewayPushServiceOnRequestGatewayPushArgs) GetReq() *RequestGatewayPush {
  if !p.IsSetReq() {
    return RequestGatewayPushServiceOnRequestGatewayPushArgs_Req_DEFAULT
  }
return p.Req
}
func (p *RequestGatewayPushServiceOnRequestGatewayPushArgs) IsSetReq() bool {
  return p.Req != nil
}

func (p *RequestGatewayPushServiceOnRequestGatewayPushArgs) Read(ctx context.Context, iprot thrift.TProtocol) error {
  if _, err := iprot.ReadStructBegin(ctx); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T read error: ", p), err)
  }


  for {
    _, fieldTypeId, fieldId, err := iprot.ReadFieldBegin(ctx)
    if err != nil {
      return thrift.PrependError(fmt.Sprintf("%T field %d read error: ", p, fieldId), err)
    }
    if fieldTypeId == thrift.STOP { break; }
    switch fieldId {
    case 1:
      if fieldTypeId == thrift.STRUCT {
        if err := p.ReadField1(ctx, iprot); err != nil {
          return err
        }
      } else {
        if err := iprot.Skip(ctx, fieldTypeId); err != nil {
          return err
        }
      }
    default:
      if err := iprot.Skip(ctx, fieldTypeId); err != nil {
        return err
      }
    }
    if err := iprot.ReadFieldEnd(ctx); err != nil {
      return err
    }
  }
  if err := iprot.ReadStructEnd(ctx); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T read struct end error: ", p), err)
  }
  return nil
}

func (p *RequestGatewayPushServiceOnRequestGatewayPushArgs)  ReadField1(ctx context.Context, iprot thrift.TProtocol) error {
  p.Req = &RequestGatewayPush{}
  if err := p.Req.Read(ctx, iprot); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T error reading struct: ", p.Req), err)
  }
  return nil
}

func (p *RequestGatewayPushServiceOnRequestGatewayPushArgs) Write(ctx context.Context, oprot thrift.TProtocol) error {
  if err := oprot.WriteStructBegin(ctx, "OnRequestGatewayPush_args"); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write struct begin error: ", p), err) }
  if p != nil {
    if err := p.writeField1(ctx, oprot); err != nil { return err }
  }
  if err := oprot.WriteFieldStop(ctx); err != nil {
    return thrift.PrependError("write field stop error: ", err) }
  if err := oprot.WriteStructEnd(ctx); err != nil {
    return thrift.PrependError("write struct stop error: ", err) }
  return nil
}

func (p *RequestGatewayPushServiceOnRequestGatewayPushArgs) writeField1(ctx context.Context, oprot thrift.TProtocol) (err error) {
  if err := oprot.WriteFieldBegin(ctx, "req", thrift.STRUCT, 1); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write field begin error 1:req: ", p), err) }
  if err := p.Req.Write(ctx, oprot); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T error writing struct: ", p.Req), err)
  }
  if err := oprot.WriteFieldEnd(ctx); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write field end error 1:req: ", p), err) }
  return err
}

func (p *RequestGatewayPushServiceOnRequestGatewayPushArgs) String() string {
  if p == nil {
    return "<nil>"
  }
  return fmt.Sprintf("RequestGatewayPushServiceOnRequestGatewayPushArgs(%+v)", *p)
}

type RequestGatewayPushServiceOnRequestGatewayPushResult struct {
}

func NewRequestGatewayPushServiceOnRequestGatewayPushResult() *RequestGatewayPushServiceOnRequestGatewayPushResult {
  return &RequestGatewayPushServiceOnRequestGatewayPushResult{}
}

func (p *RequestGatewayPushServiceOnRequestGatewayPushResult) Read(ctx context.Context, iprot thrift.TProtocol) error {
  if _, err := iprot.ReadStructBegin(ctx); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T read error: ", p), err)
  }


  for {
    _, fieldTypeId, fieldId, err := iprot.ReadFieldBegin(ctx)
    if err != nil {
      return thrift.PrependError(fmt.Sprintf("%T field %d read error: ", p, fieldId), err)
    }
    if fieldTypeId == thrift.STOP { break; }
    if err := iprot.Skip(ctx, fieldTypeId); err != nil {
      return err
    }
    if err := iprot.ReadFieldEnd(ctx); err != nil {
      return err
    }
  }
  if err := iprot.ReadStructEnd(ctx); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T read struct end error: ", p), err)
  }
  return nil
}

func (p *RequestGatewayPushServiceOnRequestGatewayPushResult) Write(ctx context.Context, oprot thrift.TProtocol) error {
  if err := oprot.WriteStructBegin(ctx, "OnRequestGatewayPush_result"); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write struct begin error: ", p), err) }
  if p != nil {
  }
  if err := oprot.WriteFieldStop(ctx); err != nil {
    return thrift.PrependError("write field stop error: ", err) }
  if err := oprot.WriteStructEnd(ctx); err != nil {
    return thrift.PrependError("write struct stop error: ", err) }
  return nil
}

func (p *RequestGatewayPushServiceOnRequestGatewayPushResult) String() string {
  if p == nil {
    return "<nil>"
  }
  return fmt.Sprintf("RequestGatewayPushServiceOnRequestGatewayPushResult(%+v)", *p)
}

