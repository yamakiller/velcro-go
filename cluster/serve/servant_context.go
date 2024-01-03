package serve

import "context"

type ctxServantInfoKeyType struct{}

var ctxServantInfoKey ctxServantInfoKeyType

func NewCtxWithServantClientInfo(ctx context.Context, si *ServantClientInfo) context.Context {
	if si != nil {
		return context.WithValue(ctx, ctxServantInfoKey, si)
	}
	return ctx
}

func FreeCtxWithServantClientInfo(ctx context.Context) context.Context {
	ci := GetServantClientInfo(ctx)
	if ci == nil {
		return ctx
	}

	ci.Recycle()
	ci = nil

	return context.WithValue(ctx, ctxServantInfoKey, ci)
}

func GetServantClientInfo(ctx context.Context) *ServantClientInfo {
	if ri, ok := ctx.Value(ctxServantInfoKey).(*ServantClientInfo); ok {
		return ri
	}
	return nil
}
