package di

import "github.com/google/wire"

var ContextProvider = wire.NewSet(
	NewContext,
)
