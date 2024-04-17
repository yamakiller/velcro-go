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

type GatewayService interface {
  // Parameters:
  //  - Req
  OnRequestGatewayPush(ctx context.Context, req *RequestGatewayPush) (_r *RequestGatewayPush, _err error)
  // Parameters:
  //  - Req
  OnRequestGatewayAlterRule(ctx context.Context, req *RequestGatewayAlterRule) (_r *RequestGatewayAlterRule, _err error)
  // Parameters:
  //  - Req
  OnRequestGatewayCloseClient(ctx context.Context, req *RequestGatewayCloseClient) (_r *RequestGatewayCloseClient, _err error)
}

type GatewayServiceClient struct {
  c thrift.TClient
  meta thrift.ResponseMeta
}

func NewGatewayServiceClientFactory(t thrift.TTransport, f thrift.TProtocolFactory) *GatewayServiceClient {
  return &GatewayServiceClient{
    c: thrift.NewTStandardClient(f.GetProtocol(t), f.GetProtocol(t)),
  }
}

func NewGatewayServiceClientProtocol(t thrift.TTransport, iprot thrift.TProtocol, oprot thrift.TProtocol) *GatewayServiceClient {
  return &GatewayServiceClient{
    c: thrift.NewTStandardClient(iprot, oprot),
  }
}

func NewGatewayServiceClient(c thrift.TClient) *GatewayServiceClient {
  return &GatewayServiceClient{
    c: c,
  }
}

func (p *GatewayServiceClient) Client_() thrift.TClient {
  return p.c
}

func (p *GatewayServiceClient) LastResponseMeta_() thrift.ResponseMeta {
  return p.meta
}

func (p *GatewayServiceClient) SetLastResponseMeta_(meta thrift.ResponseMeta) {
  p.meta = meta
}

// Parameters:
//  - Req
func (p *GatewayServiceClient) OnRequestGatewayPush(ctx context.Context, req *RequestGatewayPush) (_r *RequestGatewayPush, _err error) {
  var _args0 GatewayServiceOnRequestGatewayPushArgs
  _args0.Req = req
  var _result2 GatewayServiceOnRequestGatewayPushResult
  var _meta1 thrift.ResponseMeta
  _meta1, _err = p.Client_().Call(ctx, "OnRequestGatewayPush", &_args0, &_result2)
  p.SetLastResponseMeta_(_meta1)
  if _err != nil {
    return
  }
  if _ret3 := _result2.GetSuccess(); _ret3 != nil {
    return _ret3, nil
  }
  return nil, thrift.NewTApplicationException(thrift.MISSING_RESULT, "OnRequestGatewayPush failed: unknown result")
}

// Parameters:
//  - Req
func (p *GatewayServiceClient) OnRequestGatewayAlterRule(ctx context.Context, req *RequestGatewayAlterRule) (_r *RequestGatewayAlterRule, _err error) {
  var _args4 GatewayServiceOnRequestGatewayAlterRuleArgs
  _args4.Req = req
  var _result6 GatewayServiceOnRequestGatewayAlterRuleResult
  var _meta5 thrift.ResponseMeta
  _meta5, _err = p.Client_().Call(ctx, "OnRequestGatewayAlterRule", &_args4, &_result6)
  p.SetLastResponseMeta_(_meta5)
  if _err != nil {
    return
  }
  if _ret7 := _result6.GetSuccess(); _ret7 != nil {
    return _ret7, nil
  }
  return nil, thrift.NewTApplicationException(thrift.MISSING_RESULT, "OnRequestGatewayAlterRule failed: unknown result")
}

