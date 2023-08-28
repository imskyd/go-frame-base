package router

import (
	redisPkgBase "github.com/deng00/go-base/cache/redis"
	"github.com/deng00/go-base/db/mysql"
	"github.com/imskyd/go-frame-base/actions"
	auth02 "github.com/imskyd/go-frame-base/auth0"
	"github.com/imskyd/go-frame-base/base"
	"github.com/imskyd/go-frame-base/database"
	redisPkg "github.com/imskyd/go-frame-base/redis"
	"github.com/raven-ruiwen/go-helper/auth0"
	"github.com/sirupsen/logrus"
)

func init() {

}

// New create new router service
func New(prefix string, dbConfig *mysql.Config, jwtConfigFromFile *base.JwtConfig, auth0Config *auth0.Config, redisConfig *redisPkgBase.Config, smtpConfig *base.SmtpConfig) (srv *Service, err error) {
	actions.SetSmtpConfig(smtpConfig)

	db := database.NewMysql(dbConfig)
	srv = &Service{}
	srv.mysql = db
	srv.auth0Config = auth0Config
	srv.jwtConfig = jwtConfigFromFile
	//redis
	srv.redis = redisPkg.NewRedis(redisConfig)
	//mysqlClient.Client.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4").AutoMigrate(&Users{})
	//mysqlClient.Client.HasTable(&Users{})
	//mysqlClient.Client.HasTable(&Users{})

	logF := &logrus.JSONFormatter{}
	logrus.SetFormatter(logF)
	logger := logrus.StandardLogger()
	srv.logger = logger
	auth02.SetLogger(logger)
	srv.prefix = prefix
	return srv, nil
}

//func (srv *Service) syncAuth0Token() {
//	srv.auth0AccessTokenPersist()
//
//	ticker := time.NewTicker(10 * time.Minute)
//	for true {
//		select {
//		case <-ticker.C:
//			if tokenCache := GetAuth0AccessToken(); tokenCache == "" {
//				srv.auth0AccessTokenPersist()
//			}
//		}
//	}
//}

//func (srv *Service) auth0AccessTokenPersist() {
//	tokenInfo := auth02.GetAccessToken(srv.auth0Config.Domain, srv.auth0Config.ClientId, srv.auth0Config.ClientSecret)
//	if tokenInfo.Err != nil {
//		//create issue
//	} else {
//		token := tokenInfo.Info.AccessToken
//		expireTime := tokenInfo.Info.ExpiresIn - 20*60
//		cacheKey := auth02.CacheKey
//		err := redisClient.Set(context.Background(), cacheKey, token, time.Duration(expireTime)*time.Second).Err()
//		if err != nil {
//			srv.logger.Errorf("persist auth0 api access token error: %s", err.Error())
//		}
//	}
//}
