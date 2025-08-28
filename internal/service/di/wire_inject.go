//go:build wireinject
// +build wireinject

package di

import "github.com/google/wire"

// SetupApplication is where we define the dependencies wire will inject
func SetupApplication() (Application, error) {
	panic(wire.Build(wire.NewSet(
		ApiProvider,
		ContextProvider,
		StorageProvider,
		UseCaseProvider,
		wire.NewSet(NewApplication),
	)))
}
