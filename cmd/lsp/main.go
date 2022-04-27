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
	"github.com/satimoto/go-datastore/db"
	"github.com/satimoto/go-datastore/util"
	"github.com/satimoto/go-lsp/internal/monitor"
	"github.com/satimoto/go-lsp/internal/rest"
	"github.com/satimoto/go-lsp/internal/rpc"
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

	dataSourceName := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s", dbUser, dbPass, dbHost, dbName, sslMode)
	d, err := sql.Open("postgres", dataSourceName)

	if err != nil {
		log.Fatal(err)
	}

	database = d
}

func main() {
	defer database.Close()
	
	log.Printf("Starting up LSP server")
	repositoryService := db.NewRepositoryService(database)

	shutdownCtx, cancelFunc := context.WithCancel(context.Background())
	waitGroup := &sync.WaitGroup{}

	restService := rest.NewRest(database)
	restService.StartRest(shutdownCtx, waitGroup)

	rpcService := rpc.NewRpc(shutdownCtx, database)
	rpcService.StartRpc(waitGroup)

	monitor := monitor.NewMonitor(shutdownCtx, repositoryService)
	monitor.StartMonitor(waitGroup)

	sigtermChan := make(chan os.Signal)
	signal.Notify(sigtermChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigtermChan

	log.Printf("Shutting down LSP server")

	cancelFunc()
	waitGroup.Wait()

	log.Printf("LSP server shut down")

}
