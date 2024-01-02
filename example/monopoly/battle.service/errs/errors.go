package errs

import "errors"

var (
	ErrorPlayerAlreadyInBattleSpace = errors.New("Player is already in a certain battle space and cannot create a separate battle space. Please exit other spaces before creating one.")
	ErrorPlayerOnlineDataLost       = errors.New("Player online data lost")
)
