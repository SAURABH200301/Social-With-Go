package main

import (
	"time"

	"github.com/SAURABH200301/Social/internal/auth"
	"github.com/SAURABH200301/Social/internal/db"
	"github.com/SAURABH200301/Social/internal/env"
	"github.com/SAURABH200301/Social/internal/mailer"
	"github.com/SAURABH200301/Social/internal/ratelimiter"
	"github.com/SAURABH200301/Social/internal/store"
	"github.com/SAURABH200301/Social/internal/store/cache"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const version = "0.0.2"

//	@title			SocialWithGO API
//	@description	This is a sample server for a Social Media Application.
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath					localhost:8080/v1
//
// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
// @description				Type "Bearer" followed by a space and JWT token.
func main() {
	cfg := config{
		addr:        env.GetString("ADDR", ":8080"),
		apiURL:      env.GetString("EXTERNAL_URL", "localhost:8080"),
		frontendURL: env.GetString("FRONTEND_URL", "http://localhost:3000"),
		db: dbConfig{
			addr:         env.GetString("DB_ADDR", "postgres://postgres:Test1234@localhost:5432/postgres_db?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
		redisCfg: redisConfig{
			addr:    env.GetString("REDIS_ADDR", "localhost:6379"),
			pw:      env.GetString("REDIS_PW", ""),
			db:      env.GetInt("REDIS_DB", 0),
			enabled: env.GetBool("REDIS_ENABLED", false),
		},
		env: env.GetString("ENV", "development"),
		mail: mailConfig{
			invitationExp: time.Hour * 24 * 3, // 3 days
			fromEmail:     env.GetString("FROM_EMAIL", "noreply@example.com"),
			sendGrid: sendGridConfig{
				apiKey: env.GetString("SENDGRID_API_KEY", ""),
			},
		},
		auth: authConfig{
			basic: basicConfig{
				user: env.GetString("AUTH_BASIC_USER", "admin"),
				pass: env.GetString("AUTH_BASIC_PASS", "admin"),
			},
			token: tokenConfig{
				secret: env.GetString("AUTH_TOKEN_SECRET", ""),
				exp:    time.Hour * 24 * 3,
				iss:    "socialwithgo",
			},
		},
		rateLimiter: ratelimiter.Config{
			RequestsPerTimeFrame: env.GetInt("RATELIMITING_REQUESTS_COUNT", 20),
			TimeFrame:            5 * time.Second,
			Enabled:              env.GetBool("RATE_LIMITER_ENABLED", true),
		},
	}

	//Logger
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()
	logger.Infof("Starting application in %s mode", cfg.env)

	//Database Connection Pool
	db, err := db.New(cfg.db.addr, cfg.db.maxOpenConns, cfg.db.maxIdleConns, cfg.db.maxIdleTime)
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()

	//redis
	var rdb *redis.Client
	if cfg.redisCfg.enabled {
		rdb = cache.NewRedisClient(cfg.redisCfg.addr, cfg.redisCfg.pw, cfg.redisCfg.db)
		logger.Info("Redis connection established")
	}

	//ratelimiting
	rateLimiter := ratelimiter.NewFixedWindowRateLimiter(
		cfg.rateLimiter.RequestsPerTimeFrame,
		cfg.rateLimiter.TimeFrame,
	)

	cacheStorage := cache.NewRedisStorage(rdb)

	logger.Info("Database connection pool established")
	store := store.NewPostgresStorage(db)

	mailerClient := mailer.NewSendGridMailer(
		cfg.mail.sendGrid.apiKey,
		cfg.mail.fromEmail,
	)

	tokenHost := cfg.auth.token.iss
	JWTAuthenicator := auth.NewJWTAuthenicator(cfg.auth.token.secret, tokenHost, tokenHost)
	app := &application{
		config:        cfg,
		store:         store,
		cacheStorage:  cacheStorage,
		logger:        logger,
		mailer:        mailerClient,
		Authonticator: JWTAuthenicator,
		rateLimiter:   rateLimiter,
	}

	mux := app.mount()
	logger.Fatal(app.run(mux))
}
