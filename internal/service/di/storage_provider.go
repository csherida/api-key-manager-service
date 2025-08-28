package di

import (
	"github.com/csherida/api-key-manager-service/internal/api-key-manager-service/infra"
	"github.com/csherida/api-key-manager-service/internal/api-key-manager-service/usecase"
	"github.com/google/wire"
)

var StorageProvider = wire.NewSet( //nolint:gochecknoglobals
	infra.NewDataStore,
	wire.Bind(new(usecase.Repository), new(*infra.DataStore)),
)
