package main

import (
	"net"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"go-api/api/v1"
	"go-api/database"
	"go-api/middleware"
	"go-api/utils"
)

func main() {
	if err := godotenv.Load(); err != nil {
		utils.LogInfo("Warning: .env file not found")
	}

	db, err := database.ConnectPostgres()

	if err != nil {
		utils.LogFatal("Database connection failed: %v", err)
	}

	if err := database.SeedUser(db); err != nil {
		utils.LogFatal("User seeding failed: %v", err)
	}

	sqlDB, err := db.DB()
	if err == nil {
		defer sqlDB.Close()
	}

	utils.LogSuccess("Database connected successfully")

	mux := http.NewServeMux()

	v1.RegisterRoutes(mux, db)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	localIP := getLocalIP()
	utils.LogInfo("Server running on:")
	utils.LogInfo("- Local:   http://localhost:%s", port)
	utils.LogInfo("- Network: http://%s:%s", localIP, port)

	if err := http.ListenAndServe("0.0.0.0:"+port, middleware.CORS(mux)); err != nil {
		utils.LogFatal("Server failed: %v", err)
	}
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "localhost"
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "localhost"
}
