package sign

type Sign interface {
	Init() error 
	In(string) (*Account, error) 
	Out() error 
}