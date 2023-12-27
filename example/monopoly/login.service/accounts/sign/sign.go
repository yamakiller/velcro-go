package sign

import "context"

type Sign interface {
	Init() error
	In(context.Context, string) (*Account, error)
	Out() error
}
