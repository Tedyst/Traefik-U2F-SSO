package main

import (
	"go.uber.org/zap"
	"net/http"
	"time"
)

var logger *zap.SugaredLogger

// RequestLogger logs every request
func RequestLogger(targetMux http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		targetMux.ServeHTTP(w, r)

		//log request by who(IP address)
		requesterIP := r.RemoteAddr

		logger.Infow("Loaded page",
			"Method", r.Method,
			"RequestURI", r.RequestURI,
			"RequesterIP", requesterIP,
			"Time", time.Since(start),
		)
	})
}

func initLogger(config Configuration) error {
	zaplog, err := zap.NewProduction()
	if err != nil {
		return err
	}
	if Config.Debug == true {
		zaplog, _ = zap.NewDevelopment()
	}
	defer zaplog.Sync()
	logger = zaplog.Sugar()
	return nil
}
