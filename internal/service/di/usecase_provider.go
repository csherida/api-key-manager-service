package di

import (
	"github.com/csherida/api-key-manager-service/internal/api-key-manager-service/api"
	"github.com/csherida/api-key-manager-service/internal/api-key-manager-service/usecase"
	"github.com/google/wire"
)

var UseCaseProvider = wire.NewSet(
	usecase.NewApiKeyGeneration,
	wire.Bind(new(api.ApiKeyGenerator), new(usecase.ApiKeyGeneration)),
	usecase.NewApiKeyValidation,
	wire.Bind(new(api.ApiKeyValidator), new(usecase.ApiKeyValidation)),
	usecase.NewApiKeyDeletion,
	wire.Bind(new(api.ApiKeyDeleter), new(usecase.ApiKeyDeletion)),
	usecase.NewApiKeyListing,
	wire.Bind(new(api.ApiKeyLister), new(usecase.ApiKeyListing)),
)
