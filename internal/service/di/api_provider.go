package di

import (
	"github.com/csherida/api-key-manager-service/internal/api-key-manager-service/api"
	"github.com/google/wire"
)

var ApiProvider = wire.NewSet(
	api.NewApiKeyGeneratorHandler,
	api.NewApiKeyValidationHandler,
	api.NewApiKeyDeletionHandler,
)