// Parameters:
//  - Req
func (p *GatewayServiceClient) OnRequestGatewayCloseClient(ctx context.Context, req *RequestGatewayCloseClient) (_r *RequestGatewayCloseClient, _err error) {
  var _args8 GatewayServiceOnRequestGatewayCloseClientArgs
  _args8.Req = req
  var _result10 GatewayServiceOnRequestGatewayCloseClientResult
  var _meta9 thrift.ResponseMeta
  _meta9, _err = p.Client_().Call(ctx, "OnRequestGatewayCloseClient", &_args8, &_result10)
  p.SetLastResponseMeta_(_meta9)
  if _err != nil {
    return
  }
  if _ret11 := _result10.GetSuccess(); _ret11 != nil {
    return _ret11, nil
  }
  return nil, thrift.NewTApplicationException(thrift.MISSING_RESULT, "OnRequestGatewayCloseClient failed: unknown result")
}

type GatewayServiceProcessor struct {
  processorMap map[string]thrift.TProcessorFunction
  handler GatewayService
}

func (p *GatewayServiceProcessor) AddToProcessorMap(key string, processor thrift.TProcessorFunction) {
  p.processorMap[key] = processor
}

func (p *GatewayServiceProcessor) GetProcessorFunction(key string) (processor thrift.TProcessorFunction, ok bool) {
  processor, ok = p.processorMap[key]
  return processor, ok
}

func (p *GatewayServiceProcessor) ProcessorMap() map[string]thrift.TProcessorFunction {
  return p.processorMap
}

func NewGatewayServiceProcessor(handler GatewayService) *GatewayServiceProcessor {

  self12 := &GatewayServiceProcessor{handler:handler, processorMap:make(map[string]thrift.TProcessorFunction)}
  self12.processorMap["OnRequestGatewayPush"] = &gatewayServiceProcessorOnRequestGatewayPush{handler:handler}
  self12.processorMap["OnRequestGatewayAlterRule"] = &gatewayServiceProcessorOnRequestGatewayAlterRule{handler:handler}
  self12.processorMap["OnRequestGatewayCloseClient"] = &gatewayServiceProcessorOnRequestGatewayCloseClient{handler:handler}
return self12
}

func (p *GatewayServiceProcessor) Process(ctx context.Context, iprot, oprot thrift.TProtocol) (success bool, err thrift.TException) {
  name, _, seqId, err2 := iprot.ReadMessageBegin(ctx)
  if err2 != nil { return false, thrift.WrapTException(err2) }
  if processor, ok := p.GetProcessorFunction(name); ok {
    return processor.Process(ctx, seqId, iprot, oprot)
  }
  iprot.Skip(ctx, thrift.STRUCT)
  iprot.ReadMessageEnd(ctx)
  x13 := thrift.NewTApplicationException(thrift.UNKNOWN_METHOD, "Unknown function " + name)
  oprot.WriteMessageBegin(ctx, name, thrift.EXCEPTION, seqId)
  x13.Write(ctx, oprot)
  oprot.WriteMessageEnd(ctx)
  oprot.Flush(ctx)
  return false, x13

}

type gatewayServiceProcessorOnRequestGatewayPush struct {
  handler GatewayService
}

