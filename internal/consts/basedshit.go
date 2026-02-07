package consts

import "time"

const (
	AppName      = "grpc-gui"
	AppConfigDir = ".grpc-gui"
	AppDbName    = "grpc-gui.db"

	MaxHistorySize = 500

	ReflectionCacheTTL           = 10 * time.Minute
	ReflectionCacheRefreshEvery  = 20
)
