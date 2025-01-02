package server

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/didip/tollbooth/v8"
	log "github.com/go-pkgz/lgr"
	"github.com/go-pkgz/rest"
	"github.com/go-pkgz/routegroup"
)

//go:generate moq -out mocks/config.go -pkg mocks -skip-ensure -fmt goimports . Config
//go:generate moq -out mocks/runner.go -pkg mocks -skip-ensure -fmt goimports . Runner

// Rest implement http api invoking remote execution for requested tasks
type Rest struct {
	Listen      string
	Version     string
	SecretKey   string
	Config      Config
	Runner      Runner
	UpdateDelay time.Duration
	Timeout     time.Duration
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
		WriteTimeout:      s.Timeout + 10*time.Second, // Give enough time to finish the task and respond
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
	router := routegroup.New(http.NewServeMux())
	router.Use(rest.Recoverer(log.Default()))
	router.Use(rest.Throttle(100)) // limit the total number of the running requests
	router.Use(rest.AppInfo("updater", "umputun", s.Version))
	router.Use(rest.Ping)
	router.Use(tollbooth.HTTPMiddleware(tollbooth.NewLimiter(10, nil)))
	if s.UpdateDelay > 0 {
		router.Use(s.slowMiddleware)
	}

	router.HandleFunc("GET /update/{task}/{key}", s.taskCtrl)
	router.HandleFunc("POST /update", s.taskPostCtrl)
	return router
}

// GET /update/{task}/{key}?async=[0|1]
func (s *Rest) taskCtrl(w http.ResponseWriter, r *http.Request) {
	taskName := r.PathValue("task")
	key := r.PathValue("key")
	isAsync := r.URL.Query().Get("async") == "1" || r.URL.Query().Get("async") == "yes"
	s.execTask(w, r, key, taskName, isAsync)
}

// POST /update
func (s *Rest) taskPostCtrl(w http.ResponseWriter, r *http.Request) {
	req := struct {
		Task   string `json:"task"`
		Secret string `json:"secret"`
		Async  bool   `json:"async"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "failed to decode request", http.StatusBadRequest)
		return
	}
	if req.Task == "" || req.Secret == "" {
		http.Error(w, "task and secret required", http.StatusBadRequest)
		return
	}
	s.execTask(w, r, req.Secret, req.Task, req.Async)
}

func (s *Rest) execTask(w http.ResponseWriter, r *http.Request, secret, taskName string, isAsync bool) {
	if subtle.ConstantTimeCompare([]byte(secret), []byte(s.SecretKey)) != 1 {
		http.Error(w, "rejected", http.StatusForbidden)
		return
	}

	command, ok := s.Config.GetTaskCommand(taskName)
	if !ok {
		http.Error(w, "unknown command", http.StatusBadRequest)
		return
	}

	log.Printf("[INFO] invoke task %s", taskName)

	if isAsync {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), s.Timeout)
			defer cancel()
			if err := s.Runner.Run(ctx, command, log.ToWriter(log.Default(), ">")); err != nil {
				log.Printf("[WARN] failed command")
				return
			}
		}()
		rest.RenderJSON(w, rest.JSON{"submitted": "ok", "task": taskName})
		return
	}

	if err := s.Runner.Run(r.Context(), command, log.ToWriter(log.Default(), ">")); err != nil {
		http.Error(w, "failed command", http.StatusInternalServerError)
		return
	}

	rest.RenderJSON(w, rest.JSON{"updated": "ok", "task": taskName})
}

// middleware for slowing requests downs
func (s *Rest) slowMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(s.UpdateDelay)
		next.ServeHTTP(w, r)
	})
}