func (p *gatewayServiceProcessorOnRequestGatewayPush) Process(ctx context.Context, seqId int32, iprot, oprot thrift.TProtocol) (success bool, err thrift.TException) {
  var _write_err14 error
  args := GatewayServiceOnRequestGatewayPushArgs{}
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

  result := GatewayServiceOnRequestGatewayPushResult{}
  if retval, err2 := p.handler.OnRequestGatewayPush(ctx, args.Req); err2 != nil {
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
    _exc15 := thrift.NewTApplicationException(thrift.INTERNAL_ERROR, "Internal error processing OnRequestGatewayPush: " + err2.Error())
    if err2 := oprot.WriteMessageBegin(ctx, "OnRequestGatewayPush", thrift.EXCEPTION, seqId); err2 != nil {
      _write_err14 = thrift.WrapTException(err2)
    }
    if err2 := _exc15.Write(ctx, oprot); _write_err14 == nil && err2 != nil {
      _write_err14 = thrift.WrapTException(err2)
    }
    if err2 := oprot.WriteMessageEnd(ctx); _write_err14 == nil && err2 != nil {
      _write_err14 = thrift.WrapTException(err2)
    }
    if err2 := oprot.Flush(ctx); _write_err14 == nil && err2 != nil {
      _write_err14 = thrift.WrapTException(err2)
    }
    if _write_err14 != nil {
      return false, thrift.WrapTException(_write_err14)
    }
    return true, err
  } else {
    result.Success = retval
  }
  tickerCancel()
  if err2 := oprot.WriteMessageBegin(ctx, "OnRequestGatewayPush", thrift.REPLY, seqId); err2 != nil {
    _write_err14 = thrift.WrapTException(err2)
  }
  if err2 := result.Write(ctx, oprot); _write_err14 == nil && err2 != nil {
    _write_err14 = thrift.WrapTException(err2)
  }
  if err2 := oprot.WriteMessageEnd(ctx); _write_err14 == nil && err2 != nil {
    _write_err14 = thrift.WrapTException(err2)
  }
  if err2 := oprot.Flush(ctx); _write_err14 == nil && err2 != nil {
    _write_err14 = thrift.WrapTException(err2)
  }
  if _write_err14 != nil {
    return false, thrift.WrapTException(_write_err14)
  }
  return true, err
}

type gatewayServiceProcessorOnRequestGatewayAlterRule struct {
  handler GatewayService
}

