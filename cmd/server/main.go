// cmd/server/main.go
package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bete7512/bill-aggregator/internal/adapter/inbound/rest/handler"
	"github.com/bete7512/bill-aggregator/internal/adapter/inbound/rest/middleware"
	"github.com/bete7512/bill-aggregator/internal/adapter/outbound/provider/client"
	"github.com/bete7512/bill-aggregator/internal/adapter/outbound/repository/postgres"
	"github.com/bete7512/bill-aggregator/internal/config"
	"github.com/bete7512/bill-aggregator/internal/core/domain/service"
	"github.com/bete7512/bill-aggregator/pkg/util/logger"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database connection
	db, err := initDatabase(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize repositories
	accountRepo := postgres.NewAccountRepository(db)
	billRepo := postgres.NewBillRepository(db)

	// Initialize provider client
	providerClient := client.NewProviderClient(cfg.ProviderAPI.BaseURL)

	// Initialize services
	accountService := service.NewAccountService(accountRepo, providerClient)
	billService := service.NewBillService(billRepo, accountRepo, providerClient)
	logger := logger.NewLogger(logger.Config{
		Format: cfg.Logger.Format,
		Level:  cfg.Logger.Level,
	})
		
	// Initialize HTTP handlers
	accountHandler := handler.NewAccountHandler(accountService, providerClient, logger)
	billHandler := handler.NewBillHandler(billService, providerClient, logger)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(cfg.JWT.Secret, cfg.JWT.Expiration, logger)

	// Initialize router
	router := setupRouter(authMiddleware, accountHandler, billHandler)

	// Start HTTP server
	server := &http.Server{
		Addr:    cfg.Server.Address,
		Handler: router,
	}

	// Start server in a separate goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("Server started on %s", cfg.Server.Address)

	// Wait for termination signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Create a deadline for the shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown server gracefully
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}

// initDatabase initializes the database connection
func initDatabase(cfg config.DatabaseConfig) (*sql.DB, error) {
	connStr := cfg.ConnectionString


	log.Printf("Connecting to database: %s", connStr)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Set connection pool parameters
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetimeSecs) * time.Second)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

// setupRouter sets up the HTTP router
func setupRouter(
	authMiddleware *middleware.AuthMiddleware,
	accountHandler *handler.AccountHandler,
	billHandler *handler.BillHandler,
) *gin.Engine {
	router := gin.Default()

	// Middleware
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})

	// API routes
	api := router.Group("/api/v1")
	{
		// Accounts
		accounts := api.Group("/accounts")
		{
			accounts.POST("/link", authMiddleware.Authenticate(), accountHandler.LinkAccount)
			accounts.GET("", authMiddleware.Authenticate(), accountHandler.GetLinkedAccounts)
			accounts.DELETE("/:accountID", authMiddleware.Authenticate(), accountHandler.DeleteLinkedAccount)
		}

		// Bills
		bills := api.Group("/bills")
		{
			bills.GET("", authMiddleware.Authenticate(), billHandler.GetAggregatedBills)
			bills.GET("/:providerID", authMiddleware.Authenticate(), billHandler.GetBillsByProvider)
			bills.POST("/refresh", authMiddleware.Authenticate(), billHandler.RefreshBills)
		}
	}

	return router
}
