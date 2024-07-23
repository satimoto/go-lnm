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

	_ "github.com/joho/godotenv/autoload"
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/util"
	metrics "github.com/satimoto/go-lnm/internal/metric"
	"github.com/satimoto/go-lnm/internal/monitor"
	"github.com/satimoto/go-lnm/internal/rest"
	"github.com/satimoto/go-lnm/internal/rpc"
	"github.com/satimoto/go-lnm/internal/service"
)

var (
	database *sql.DB

	dbHost  = os.Getenv("DB_HOST")
	dbName  = os.Getenv("DB_NAME")
	dbPass  = os.Getenv("DB_PASS")
	dbUser  = os.Getenv("DB_USER")
	sslMode = util.GetEnv("SSL_MODE", "disable")
)

func init() {
	if len(dbHost) == 0 || len(dbName) == 0 || len(dbPass) == 0 || len(dbUser) == 0 {
		log.Fatalf("Database env variables not defined")
	}

	dataSourceName := fmt.Sprintf("postgres://%s:%s@%s/%s?binary_parameters=yes&sslmode=%s", dbUser, dbPass, dbHost, dbName, sslMode)
	d, err := sql.Open("postgres", dataSourceName)

	if err != nil {
		log.Fatal(err)
	}

	database = d
}

func main() {
	defer database.Close()

	log.Printf("Starting up LNM server")
	repositoryService := db.NewRepositoryService(database)

	shutdownCtx, cancelFunc := context.WithCancel(context.Background())
	waitGroup := &sync.WaitGroup{}

	services := service.NewService()
	services.FerpService.Start(shutdownCtx, waitGroup)

	metricsService := metrics.NewMetrics()
	metricsService.StartMetrics(shutdownCtx, waitGroup)

	monitorService := monitor.NewMonitor(shutdownCtx, repositoryService, services)

	restService := rest.NewRest(database)
	restService.StartRest(shutdownCtx, waitGroup)

	rpcService := rpc.NewRpc(shutdownCtx, database, services, monitorService)
	rpcService.StartRpc(waitGroup)

	monitorService.StartMonitor(waitGroup)

	sigtermChan := make(chan os.Signal, 1)
	signal.Notify(sigtermChan, os.Kill, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-sigtermChan

	log.Printf("Shutting down LNM server")

	cancelFunc()
	waitGroup.Wait()

	log.Printf("LNM server shut down")
}
