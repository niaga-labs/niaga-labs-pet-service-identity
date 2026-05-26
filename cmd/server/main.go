package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Kilat-Pet-Delivery/lib-common/auth"
	"github.com/Kilat-Pet-Delivery/lib-common/database"
	"github.com/Kilat-Pet-Delivery/lib-common/health"
	"github.com/Kilat-Pet-Delivery/lib-common/kafka"
	"github.com/Kilat-Pet-Delivery/lib-common/logger"
	"github.com/Kilat-Pet-Delivery/lib-common/middleware"
	protoEvents "github.com/Kilat-Pet-Delivery/lib-proto/events"
	"github.com/Kilat-Pet-Delivery/service-identity/internal/application"
	svcconfig "github.com/Kilat-Pet-Delivery/service-identity/internal/config"
	identityEvents "github.com/Kilat-Pet-Delivery/service-identity/internal/events"
	"github.com/Kilat-Pet-Delivery/service-identity/internal/handler"
	"github.com/Kilat-Pet-Delivery/service-identity/internal/repository"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// 1. Load configuration
	cfg, err := svcconfig.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// 2. Initialize zap logger
	zapLogger, err := logger.NewNamed(cfg.AppEnv, "service-identity")
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer func() { _ = zapLogger.Sync() }()

	// 3. Connect to PostgreSQL
	dbConfig := database.PostgresConfig{
		Host:     cfg.DBConfig.Host,
		Port:     cfg.DBConfig.Port,
		User:     cfg.DBConfig.User,
		Password: cfg.DBConfig.Password,
		DBName:   cfg.DBConfig.DBName,
		SSLMode:  cfg.DBConfig.SSLMode,
	}

	db, err := database.Connect(dbConfig, zapLogger)
	if err != nil {
		zapLogger.Fatal("failed to connect to database", zap.Error(err))
	}

	// 4. Run database migrations
	if cfg.AppEnv == "development" {
		// RunnerApplicationModel is intentionally omitted: GORM's migrator drops its
		// conventional unique-constraint name (uni_runner_applications_ic_number)
		// which doesn't match the SQL migration's name (runner_applications_ic_number_key).
		// SQL migrations own this table.
		if err := db.AutoMigrate(&repository.UserModel{}, &repository.UserRoleModel{}, &repository.RefreshTokenModel{}, &repository.PasswordResetModel{}, &repository.ReferralModel{}, &repository.UserReferralCodeModel{}); err != nil {
			zapLogger.Fatal("failed to auto-migrate", zap.Error(err))
		}
		zapLogger.Info("database migration completed (dev auto-migrate)")
	} else {
		dbURL := dbConfig.DatabaseURL()
		if err := database.RunMigrations(dbURL, "migrations", zapLogger); err != nil {
			zapLogger.Fatal("failed to run migrations", zap.Error(err))
		}
	}

	// 5. Initialize JWT manager with durations parsed from config
	accessExpiry, err := time.ParseDuration(cfg.JWTConfig.AccessExpiry)
	if err != nil {
		accessExpiry = 15 * time.Minute
		zapLogger.Warn("invalid JWT_ACCESS_EXPIRY, using default 15m", zap.Error(err))
	}

	refreshExpiry, err := time.ParseDuration(cfg.JWTConfig.RefreshExpiry)
	if err != nil {
		refreshExpiry = 7 * 24 * time.Hour
		zapLogger.Warn("invalid JWT_REFRESH_EXPIRY, using default 7d", zap.Error(err))
	}

	jwtSecret := cfg.JWTConfig.Secret
	if jwtSecret == "" {
		jwtSecret = "default-secret-change-me"
		zapLogger.Warn("JWT_SECRET not set, using insecure default")
	}

	jwtManager := auth.NewJWTManager(jwtSecret, accessExpiry, refreshExpiry)

	// 6. Create repositories
	userRepo := repository.NewGormUserRepository(db)
	tokenRepo := repository.NewGormTokenRepository(db)
	passwordResetRepo := repository.NewGormPasswordResetRepository(db)

	// 7. Create auth service
	notifier := application.NewLogOnlyPasswordResetNotifier(zapLogger)
	authService := application.NewAuthService(userRepo, tokenRepo, passwordResetRepo, notifier, jwtManager, zapLogger)
	consumerCtx, cancelConsumers := context.WithCancel(context.Background())
	defer cancelConsumers()
	shopKafkaConsumer := kafka.NewConsumer(cfg.KafkaConfig.Brokers, cfg.KafkaConfig.GroupPrefix+"identity-shop-roles", protoEvents.TopicShopEvents, zapLogger)
	defer func() { _ = shopKafkaConsumer.Close() }()
	shopConsumer := identityEvents.NewShopEventConsumer(authService, zapLogger)
	go func() {
		if err := shopKafkaConsumer.Consume(consumerCtx, shopConsumer.HandleKafkaMessage); err != nil && consumerCtx.Err() == nil {
			zapLogger.Error("shop event consumer stopped", zap.Error(err))
		}
	}()

	// 8. Create Gin router with global middleware
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(
		middleware.RequestIDMiddleware(),
		middleware.CORSMiddleware(),
		middleware.SecurityHeadersMiddleware(),
		middleware.LoggerMiddleware(zapLogger),
		middleware.RecoveryMiddleware(zapLogger),
		middleware.RateLimitMiddleware(100, time.Minute),
	)

	// 9. Register health routes
	healthHandler := health.NewHandler(db, "service-identity")
	healthHandler.RegisterRoutes(router)

	// Initialize referral service and handler
	referralRepo := repository.NewGormReferralRepository(db)
	referralService := application.NewReferralService(referralRepo, zapLogger)
	referralHandler := handler.NewReferralHandler(referralService)

	// 10. Register auth handler routes
	apiV1 := router.Group("/api/v1")
	authHandler := handler.NewAuthHandler(authService, zapLogger)
	authHandler.RegisterRoutes(apiV1, jwtManager)
	referralHandler.RegisterRoutes(&router.RouterGroup, jwtManager)

	forgotPasswordHandler := handler.NewForgotPasswordHandler(authService, zapLogger)
	forgotPasswordHandler.RegisterRoutes(apiV1)

	resetPasswordHandler := handler.NewResetPasswordHandler(authService, zapLogger)
	resetPasswordHandler.RegisterRoutes(apiV1)

	runnerApplicationRepo := repository.NewGormRunnerApplicationRepository(db)
	runnerApplicationService := application.NewRunnerApplicationService(runnerApplicationRepo, zapLogger)
	runnerApplyHandler := handler.NewRunnerApplyHandler(runnerApplicationService, zapLogger)
	runnerApplyHandler.RegisterRoutes(apiV1)

	// Register admin handler routes
	adminHandler := handler.NewAdminHandler(authService)
	adminHandler.RegisterRoutes(&router.RouterGroup, jwtManager)

	// 11. Start HTTP server
	srv := &http.Server{
		Addr:         cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		zapLogger.Info("starting service-identity", zap.String("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zapLogger.Fatal("server failed", zap.Error(err))
		}
	}()

	// 12. Graceful shutdown on SIGINT/SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	zapLogger.Info("shutting down server...")
	cancelConsumers()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		zapLogger.Fatal("server forced to shutdown", zap.Error(err))
	}

	zapLogger.Info("server stopped gracefully")
}
