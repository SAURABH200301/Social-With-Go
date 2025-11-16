package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/SAURABH200301/Social/cmd/docs"
	"github.com/SAURABH200301/Social/internal/auth"
	"github.com/SAURABH200301/Social/internal/mailer"
	"github.com/SAURABH200301/Social/internal/ratelimiter"
	"github.com/SAURABH200301/Social/internal/store"
	"github.com/SAURABH200301/Social/internal/store/cache"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	httpSwagger "github.com/swaggo/http-swagger"
)

type application struct {
	config        config
	store         store.Storage
	cacheStorage  cache.Storage
	logger        *zap.SugaredLogger
	mailer        mailer.Client
	Authonticator auth.Authenicator
	rateLimiter   ratelimiter.Limiter
}
type dbConfig struct {
	addr         string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
}
type config struct {
	addr        string
	db          dbConfig
	env         string
	apiURL      string
	mail        mailConfig
	frontendURL string
	auth        authConfig
	redisCfg    redisConfig
	rateLimiter ratelimiter.Config
}

type redisConfig struct {
	addr    string
	pw      string
	db      int
	enabled bool
}

type authConfig struct {
	basic basicConfig
	token tokenConfig
}

type tokenConfig struct {
	secret string
	exp    time.Duration
	iss    string
}

type basicConfig struct {
	user string
	pass string
}

type mailConfig struct {
	invitationExp time.Duration
	fromEmail     string
	sendGrid      sendGridConfig
}
type sendGridConfig struct {
	apiKey string
}

func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	//GLOBAL MIDDLEWARE
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)
	r.Use(app.RateLimiterMiddleware)

	r.Use(middleware.Timeout(60 * time.Second))

	r.Route("/v1", func(r chi.Router) {
		r.With(app.BasicAuthMiddleware()).Get("/health", app.healthCheckHandler)

		host := "http://" + app.config.apiURL
		if host == "" {
			if strings.HasPrefix(host, ":") {
				host = "http://" + host
			}
		}
		r.Get("/swagger/doc.json", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			swaggerFile, err := os.ReadFile("../docs/swagger.yaml")
			if err != nil {
				http.Error(w, "Swagger documentation not found", http.StatusInternalServerError)
				return
			}

			w.Write(swaggerFile)
		})
		docsURL := fmt.Sprintf("%s/v1/swagger/doc.json", host)
		r.Get("/swagger/*", httpSwagger.Handler(
			httpSwagger.URL(docsURL),
		))

		//v1/posts endpoints
		r.Route("/posts", func(r chi.Router) {
			r.Use(app.AuthTokenMiddleware)
			r.Post("/", app.createPostHandler)

			r.Route("/{postID}", func(r chi.Router) {
				//CONSUME MIDDLEWARE
				r.Use(app.postContextMiddleware)

				//INTERNAL ROUTES
				r.Get("/", app.getPostHandler)
				r.Patch("/", app.checkPostOwnershipMiddleware("moderator", app.updatePostHandler))
				r.Delete("/", app.checkPostOwnershipMiddleware("moderator", app.deletePostHandler))
			})
		})
		r.Route("/users", func(r chi.Router) {

			r.Put("/activate/{token}", app.activateUserHandler)

			r.Route("/{userID}", func(r chi.Router) {
				r.Use(app.AuthTokenMiddleware)
				//INTERNAL ROUTES
				r.Get("/", app.getUserHandler)
				// r.Put("/follow", app.followUserHandler)
				// r.Put("/unfollow", app.unfollowUserHandler)
			})

			r.Group(func(r chi.Router) {
				r.Use(app.AuthTokenMiddleware)
				r.Get("/feed", app.getUserFeedHandler)
			})

		})
		r.Route("/authenticate", func(r chi.Router) {
			r.Post("/user", app.registerUserHandler)
			r.Post("/token", app.createTokenHandler)
		})
	})

	return r
}

func (app *application) run(mux http.Handler) error {
	//Docs
	docs.SwaggerInfo.Version = version
	docs.SwaggerInfo.Host = app.config.apiURL
	docs.SwaggerInfo.BasePath = app.config.apiURL + "/v1"

	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: time.Second * 30, // Server will write a response in 30 seconds
		ReadTimeout:  time.Second * 10, // Server will read a request in 10 seconds
		IdleTimeout:  time.Minute,      // Server will idle for 1 minute
	}

	shutdown := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		app.logger.Infow("signal caught", "signal", s.String())

		shutdown <- srv.Shutdown(ctx)
	}()

	app.logger.Infow("Starting server", "addr", app.config.addr, "env", app.config.env)

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	err = <-shutdown
	if err != nil {
		return err
	}
	app.logger.Infow("Stopped server", "addr", app.config.addr, "env", app.config.env)
	return nil
}
