package platform

import "go.uber.org/zap"

func NewLogger(env string) *zap.Logger {
	if env == "prod" {
		return zap.Must(zap.NewProduction())
	}
	return zap.Must(zap.NewDevelopment())
}
