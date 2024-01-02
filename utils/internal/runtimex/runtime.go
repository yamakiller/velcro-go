package runtimex

import (
	_ "unsafe" // for linkname
)

//go:linkname Fastrand runtime.fastrand
func Fastrand() uint32
