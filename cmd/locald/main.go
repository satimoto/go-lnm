package main

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"os"

	_ "github.com/joho/godotenv/autoload"
	"github.com/satimoto/go-datastore/db"
	dbUtil "github.com/satimoto/go-datastore/util"
	"github.com/satimoto/go-lsp/intercept"
	lspUtil "github.com/satimoto/go-lsp/util"
	"google.golang.org/grpc"
)

var (
	database *sql.DB

	dbHost  = os.Getenv("DB_HOST")
	dbName  = os.Getenv("DB_NAME")
	dbPass  = os.Getenv("DB_PASS")
	dbUser  = os.Getenv("DB_USER")
	sslMode = dbUtil.GetEnv("SSL_MODE", "disable")
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
	repositoryService := db.NewRepositoryService(database)

	lndTlsCert, err := base64.StdEncoding.DecodeString(os.Getenv("LND_TLS_CERT"))
	lspUtil.PanicOnError("Invalid LND TLS Certificate", err)

	credentials, err := lspUtil.NewCredential(string(lndTlsCert))
	lspUtil.PanicOnError("Error creating transport credentials", err)

	clientConn, err := grpc.Dial(os.Getenv("LND_GRPC_HOST"), grpc.WithTransportCredentials(credentials))
	lspUtil.PanicOnError("Error connecting to LND host", err)
	defer clientConn.Close()

	interceptor := intercept.NewInterceptor(repositoryService, clientConn)
	if err = interceptor.Register(); err == nil {
		go interceptor.InterceptHtlcs()
		go interceptor.SubscribeHtlcEvents()
		go interceptor.SubscribeChannelEvents()
		interceptor.SubscribeTransactions()
	}
}
