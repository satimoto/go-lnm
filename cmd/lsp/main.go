package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/util"
	"github.com/satimoto/go-lsp/internal/metric"
	"github.com/satimoto/go-lsp/internal/monitor"
	"github.com/satimoto/go-lsp/internal/rest"
	"github.com/satimoto/go-lsp/internal/rpc"
	"github.com/satimoto/go-lsp/internal/service"
	"github.com/spf13/cobra"
)

var runCommand = &cobra.Command{
	Use:   "lsp",
	Short: "Run the Lightning Service Provider",
	Long:  "Run the Lightning Service Provider",
	Run:   startLsp,
}

func main() {
	configFile, err := os.UserHomeDir()

	if err == nil {
		configFile = configFile + "/.lsp/"
	}

	configFile = configFile + "lsp.conf"

	runCommand.Flags().StringP("configfile", "C", configFile, "Config")
	runCommand.Execute()
}

func startLsp(cmd *cobra.Command, args []string) {
	configFile, _ := cmd.Flags().GetString("configfile")

	godotenv.Load(configFile)

	dbHost := os.Getenv("DB_HOST")
	dbName := os.Getenv("DB_NAME")
	dbPass := os.Getenv("DB_PASS")
	dbUser := os.Getenv("DB_USER")
	sslMode := util.GetEnv("SSL_MODE", "disable")

	if len(dbHost) == 0 || len(dbName) == 0 || len(dbPass) == 0 || len(dbUser) == 0 {
		log.Fatalf("Database env variables not defined")
	}

	dataSourceName := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s", dbUser, dbPass, dbHost, dbName, sslMode)
	database, err := sql.Open("postgres", dataSourceName)

	if err != nil {
		log.Fatal(err)
	}

	defer database.Close()

	log.Printf("Starting up LSP server")
	repositoryService := db.NewRepositoryService(database)

	shutdownCtx, cancelFunc := context.WithCancel(context.Background())
	waitGroup := &sync.WaitGroup{}	

	services := service.NewService()
	services.FerpService.Start(shutdownCtx, waitGroup)

	metricsService := metrics.NewMetrics()
	metricsService.StartMetrics(shutdownCtx, waitGroup)

	restService := rest.NewRest(database)
	restService.StartRest(shutdownCtx, waitGroup)

	rpcService := rpc.NewRpc(shutdownCtx, database, services)
	rpcService.StartRpc(waitGroup)

	monitor := monitor.NewMonitor(shutdownCtx, repositoryService, services)
	monitor.StartMonitor(waitGroup)

	sigtermChan := make(chan os.Signal, 1)
	signal.Notify(sigtermChan, os.Kill, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-sigtermChan

	log.Printf("Shutting down LSP server")

	cancelFunc()
	waitGroup.Wait()

	log.Printf("LSP server shut down")
}
