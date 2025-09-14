package main

// first run (go mod init restapi)

import (
	"crypto/tls"
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	mw "restapi/internal/api/middlewares"
	"restapi/internal/api/router"
	"restapi/pkg/utils"
	"time"

	"github.com/joho/godotenv"
)

//go:embed .env
var envFile embed.FS

func loadEnvFromEmbeddedFile() {
	// read the embedded.env file
	content, err := envFile.ReadFile(".env")
	if err != nil {
		log.Fatalf("Error reading .env file: %v", err)
	} 

	// create a temp file to load the env vars
	tempFile, err := os.CreateTemp("", ".env")
	if err != nil {
		log.Fatalf("Error creating temp .env file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Write content of embedded .env file to the time file
	_, err = tempFile.Write(content)
	if err != nil {
		log.Fatalf("Error writing .env file: %v", err)
	} 

	err = tempFile.Close()
	if err != nil {
		log.Fatalf("Error closing .env file: %v", err)
	} 

	// Load the env vars from the temp file
	err = godotenv.Load(tempFile.Name())
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}


func main() {
	// Only in production, for running source code
	// err := godotenv.Load()
	// if err != nil {
	// 	return
	// }
	
	// Load env vars from the embedded .env file
	loadEnvFromEmbeddedFile()

	fmt.Println("Env Var CERT_FILE:", os.Getenv("CERT_FILE"))

	port := os.Getenv("API_PORT")

	// cert := "cert.pem"
	// key := "key.pem"
	cert := os.Getenv("CERT_FILE")
	key := os.Getenv("KEY_FILE")

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	rl := mw.NewRateLimiter(5, time.Minute)

	hppOptions := mw.HPPOptions{
		CheckQuery: true,
		CheckBody: true,
		CheckBodyOnlyForContentType: "application/x-www-form-encoded",
		WhiteList: []string{"sortBy", "sortOrder", "name", "age", "class"},
	}

	// secureMux := mw.Cors(rl.Middleware(mw.ResponsetimeMiddleware(mw.SecurityHeaders(mw.Compression(mw.Hpp(hppOptions)(mux))))))
	router := router.MainRouter()
	jwtMiddleware := mw.MiddlewaresExcludePaths(mw.JWTMiddleware, "/execs/login", "/execs/forgotpassword", "/execs/resetpassword/reset")
	secureMux := utils.ApplyMiddlewares(router, mw.SecurityHeaders, mw.Compression, mw.Hpp(hppOptions), mw.XSSMiddleware, jwtMiddleware, mw.ResponsetimeMiddleware, rl.Middleware, mw.Cors)

	// secureMux := mw.XSSMiddleware(router)

	// create custom server
	server := &http.Server{
		Addr: port,
		Handler: secureMux,
		TLSConfig: tlsConfig,
	}

	fmt.Println("Server is running on port: ", port)
	err := server.ListenAndServeTLS(cert, key)
	if err != nil {
		log.Fatalln("Error starting the server", err)
	}
}