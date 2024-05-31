package errs

import "errors"

var (
	ErrorPlayerAlreadyInBattleSpace = errors.New("Player is already in a certain battle space and cannot create a separate battle space. Please exit other spaces before creating one.")
	ErrorPlayerOnlineDataLost       = errors.New("Player online data lost")
	ErrorPermissionsLost            = errors.New("permissions lost")
	ErrorPlayerIsNotInBattleSpace   = errors.New("Player not in battle space")
	ErrorSpacePlayerIsFull          = errors.New("Battle space is full")
	ErrorPlayerRepeatOperation      = errors.New("Player repeat operation")
	ErrorSpaceOnlineDataLost        = errors.New("Space online data lost")
	ErrorSpaceIsRunning             = errors.New("Space is running")
	ErrorSpacePassword              = errors.New("Space password error")
)
