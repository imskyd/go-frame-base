package base

import (
	"github.com/deng00/go-base/cache/redis"
	"github.com/deng00/go-base/config"
	"github.com/deng00/go-base/db/mysql"
	"github.com/deng00/go-base/logging"
)

var configManager = config.Manager{}
var logger *logging.SugaredLogger

const ServiceName = "products-frame-base"

func init() {
	// 配置解析
	err := configManager.Init(ServiceName)
	if err != nil {
		panic("init config failed, " + err.Error())
	}
	logger = GetLogger("base").Sugar()
}

func GetConfig() *config.Config {
	c := configManager.GetIns()
	return c
}

func GetLogger(module string) *logging.Logger {
	logConfig := logging.GetLogConfig(GetConfig())
	return logging.GetLogger(ServiceName, module, logConfig)
}

func GetMysqlConfig() *mysql.Config {
	config := new(mysql.Config)
	if err := GetConfig().UnmarshalKey("mysql", config); err != nil {
		logger.Fatalf("invalid mysql config:%s", err)
	}
	return config
}

func GetJwtConfig() *JwtConfig {
	config := new(JwtConfig)
	if err := GetConfig().UnmarshalKey("jwt", config); err != nil {
		logger.Fatalf("invalid consul config:%s", err)
	}
	return config
}

func GetTelegramConfig() *TelegramConfig {
	config := new(TelegramConfig)
	if err := GetConfig().UnmarshalKey("telegram", config); err != nil {
		logger.Fatalf("invalid telegram config:%s", err)
	}
	return config
}

func GetRedisConfig() *redis.Config {
	config := new(redis.Config)
	if err := GetConfig().UnmarshalKey("redis", config); err != nil {
		logger.Fatalf("invalid redis config:%s", err)
	}
	return config
}
