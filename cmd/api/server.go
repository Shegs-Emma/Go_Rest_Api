package main

// first run (go mod init restapi)

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	mw "restapi/internal/api/middlewares"
	"restapi/internal/api/router"
	"restapi/internal/repository/sqlconnect"
	"restapi/pkg/utils"
	"time"

	"github.com/joho/godotenv"
)


func main() {
	err := godotenv.Load()
	if err != nil {
		return
	}
	
	_, err = sqlconnect.ConnectDb()
	if err != nil {
		utils.ErrorHandler(err, "")
		return
	}
	port := os.Getenv("API_PORT")

	cert := "cert.pem"
	key := "key.pem"

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
	err = server.ListenAndServeTLS(cert, key)
	if err != nil {
		log.Fatalln("Error starting the server", err)
	}
}