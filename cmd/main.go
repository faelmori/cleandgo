package main

import (
	gl "github.com/faelmori/cleandgo/logger"
	l "github.com/faelmori/logz"
)

var logger l.Logger

// main initializes the logger and creates a new CleandGO instance.
func main() {
	if err := RegX().Command().Execute(); err != nil {
		gl.Log("fatal", err.Error())
	}
}
