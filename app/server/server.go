package server

import (
	"context"
	"crypto/subtle"
	"io"
	"net/http"
	"time"

	"github.com/didip/tollbooth/v6"
	"github.com/didip/tollbooth_chi"
	"github.com/go-chi/chi/v5"
	log "github.com/go-pkgz/lgr"
	"github.com/go-pkgz/rest"
)

//go:generate moq -out mocks/config.go -pkg mocks -skip-ensure -fmt goimports . Config
//go:generate moq -out mocks/runner.go -pkg mocks -skip-ensure -fmt goimports . Runner

// Rest implement http api invoking remote execution for requested tasks
type Rest struct {
	Listen    string
	Version   string
	SecretKey string
	Config    Config
	Runner    Runner
}

// Config declares command loader from config for given tasks
type Config interface {
	GetTaskCommand(name string) (command string, ok bool)
}

// Runner executes commands
type Runner interface {
	Run(ctx context.Context, command string, logWriter io.Writer) error
}

// Run starts http server and closes on context cancellation
func (s *Rest) Run(ctx context.Context) error {
	log.Printf("[INFO] start http server on %s", s.Listen)

	httpServer := &http.Server{
		Addr:              s.Listen,
		Handler:           s.router(),
		ReadHeaderTimeout: time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       time.Second,
		ErrorLog:          log.ToStdLogger(log.Default(), "WARN"),
	}

	go func() {
		<-ctx.Done()
		if httpServer != nil {
			if err := httpServer.Close(); err != nil {
				log.Printf("[ERROR] failed to close http server, %v", err)
			}
		}

	}()

	return httpServer.ListenAndServe()
}

func (s *Rest) router() http.Handler {
	router := chi.NewRouter()
	router.Use(rest.Recoverer(log.Default()))
	router.Use(rest.AppInfo("updater", "umputun", s.Version))
	router.Use(rest.Ping)
	router.Use(tollbooth_chi.LimitHandler(tollbooth.NewLimiter(10, nil)))
	router.Get("/update/{task}/{key}", s.taskCtrl)
	return router
}

// GET /update/{task}/{key}
func (s *Rest) taskCtrl(w http.ResponseWriter, r *http.Request) {
	taskName := chi.URLParam(r, "task")
	key := chi.URLParam(r, "key")
	if subtle.ConstantTimeCompare([]byte(key), []byte(s.SecretKey)) != 1 {
		http.Error(w, "rejected", http.StatusForbidden)
		return
	}
	command, ok := s.Config.GetTaskCommand(taskName)
	if !ok {
		http.Error(w, "unknown command", http.StatusBadRequest)
		return
	}
	log.Printf("[DEBUG] request for task %s", taskName)
	if err := s.Runner.Run(r.Context(), command, log.ToWriter(log.Default(), ">")); err != nil {
		http.Error(w, "failed command", http.StatusInternalServerError)
		return
	}
}
