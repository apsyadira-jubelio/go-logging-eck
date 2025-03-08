package logging

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/sirupsen/logrus"
	"go.elastic.co/ecslogrus"
	"gopkg.in/go-extras/elogrus.v8"
)

type TenantInfo struct {
	CompanyID      int    `json:"company_id" bson:"company_id"`
	TenantID       string `json:"tenant_id" bson:"tenant_id"`
	ShopSchema     string `json:"store_schema" bson:"store_schema"`
	ShopName       string `json:"shop_name" bson:"shop_name"`
	Subdomain      string `json:"subdomain" bson:"subdomain"`
	Domain         string `json:"domain" bson:"domain"`
	StoreName      string `json:"store_name" bson:"store_name"`
	DBHost         string `json:"db_host" bson:"db_host"`
	StoreLogo      string `json:"store_logo" bson:"store_logo"`
	IsDomainActive bool   `json:"is_domain_active" bson:"is_domain_active"`
	IsEmailActive  bool   `json:"is_email_active" bson:"is_email_active"`
	ChargeFeeTo    string `json:"charge_fee_to" bson:"charge_fee_to"`
}

// StandardLogger enforces specific log message formats
type StandardLogger struct {
	*logrus.Entry
}

func ecsLogMessageModifierFunc(formatter *ecslogrus.Formatter) func(*logrus.Entry, *elogrus.Message) any {
	return func(entry *logrus.Entry, message *elogrus.Message) any {
		var data json.RawMessage
		data, err := formatter.Format(entry)
		if err != nil {
			return entry // in case of an error just preserve the original entry
		}
		return data
	}

}

type LoggerConfig struct {
	AppName     string
	Host        string
	Username    string
	Password    string
	Environment string
	EnableLog   bool
	LogLevel    string
}

func NewLogger(config LoggerConfig) *StandardLogger {
	return NewLoggerWithAppName(config)
}

// NewLogger initializes the standard logger
func NewLoggerWithAppName(config LoggerConfig) *StandardLogger {
	env := config.Environment
	// version := config.Config.Version

	if config.AppName == "" {
		config.AppName = "-"
	}

	// Create a new logger instance
	logger := logrus.New().WithFields(logrus.Fields{
		"app":         config.AppName,
		"environment": env,
	})

	// if env == "production" {
	// 	logger.Logger.Formatter = &logrus.JSONFormatter{PrettyPrint: true}
	// }

	lvl, err := logrus.ParseLevel(config.LogLevel)
	if err != nil {
		fmt.Println("Invalid log, will set to INFO")
		lvl = logrus.InfoLevel
	}

	// if lvl == logrus.TraceLevel {
	// 	logger.Logger.SetReportCaller(true)
	// }

	logger.Logger.SetLevel(lvl)
	logger.Logger.SetFormatter(&ecslogrus.Formatter{})
	logger.Logger.AddHook(&RenameLogLevelHook{})

	///////////////////////////////////////////////////////////////////////////////////
	//// ELASTICSEARCH HOOK ///////////////////////////////////////////////////////////
	logConf := config
	if logConf.EnableLog {
		logger.Trace("elastic-hook active with service-name --> ", config.AppName)
		logger = logger.WithFields(logrus.Fields{
			"fields": logrus.Fields{
				"service": config.AppName,
			},
		})

		hostName, err := os.Hostname()
		if err != nil {
			hostName = err.Error()
		}

		client, err := elasticsearch.NewClient(elasticsearch.Config{
			Addresses: []string{logConf.Host},
			Username:  logConf.Username,
			Password:  logConf.Password,
		})
		if err != nil {
			fmt.Println("Initialization elasticsearch client error ", err.Error())
			goto ignorehook
		}

		hook, err := elogrus.NewAsyncElasticHook(client, hostName, logrus.InfoLevel, "logs-app-default")
		if err != nil {
			goto ignorehook
		}
		hook.MessageModifierFunc = ecsLogMessageModifierFunc(&ecslogrus.Formatter{})
		logger.Logger.AddHook(hook)
	}
	/////////////////////////////////////////////////////////////////////////////////
	/////////////////////////////////////////////////////////////////////////////////

ignorehook:

	var standardLogger = &StandardLogger{logger}

	return standardLogger
}

func (l *StandardLogger) WithTenantInfo(traceName string, tc TenantInfo) *logrus.Entry {
	return l.WithFields(logrus.Fields{
		"trace_name":       traceName,
		"tenant_id":        tc.TenantID,
		"shop_schema":      tc.ShopSchema,
		"shop_name":        tc.ShopName,
		"domain":           tc.Domain,
		"subdomain":        tc.Subdomain,
		"db_host":          tc.DBHost,
		"store_logo":       tc.StoreLogo,
		"is_domain_active": tc.IsDomainActive,
	})
}

// RenameLogLevelHook is a custom Logrus hook to rename log.level to severity
type RenameLogLevelHook struct{}

// Levels defines on which log levels the hook should trigger
func (h *RenameLogLevelHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire is triggered on each log entry and renames log.level to severity
func (h *RenameLogLevelHook) Fire(entry *logrus.Entry) error {
	if level, exists := entry.Data["log.level"]; exists {
		entry.Data["severity"] = level  // Rename log.level to severity
		delete(entry.Data, "log.level") // Remove log.level
	} else {
		// If log.level doesn't exist, copy the entry's internal level
		entry.Data["severity"] = entry.Level.String()
	}
	return nil
}

// // CUSTOM FORMATTER
// CustomFormatter is a custom Logrus formatter to rename log.level to severity
type CustomFormatter struct {
	logrus.TextFormatter
}

// Format modifies the log entry and renames log.level to severity before formatting
func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// Check if log.level exists and rename it to severity
	if level, exists := entry.Data["log.level"]; exists {
		entry.Data["severity"] = level  // Rename log.level to severity
		delete(entry.Data, "log.level") // Remove log.level
	} else {
		// If log.level doesn't exist, copy the internal entry.Level to severity
		entry.Data["severity"] = entry.Level.String()
	}

	// Call the base formatter (TextFormatter in this case)
	return f.TextFormatter.Format(entry)
}
