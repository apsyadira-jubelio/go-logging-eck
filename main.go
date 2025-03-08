package main

import (
	"net/http"

	"github.com/joho/godotenv"
	"github.com/jubelio/go-logging/getenv"
	"github.com/jubelio/go-logging/logging"
	"github.com/sirupsen/logrus"
)

var logger *logging.StandardLogger

func init() {
	logging.InitLogger(logging.LoggerConfig{
		AppName:     "test-app",
		Host:        getenv.GetEnvString("ELASTICSEARCH_HOST", "http://localhost:9200"),
		Username:    getenv.GetEnvString("ELASTICSEARCH_USERNAME", ""),
		Password:    getenv.GetEnvString("ELASTICSEARCH_PASSWORD", ""),
		Environment: "development",
		EnableLog:   true,
		LogLevel:    "info",
	})

	logger = logging.L()
}

func main() {
	err := godotenv.Load()
	if err != nil {
		logger.Fatalf("Error loading .env file")
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logger.WithFields(logrus.Fields{
			"method": r.Method,
			"path":   r.URL.Path,
		}).Info("Request received")
		w.Write([]byte("Hello, world!"))
	})

	logger.Info("Server started at :8080")
	http.ListenAndServe(":8080", nil)
}
