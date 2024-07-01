package errs

import "errors"

var (
	ErrSignAccountOrPass       = errors.New("sign account or password errror")
	ErrPermissionsLost         = errors.New("permissions lost")
	ErrUnRegisterPlayerNoExsit = errors.New("player  no exsit")
)
