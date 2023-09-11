package router

import (
	redisPkgBase "github.com/CoinSummer/go-base/cache/redis"
	"github.com/CoinSummer/go-base/db/mysql"
	"github.com/imskyd/go-frame-base/actions"
	auth02 "github.com/imskyd/go-frame-base/auth0"
	"github.com/imskyd/go-frame-base/database"
	redisPkg "github.com/imskyd/go-frame-base/redis"
	"github.com/imskyd/go-frame-base/types"
	"github.com/raven-ruiwen/go-helper/auth0"
	"github.com/sirupsen/logrus"
)

func init() {

}

// New create new router service
func New(appName string, prefix string, dbConfig *mysql.Config, jwtConfigFromFile *types.JwtConfig, auth0Config *auth0.Config, redisConfig *redisPkgBase.Config, smtpConfig *types.SmtpConfig) (srv *Service, err error) {
	actions.SetSmtpConfig(smtpConfig)

	db := database.NewMysql(dbConfig)
	srv = &Service{}
	srv.mysql = db
	srv.auth0Config = auth0Config
	srv.jwtConfig = jwtConfigFromFile
	//redis
	srv.redis = redisPkg.NewRedis(redisConfig)

	logF := &logrus.JSONFormatter{}
	logrus.SetFormatter(logF)
	logger := logrus.StandardLogger()
	srv.logger = logger
	auth02.SetLogger(logger)
	actions.SetActionLogger(logger)
	srv.prefix = prefix
	srv.appName = appName
	return srv, nil
}

func (srv *Service) ModelCheck() {
	if !srv.mysql.Client.HasTable(&Users{}) {
		srv.mysql.Client.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4").AutoMigrate(&Users{})
	}
	if !srv.mysql.Client.HasTable(&UsersSub{}) {
		srv.mysql.Client.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4").AutoMigrate(&UsersSub{})
	}
	if !srv.mysql.Client.HasTable(&Team{}) {
		srv.mysql.Client.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4").AutoMigrate(&Team{})
	}
	if !srv.mysql.Client.HasTable(&TeamUser{}) {
		srv.mysql.Client.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4").AutoMigrate(&TeamUser{})
	}
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
