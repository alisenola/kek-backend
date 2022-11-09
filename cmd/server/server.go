package main

import (
	"context"
	"fmt"
	"kek-backend/internal/account"
	accountDB "kek-backend/internal/account/database"
	"kek-backend/internal/alert"
	alertDB "kek-backend/internal/alert/database"
	"kek-backend/internal/article"
	articleDB "kek-backend/internal/article/database"
	"kek-backend/internal/config"
	"kek-backend/internal/database"
	"kek-backend/internal/metric"
	"kek-backend/pkg/logging"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	cors "github.com/itsjamie/gin-cors"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var serverCmd = &cobra.Command{
	Use: "server",
	Run: func(cmd *cobra.Command, args []string) {
		runApplication()
	},
}

func newServer(lc fx.Lifecycle, cfg *config.Config, mp *metric.MetricsProvider) *gin.Engine {
	gin.SetMode(gin.DebugMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	metric.Route(r)
	r.Use(metric.MetricsMiddleware(mp))
	r.Use(cors.Middleware(cors.Config{
		Origins:         "*",
		Methods:         "GET, PUT, POST, DELETE",
		RequestHeaders:  "Origin, Authorization, Content-Type",
		ExposedHeaders:  "",
		MaxAge:          50 * time.Second,
		Credentials:     true,
		ValidateHeaders: false,
	}))

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.ServerConfig.Port),
		Handler: r,
	}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logging.FromContext(ctx).Infof("Start to rest api server :%d", cfg.ServerConfig.Port)
			go srv.ListenAndServe()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logging.FromContext(ctx).Infof("Stopped rest api server")
			return srv.Shutdown(ctx)
		},
	})
	return r
}

func printAppInfo(cfg *config.Config) {
	logging.DefaultLogger().Infow("application information", "config", cfg)
}

func loadConfig() (*config.Config, error) {
	return config.Load(configFile)
}

func runApplication() {
	// setup application(di + run server)
	app := fx.New(
		fx.Provide(
			// load config
			loadConfig,
			metric.NewMetricsProvider,
			// setup database
			database.NewDatabase,
			// setup account packages
			accountDB.NewAccountDB,
			account.NewAuthMiddleware,
			account.NewHandler,
			// setup article packages
			articleDB.NewArticleDB,
			article.NewHandler,
			// setup alert packages
			alertDB.NewAlertDB,
			alert.NewHandler,
			// server
			newServer,
		),
		fx.Invoke(
			account.RouteV1,
			article.RouteV1,
			alert.RouteV1,
			printAppInfo,
		),
	)
	app.Run()
}
