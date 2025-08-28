package di

import (
	"context"
	"github.com/csherida/api-key-manager-service/internal/api-key-manager-service/api"
	"github.com/csherida/api-key-manager-service/internal/api-key-manager-service/usecase"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

const _serverPort = 8080

type Application struct {
	ctx                  context.Context
	cancel               context.CancelFunc
	keyGeneratorHandler  api.ApiKeyGeneratorHandlerType
	keyValidationHandler func(http.ResponseWriter, *http.Request)
	keyDeletionHandler   func(http.ResponseWriter, *http.Request)
	repo                 usecase.Repository
}

func NewApplication(
	ctx context.Context,
	keyGeneratorHandler api.ApiKeyGeneratorHandler,
	keyValidationHandler api.ApiKeyValidationHandler,
	keyDeletionHandler api.ApiKeyDeletionHandler,
) Application {
	appCtx, cancel := context.WithCancel(ctx)
	app := Application{
		ctx:                  appCtx,
		cancel:               cancel,
		keyGeneratorHandler:  keyGeneratorHandler.ApiKeyGenerator,
		keyValidationHandler: keyValidationHandler.ValidateApiKey,
		keyDeletionHandler:   keyDeletionHandler.DeleteApiKey,
	}
	return app
}

func NewContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	// wait for a termination signal from the OS
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-quit

		log.Printf("received an OS signal - shutting down: %v\n", sig)
		cancel()
	}()

	return ctx
}

func (app *Application) Run() error {
	///TODO: move implementation to infra folder and better server handling
	router := mux.NewRouter()
	router.HandleFunc("/keys", app.keyGeneratorHandler).Methods("POST")
	router.HandleFunc("/keys/{keyId}", app.keyDeletionHandler).Methods("DELETE")
	router.HandleFunc("/keys/validate", app.keyValidationHandler).Methods("POST")

	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:" + strconv.Itoa(_serverPort)},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	srv := &http.Server{
		Addr:    ":" + strconv.Itoa(_serverPort),
		Handler: corsHandler.Handler(router),
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen and serve: %s\n", err)
		}
	}()
	log.Print("Server Started")

	<-done
	log.Print("Server Stopped")

	if err := srv.Shutdown(app.ctx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}
	log.Print("Server Exited Properly")

	return nil
}

func (app *Application) CancelContext() {
	app.cancel()
}
