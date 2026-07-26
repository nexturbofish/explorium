package app

import (
	"hi/internal/appstate"
	"hi/internal/reflect"
)

type (
	AppState       = appstate.AppState
	ReflectOptions = reflect.ReflectOptions
)

var (
	InitApp    = appstate.InitApp
	NewSession = appstate.NewSession
)
