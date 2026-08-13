package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gookit/slog"

	"github.com/walfa-labs/backend/internal/adapter/handler"
	"github.com/walfa-labs/backend/internal/adapter/middleware"
	"github.com/walfa-labs/backend/internal/adapter/repository/oracle/adw"
	"github.com/walfa-labs/backend/internal/adapter/repository/oracle/atp"
	"github.com/walfa-labs/backend/internal/adapter/repository/oracle/objectstorage"
	"github.com/walfa-labs/backend/internal/config"
	"github.com/walfa-labs/backend/internal/platform"
	"github.com/walfa-labs/backend/internal/router"
	"github.com/walfa-labs/backend/internal/service"
)

func main() {
	// --- Config ---
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// --- Logger ---
	logger := platform.NewLogger(cfg.AppEnv)
	defer func() {
		if err := slog.Close(); err != nil {
			log.Printf("failed to close logger: %v", err)
		}
	}()

	logger.Infof("starting %s in %s mode", "Portfolio API", cfg.AppEnv)

	// --- Databases (ATP = operational OLTP, ADW = analytics) ---
	ctx := context.Background()
	atpDB, err := platform.NewOracleDB(ctx, cfg.ATPDSN)
	if err != nil {
		logger.Fatalf("failed to connect to ATP database: %v", err)
	}
	defer atpDB.Close()

	adwDB, err := platform.NewOracleDB(ctx, cfg.ADWDSN)
	if err != nil {
		logger.Fatalf("failed to connect to ADW database: %v", err)
	}
	defer adwDB.Close()

	// --- Repositories (driven adapters) ---
	expRepo := atp.NewExperienceRepo(atpDB)
	projectRepo := atp.NewProjectRepo(atpDB)
	postRepo := atp.NewPostRepo(atpDB)
	tagRepo := atp.NewTagRepo(atpDB)
	assetRepo := atp.NewAssetRepo(atpDB)
	adminRepo := atp.NewAdminRepo(atpDB)
	statsRepo := atp.NewStatsRepo(atpDB)
	profileRepo := atp.NewProfileRepo(atpDB)

	analyticsStore := adw.NewAnalyticsStore(adwDB)

	assetStore, err := objectstorage.NewAssetStore(
		cfg.OCI.TenancyOCID,
		cfg.OCI.UserOCID,
		cfg.OCI.Fingerprint,
		cfg.OCI.Region,
		cfg.OCI.PrivateKeyPath,
		cfg.OCI.Namespace,
		cfg.OCI.Bucket,
	)
	if err != nil {
		logger.Fatalf("failed to create asset store: %v", err)
	}

	// --- Services (use-cases) ---
	expSvc := service.NewExperienceService(expRepo)
	projectSvc := service.NewProjectService(projectRepo)
	postSvc := service.NewPostService(postRepo, analyticsStore)
	authSvc := service.NewAuthService(adminRepo, cfg.JWTSecretKey, cfg.JWTAccessTTL, cfg.JWTRefreshTTL)
	assetSvc := service.NewAssetService(assetRepo, assetStore)
	statsSvc := service.NewStatsService(statsRepo, analyticsStore)
	profileSvc := service.NewProfileService(profileRepo)

	// --- Handlers (driving adapters) ---
	healthH := handler.NewHealthHandler(atpDB, adwDB)
	expH := handler.NewExperienceHandler(expSvc)
	projectH := handler.NewProjectHandler(projectSvc)
	postH := handler.NewPostHandler(postSvc)
	authH := handler.NewAuthHandler(authSvc, int(cfg.JWTRefreshTTL.Hours()))
	assetH := handler.NewAssetHandler(assetSvc)
	statsH := handler.NewStatsHandler(statsSvc, tagRepo)
	profileH := handler.NewProfileHandler(profileSvc)

	// --- Auth middleware ---
	authMiddleware := middleware.Auth(cfg, logger)

	// --- Server + Router ---
	app := platform.NewServer(cfg, logger)
	router.Register(app, router.Deps{
		Cfg:            cfg,
		Health:         healthH,
		Experience:     expH,
		Project:        projectH,
		Post:           postH,
		Auth:           authH,
		Asset:          assetH,
		Stats:          statsH,
		Profile:        profileH,
		Logger:         logger,
		AuthMiddleware: authMiddleware,
	})

	// --- Graceful shutdown ---
	go func() {
		logger.Infof("server listening on %s", cfg.AppPort)
		if err := app.Listen(cfg.AppPort); err != nil {
			logger.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		logger.Errorf("forced shutdown: %v", err)
	}
	logger.Info("server stopped")
}