func (p *gatewayServiceProcessorOnRequestGatewayAlterRule) Process(ctx context.Context, seqId int32, iprot, oprot thrift.TProtocol) (success bool, err thrift.TException) {
  var _write_err16 error
  args := GatewayServiceOnRequestGatewayAlterRuleArgs{}
  if err2 := args.Read(ctx, iprot); err2 != nil {
    iprot.ReadMessageEnd(ctx)
    x := thrift.NewTApplicationException(thrift.PROTOCOL_ERROR, err2.Error())
    oprot.WriteMessageBegin(ctx, "OnRequestGatewayAlterRule", thrift.EXCEPTION, seqId)
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

  result := GatewayServiceOnRequestGatewayAlterRuleResult{}
  if retval, err2 := p.handler.OnRequestGatewayAlterRule(ctx, args.Req); err2 != nil {
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
    _exc17 := thrift.NewTApplicationException(thrift.INTERNAL_ERROR, "Internal error processing OnRequestGatewayAlterRule: " + err2.Error())
    if err2 := oprot.WriteMessageBegin(ctx, "OnRequestGatewayAlterRule", thrift.EXCEPTION, seqId); err2 != nil {
      _write_err16 = thrift.WrapTException(err2)
    }
    if err2 := _exc17.Write(ctx, oprot); _write_err16 == nil && err2 != nil {
      _write_err16 = thrift.WrapTException(err2)
    }
    if err2 := oprot.WriteMessageEnd(ctx); _write_err16 == nil && err2 != nil {
      _write_err16 = thrift.WrapTException(err2)
    }
    if err2 := oprot.Flush(ctx); _write_err16 == nil && err2 != nil {
      _write_err16 = thrift.WrapTException(err2)
    }
    if _write_err16 != nil {
      return false, thrift.WrapTException(_write_err16)
    }
    return true, err
  } else {
    result.Success = retval
  }
  tickerCancel()
  if err2 := oprot.WriteMessageBegin(ctx, "OnRequestGatewayAlterRule", thrift.REPLY, seqId); err2 != nil {
    _write_err16 = thrift.WrapTException(err2)
  }
  if err2 := result.Write(ctx, oprot); _write_err16 == nil && err2 != nil {
    _write_err16 = thrift.WrapTException(err2)
  }
  if err2 := oprot.WriteMessageEnd(ctx); _write_err16 == nil && err2 != nil {
    _write_err16 = thrift.WrapTException(err2)
  }
  if err2 := oprot.Flush(ctx); _write_err16 == nil && err2 != nil {
    _write_err16 = thrift.WrapTException(err2)
  }
  if _write_err16 != nil {
    return false, thrift.WrapTException(_write_err16)
  }
  return true, err
}

type gatewayServiceProcessorOnRequestGatewayCloseClient struct {
  handler GatewayService
}

func (p *gatewayServiceProcessorOnRequestGatewayCloseClient) Process(ctx context.Context, seqId int32, iprot, oprot thrift.TProtocol) (success bool, err thrift.TException) {
  var _write_err18 error
  args := GatewayServiceOnRequestGatewayCloseClientArgs{}
  if err2 := args.Read(ctx, iprot); err2 != nil {
    iprot.ReadMessageEnd(ctx)
    x := thrift.NewTApplicationException(thrift.PROTOCOL_ERROR, err2.Error())
    oprot.WriteMessageBegin(ctx, "OnRequestGatewayCloseClient", thrift.EXCEPTION, seqId)
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

  result := GatewayServiceOnRequestGatewayCloseClientResult{}
  if retval, err2 := p.handler.OnRequestGatewayCloseClient(ctx, args.Req); err2 != nil {
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
    _exc19 := thrift.NewTApplicationException(thrift.INTERNAL_ERROR, "Internal error processing OnRequestGatewayCloseClient: " + err2.Error())
    if err2 := oprot.WriteMessageBegin(ctx, "OnRequestGatewayCloseClient", thrift.EXCEPTION, seqId); err2 != nil {
      _write_err18 = thrift.WrapTException(err2)
    }
    if err2 := _exc19.Write(ctx, oprot); _write_err18 == nil && err2 != nil {
      _write_err18 = thrift.WrapTException(err2)
    }
    if err2 := oprot.WriteMessageEnd(ctx); _write_err18 == nil && err2 != nil {
      _write_err18 = thrift.WrapTException(err2)
    }
    if err2 := oprot.Flush(ctx); _write_err18 == nil && err2 != nil {
      _write_err18 = thrift.WrapTException(err2)
    }
    if _write_err18 != nil {
      return false, thrift.WrapTException(_write_err18)
    }
    return true, err
  } else {
    result.Success = retval
  }
  tickerCancel()
  if err2 := oprot.WriteMessageBegin(ctx, "OnRequestGatewayCloseClient", thrift.REPLY, seqId); err2 != nil {
    _write_err18 = thrift.WrapTException(err2)
  }
  if err2 := result.Write(ctx, oprot); _write_err18 == nil && err2 != nil {
    _write_err18 = thrift.WrapTException(err2)
  }
  if err2 := oprot.WriteMessageEnd(ctx); _write_err18 == nil && err2 != nil {
    _write_err18 = thrift.WrapTException(err2)
  }
  if err2 := oprot.Flush(ctx); _write_err18 == nil && err2 != nil {
    _write_err18 = thrift.WrapTException(err2)
  }
  if _write_err18 != nil {
    return false, thrift.WrapTException(_write_err18)
  }
  return true, err
}


// HELPER FUNCTIONS AND STRUCTURES

// Attributes:
//  - Req
type GatewayServiceOnRequestGatewayPushArgs struct {
  Req *RequestGatewayPush `thrift:"req,1" db:"req" json:"req"`
}

func NewGatewayServiceOnRequestGatewayPushArgs() *GatewayServiceOnRequestGatewayPushArgs {
  return &GatewayServiceOnRequestGatewayPushArgs{}
}

var GatewayServiceOnRequestGatewayPushArgs_Req_DEFAULT *RequestGatewayPush
func (p *GatewayServiceOnRequestGatewayPushArgs) GetReq() *RequestGatewayPush {
  if !p.IsSetReq() {
    return GatewayServiceOnRequestGatewayPushArgs_Req_DEFAULT
  }
return p.Req
}
func (p *GatewayServiceOnRequestGatewayPushArgs) IsSetReq() bool {
  return p.Req != nil
}

func (p *GatewayServiceOnRequestGatewayPushArgs) Read(ctx context.Context, iprot thrift.TProtocol) error {
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

func (p *GatewayServiceOnRequestGatewayPushArgs)  ReadField1(ctx context.Context, iprot thrift.TProtocol) error {
  p.Req = &RequestGatewayPush{}
  if err := p.Req.Read(ctx, iprot); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T error reading struct: ", p.Req), err)
  }
  return nil
}

func (p *GatewayServiceOnRequestGatewayPushArgs) Write(ctx context.Context, oprot thrift.TProtocol) error {
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

func (p *GatewayServiceOnRequestGatewayPushArgs) writeField1(ctx context.Context, oprot thrift.TProtocol) (err error) {
  if err := oprot.WriteFieldBegin(ctx, "req", thrift.STRUCT, 1); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write field begin error 1:req: ", p), err) }
  if err := p.Req.Write(ctx, oprot); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T error writing struct: ", p.Req), err)
  }
  if err := oprot.WriteFieldEnd(ctx); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write field end error 1:req: ", p), err) }
  return err
}

func (p *GatewayServiceOnRequestGatewayPushArgs) String() string {
  if p == nil {
    return "<nil>"
  }
  return fmt.Sprintf("GatewayServiceOnRequestGatewayPushArgs(%+v)", *p)
}

// Attributes:
//  - Success
type GatewayServiceOnRequestGatewayPushResult struct {
  Success *RequestGatewayPush `thrift:"success,0" db:"success" json:"success,omitempty"`
}

func NewGatewayServiceOnRequestGatewayPushResult() *GatewayServiceOnRequestGatewayPushResult {
  return &GatewayServiceOnRequestGatewayPushResult{}
}

var GatewayServiceOnRequestGatewayPushResult_Success_DEFAULT *RequestGatewayPush
func (p *GatewayServiceOnRequestGatewayPushResult) GetSuccess() *RequestGatewayPush {
  if !p.IsSetSuccess() {
    return GatewayServiceOnRequestGatewayPushResult_Success_DEFAULT
  }
return p.Success
}
func (p *GatewayServiceOnRequestGatewayPushResult) IsSetSuccess() bool {
  return p.Success != nil
}

func (p *GatewayServiceOnRequestGatewayPushResult) Read(ctx context.Context, iprot thrift.TProtocol) error {
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
    case 0:
      if fieldTypeId == thrift.STRUCT {
        if err := p.ReadField0(ctx, iprot); err != nil {
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

func (p *GatewayServiceOnRequestGatewayPushResult)  ReadField0(ctx context.Context, iprot thrift.TProtocol) error {
  p.Success = &RequestGatewayPush{}
  if err := p.Success.Read(ctx, iprot); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T error reading struct: ", p.Success), err)
  }
  return nil
}

func (p *GatewayServiceOnRequestGatewayPushResult) Write(ctx context.Context, oprot thrift.TProtocol) error {
  if err := oprot.WriteStructBegin(ctx, "OnRequestGatewayPush_result"); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write struct begin error: ", p), err) }
  if p != nil {
    if err := p.writeField0(ctx, oprot); err != nil { return err }
  }
  if err := oprot.WriteFieldStop(ctx); err != nil {
    return thrift.PrependError("write field stop error: ", err) }
  if err := oprot.WriteStructEnd(ctx); err != nil {
    return thrift.PrependError("write struct stop error: ", err) }
  return nil
}

func (p *GatewayServiceOnRequestGatewayPushResult) writeField0(ctx context.Context, oprot thrift.TProtocol) (err error) {
  if p.IsSetSuccess() {
    if err := oprot.WriteFieldBegin(ctx, "success", thrift.STRUCT, 0); err != nil {
      return thrift.PrependError(fmt.Sprintf("%T write field begin error 0:success: ", p), err) }
    if err := p.Success.Write(ctx, oprot); err != nil {
      return thrift.PrependError(fmt.Sprintf("%T error writing struct: ", p.Success), err)
    }
    if err := oprot.WriteFieldEnd(ctx); err != nil {
      return thrift.PrependError(fmt.Sprintf("%T write field end error 0:success: ", p), err) }
  }
  return err
}

func (p *GatewayServiceOnRequestGatewayPushResult) String() string {
  if p == nil {
    return "<nil>"
  }
  return fmt.Sprintf("GatewayServiceOnRequestGatewayPushResult(%+v)", *p)
}

// Attributes:
//  - Req
type GatewayServiceOnRequestGatewayAlterRuleArgs struct {
  Req *RequestGatewayAlterRule `thrift:"req,1" db:"req" json:"req"`
}

func NewGatewayServiceOnRequestGatewayAlterRuleArgs() *GatewayServiceOnRequestGatewayAlterRuleArgs {
  return &GatewayServiceOnRequestGatewayAlterRuleArgs{}
}

var GatewayServiceOnRequestGatewayAlterRuleArgs_Req_DEFAULT *RequestGatewayAlterRule
func (p *GatewayServiceOnRequestGatewayAlterRuleArgs) GetReq() *RequestGatewayAlterRule {
  if !p.IsSetReq() {
    return GatewayServiceOnRequestGatewayAlterRuleArgs_Req_DEFAULT
  }
return p.Req
}
func (p *GatewayServiceOnRequestGatewayAlterRuleArgs) IsSetReq() bool {
  return p.Req != nil
}

func (p *GatewayServiceOnRequestGatewayAlterRuleArgs) Read(ctx context.Context, iprot thrift.TProtocol) error {
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

func (p *GatewayServiceOnRequestGatewayAlterRuleArgs)  ReadField1(ctx context.Context, iprot thrift.TProtocol) error {
  p.Req = &RequestGatewayAlterRule{}
  if err := p.Req.Read(ctx, iprot); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T error reading struct: ", p.Req), err)
  }
  return nil
}

func (p *GatewayServiceOnRequestGatewayAlterRuleArgs) Write(ctx context.Context, oprot thrift.TProtocol) error {
  if err := oprot.WriteStructBegin(ctx, "OnRequestGatewayAlterRule_args"); err != nil {
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

func (p *GatewayServiceOnRequestGatewayAlterRuleArgs) writeField1(ctx context.Context, oprot thrift.TProtocol) (err error) {
  if err := oprot.WriteFieldBegin(ctx, "req", thrift.STRUCT, 1); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write field begin error 1:req: ", p), err) }
  if err := p.Req.Write(ctx, oprot); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T error writing struct: ", p.Req), err)
  }
  if err := oprot.WriteFieldEnd(ctx); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write field end error 1:req: ", p), err) }
  return err
}

func (p *GatewayServiceOnRequestGatewayAlterRuleArgs) String() string {
  if p == nil {
    return "<nil>"
  }
  return fmt.Sprintf("GatewayServiceOnRequestGatewayAlterRuleArgs(%+v)", *p)
}

// Attributes:
//  - Success
type GatewayServiceOnRequestGatewayAlterRuleResult struct {
  Success *RequestGatewayAlterRule `thrift:"success,0" db:"success" json:"success,omitempty"`
}

func NewGatewayServiceOnRequestGatewayAlterRuleResult() *GatewayServiceOnRequestGatewayAlterRuleResult {
  return &GatewayServiceOnRequestGatewayAlterRuleResult{}
}

var GatewayServiceOnRequestGatewayAlterRuleResult_Success_DEFAULT *RequestGatewayAlterRule
func (p *GatewayServiceOnRequestGatewayAlterRuleResult) GetSuccess() *RequestGatewayAlterRule {
  if !p.IsSetSuccess() {
    return GatewayServiceOnRequestGatewayAlterRuleResult_Success_DEFAULT
  }
return p.Success
}
func (p *GatewayServiceOnRequestGatewayAlterRuleResult) IsSetSuccess() bool {
  return p.Success != nil
}

func (p *GatewayServiceOnRequestGatewayAlterRuleResult) Read(ctx context.Context, iprot thrift.TProtocol) error {
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
    case 0:
      if fieldTypeId == thrift.STRUCT {
        if err := p.ReadField0(ctx, iprot); err != nil {
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

func (p *GatewayServiceOnRequestGatewayAlterRuleResult)  ReadField0(ctx context.Context, iprot thrift.TProtocol) error {
  p.Success = &RequestGatewayAlterRule{}
  if err := p.Success.Read(ctx, iprot); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T error reading struct: ", p.Success), err)
  }
  return nil
}

func (p *GatewayServiceOnRequestGatewayAlterRuleResult) Write(ctx context.Context, oprot thrift.TProtocol) error {
  if err := oprot.WriteStructBegin(ctx, "OnRequestGatewayAlterRule_result"); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write struct begin error: ", p), err) }
  if p != nil {
    if err := p.writeField0(ctx, oprot); err != nil { return err }
  }
  if err := oprot.WriteFieldStop(ctx); err != nil {
    return thrift.PrependError("write field stop error: ", err) }
  if err := oprot.WriteStructEnd(ctx); err != nil {
    return thrift.PrependError("write struct stop error: ", err) }
  return nil
}

func (p *GatewayServiceOnRequestGatewayAlterRuleResult) writeField0(ctx context.Context, oprot thrift.TProtocol) (err error) {
  if p.IsSetSuccess() {
    if err := oprot.WriteFieldBegin(ctx, "success", thrift.STRUCT, 0); err != nil {
      return thrift.PrependError(fmt.Sprintf("%T write field begin error 0:success: ", p), err) }
    if err := p.Success.Write(ctx, oprot); err != nil {
      return thrift.PrependError(fmt.Sprintf("%T error writing struct: ", p.Success), err)
    }
    if err := oprot.WriteFieldEnd(ctx); err != nil {
      return thrift.PrependError(fmt.Sprintf("%T write field end error 0:success: ", p), err) }
  }
  return err
}

func (p *GatewayServiceOnRequestGatewayAlterRuleResult) String() string {
  if p == nil {
    return "<nil>"
  }
  return fmt.Sprintf("GatewayServiceOnRequestGatewayAlterRuleResult(%+v)", *p)
}

// Attributes:
//  - Req
type GatewayServiceOnRequestGatewayCloseClientArgs struct {
  Req *RequestGatewayCloseClient `thrift:"req,1" db:"req" json:"req"`
}

func NewGatewayServiceOnRequestGatewayCloseClientArgs() *GatewayServiceOnRequestGatewayCloseClientArgs {
  return &GatewayServiceOnRequestGatewayCloseClientArgs{}
}

var GatewayServiceOnRequestGatewayCloseClientArgs_Req_DEFAULT *RequestGatewayCloseClient
func (p *GatewayServiceOnRequestGatewayCloseClientArgs) GetReq() *RequestGatewayCloseClient {
  if !p.IsSetReq() {
    return GatewayServiceOnRequestGatewayCloseClientArgs_Req_DEFAULT
  }
return p.Req
}
func (p *GatewayServiceOnRequestGatewayCloseClientArgs) IsSetReq() bool {
  return p.Req != nil
}

func (p *GatewayServiceOnRequestGatewayCloseClientArgs) Read(ctx context.Context, iprot thrift.TProtocol) error {
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

func (p *GatewayServiceOnRequestGatewayCloseClientArgs)  ReadField1(ctx context.Context, iprot thrift.TProtocol) error {
  p.Req = &RequestGatewayCloseClient{}
  if err := p.Req.Read(ctx, iprot); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T error reading struct: ", p.Req), err)
  }
  return nil
}

func (p *GatewayServiceOnRequestGatewayCloseClientArgs) Write(ctx context.Context, oprot thrift.TProtocol) error {
  if err := oprot.WriteStructBegin(ctx, "OnRequestGatewayCloseClient_args"); err != nil {
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

func (p *GatewayServiceOnRequestGatewayCloseClientArgs) writeField1(ctx context.Context, oprot thrift.TProtocol) (err error) {
  if err := oprot.WriteFieldBegin(ctx, "req", thrift.STRUCT, 1); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write field begin error 1:req: ", p), err) }
  if err := p.Req.Write(ctx, oprot); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T error writing struct: ", p.Req), err)
  }
  if err := oprot.WriteFieldEnd(ctx); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write field end error 1:req: ", p), err) }
  return err
}

func (p *GatewayServiceOnRequestGatewayCloseClientArgs) String() string {
  if p == nil {
    return "<nil>"
  }
  return fmt.Sprintf("GatewayServiceOnRequestGatewayCloseClientArgs(%+v)", *p)
}

// Attributes:
//  - Success
type GatewayServiceOnRequestGatewayCloseClientResult struct {
  Success *RequestGatewayCloseClient `thrift:"success,0" db:"success" json:"success,omitempty"`
}

func NewGatewayServiceOnRequestGatewayCloseClientResult() *GatewayServiceOnRequestGatewayCloseClientResult {
  return &GatewayServiceOnRequestGatewayCloseClientResult{}
}

var GatewayServiceOnRequestGatewayCloseClientResult_Success_DEFAULT *RequestGatewayCloseClient
func (p *GatewayServiceOnRequestGatewayCloseClientResult) GetSuccess() *RequestGatewayCloseClient {
  if !p.IsSetSuccess() {
    return GatewayServiceOnRequestGatewayCloseClientResult_Success_DEFAULT
  }
return p.Success
}
func (p *GatewayServiceOnRequestGatewayCloseClientResult) IsSetSuccess() bool {
  return p.Success != nil
}

func (p *GatewayServiceOnRequestGatewayCloseClientResult) Read(ctx context.Context, iprot thrift.TProtocol) error {
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
    case 0:
      if fieldTypeId == thrift.STRUCT {
        if err := p.ReadField0(ctx, iprot); err != nil {
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

func (p *GatewayServiceOnRequestGatewayCloseClientResult)  ReadField0(ctx context.Context, iprot thrift.TProtocol) error {
  p.Success = &RequestGatewayCloseClient{}
  if err := p.Success.Read(ctx, iprot); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T error reading struct: ", p.Success), err)
  }
  return nil
}

func (p *GatewayServiceOnRequestGatewayCloseClientResult) Write(ctx context.Context, oprot thrift.TProtocol) error {
  if err := oprot.WriteStructBegin(ctx, "OnRequestGatewayCloseClient_result"); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write struct begin error: ", p), err) }
  if p != nil {
    if err := p.writeField0(ctx, oprot); err != nil { return err }
  }
  if err := oprot.WriteFieldStop(ctx); err != nil {
    return thrift.PrependError("write field stop error: ", err) }
  if err := oprot.WriteStructEnd(ctx); err != nil {
    return thrift.PrependError("write struct stop error: ", err) }
  return nil
}

func (p *GatewayServiceOnRequestGatewayCloseClientResult) writeField0(ctx context.Context, oprot thrift.TProtocol) (err error) {
  if p.IsSetSuccess() {
    if err := oprot.WriteFieldBegin(ctx, "success", thrift.STRUCT, 0); err != nil {
      return thrift.PrependError(fmt.Sprintf("%T write field begin error 0:success: ", p), err) }
    if err := p.Success.Write(ctx, oprot); err != nil {
      return thrift.PrependError(fmt.Sprintf("%T error writing struct: ", p.Success), err)
    }
    if err := oprot.WriteFieldEnd(ctx); err != nil {
      return thrift.PrependError(fmt.Sprintf("%T write field end error 0:success: ", p), err) }
  }
  return err
}

func (p *GatewayServiceOnRequestGatewayCloseClientResult) String() string {
  if p == nil {
    return "<nil>"
  }
  return fmt.Sprintf("GatewayServiceOnRequestGatewayCloseClientResult(%+v)", *p)
}


