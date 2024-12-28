package main

import (
	"expvar"
	"runtime"
	"time"

	"github.com/ekachaikeaw/social/internal/auth"
	"github.com/ekachaikeaw/social/internal/db"
	"github.com/ekachaikeaw/social/internal/env"
	"github.com/ekachaikeaw/social/internal/mailer"
	"github.com/ekachaikeaw/social/internal/ratelimiter"
	"github.com/ekachaikeaw/social/internal/store"
	"github.com/ekachaikeaw/social/internal/store/cache"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

const version = "1.1.2"

//	@title			GopherSocial API
//	@description	API for GopherSocial.
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath					/v2
//
// @securityDefinitions.apikey	ApiKeyAuth
// @in							header
// @name						Authorization
// @description
func main() {
	cfg := config{
		addr:   env.GetString("ADDR", ":8080"),
		apiURL: env.GetString("EXTERNAL_URL", "localhost:8080"),
		db: dbConfig{
			addr:         env.GetString("DB_ADDR", "postgres://myuser:mypassword@localhost/mydatabase?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
		redis: redisConfig{
			addr:   env.GetString("REDIS_ADDR", "localhost:6379"),
			pass:   env.GetString("REDIS_PASS", ""),
			db:     env.GetInt("REDIS_DB", 0),
			enable: env.GetBool("REDIS_ENABLE", false),
		},
		env: env.GetString("ENV", "development"),
		mail: mailConfig{
			exp:       time.Hour * 24 * 3,
			fromEmail: env.GetString("FROM_EMAIL", "hellofallback@demomailtrap.com"),
			sendGrid: sendGridConfig{
				apiKey: env.GetString("SENDGRID_API_KEY", ""),
			},
			mailTrap: mailTrapConfig{
				apiKey: env.GetString("MAILTRAP_API_KEY", ""),
			},
		},
		auth: authConfig{
			basic: basicConfig{
				user: env.GetString("AUTH_BASIC_USER", "admin"),
				pass: env.GetString("AUTH_BASIC_PASS", "admin"),
			},
			token: tokenConfig{
				secret: env.GetString("AUTH_TOKEN_SECRET", "example"),
				exp:    time.Hour * 24 * 3, // 3 days
				iss:    "gohersocial",
			},
		},
		rateLimiter: ratelimiter.Config{
			RequestPerTimeFrame: env.GetInt("RATELIMITER_REQUESTS_COUNT", 20),
			TimeFram:            time.Second * 5,
			Enable:              env.GetBool("RATELIMITER_ENABLE", true),
		},
	}

	// Logger
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "ts"
	config.EncoderConfig.LevelKey = "level"
	config.EncoderConfig.MessageKey = "msg"
	config.DisableStacktrace = true // ปิดการแสดง stacktrace

	logger := zap.Must(config.Build()).Sugar()
	defer logger.Sync()

	// ratelimiter
	ratelimiter := ratelimiter.NewFixedWindowRateLimiter(cfg.rateLimiter.RequestPerTimeFrame, cfg.rateLimiter.TimeFram)

	// Sendgrid
	// mailer := mailer.NewSendGrid(cfg.mail.fromEmail, cfg.mail.sendGrid.apiKey)
	mailtrap, err := mailer.NewMailTrapClient(cfg.mail.fromEmail, cfg.mail.mailTrap.apiKey)
	if err != nil {
		logger.Fatal(err)
	}

	// authenticator
	jwtAuthenticator := auth.NewJWTAuthenticator(
		cfg.auth.token.secret,
		cfg.auth.token.iss,
		cfg.auth.token.iss,
	)

	// DB
	db, err := db.New(
		cfg.db.addr,
		cfg.db.maxOpenConns,
		cfg.db.maxIdleConns,
		cfg.db.maxIdleTime,
	)
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()
	logger.Info("database connection pool established")

	// Redis
	var rdb *redis.Client
	if cfg.redis.enable {
		rdb = cache.NewRedisClient(cfg.redis.addr, cfg.redis.pass, cfg.redis.db)
		logger.Info("redis cache connection is established")

		defer rdb.Close()
	}

	store := store.NewStorage(db)
	cache := cache.NewCacheStorage(rdb)
	app := &application{
		config:        cfg,
		store:         store,
		cacheStore:    cache,
		logger:        logger,
		mailer:        mailtrap,
		authenticator: jwtAuthenticator,
		ratelimiter:   ratelimiter,
	}

	// Metrics collected
	expvar.NewString("version").Set(version)
	expvar.Publish("database", expvar.Func(func() any {
		return db.Stats()
	}))
	expvar.Publish("goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))

	mux := app.mount()
	logger.Fatal(app.run(mux))
}
