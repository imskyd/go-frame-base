package router

import (
	"encoding/json"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type BrowserUser struct {
	UserId     int64 `json:"user_id"`
	Role       Role  `json:"role"`
	SubId      int64 `json:"sub_id"`
	ExpireTime int64 `json:"expire_time"`
	Token      string
}

func (srv *Service) UserMustLoginMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, err := srv.getBrowserUser(c)
		if err != nil {
			srv.logger.Warnf("jwt parse error::" + err.Error())
			c.AbortWithStatusJSON(403, ResponseMsg{
				Code: 403,
				Msg:  "Forbidden",
				Data: "",
			})
			return
		}
	}
}

func (srv *Service) TeamBasicMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodGet {
			browserUser, _ := srv.getBrowserUser(c)
			tId := c.Param("team_id")
			teamId, _ := strconv.Atoi(tId)
			if !srv.IsUserHaveBasicAccess(int64(teamId), browserUser.UserId) {
				c.AbortWithStatusJSON(403, ResponseMsg{
					Code: 403,
					Msg:  "Forbidden",
					Data: "",
				})
				return
			}
		}
	}
}

func (srv *Service) TeamOperateMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodGet {
			browserUser, _ := srv.getBrowserUser(c)
			tId := c.Param("team_id")
			teamId, _ := strconv.Atoi(tId)
			if !srv.IsUserHaveOperateAccess(int64(teamId), browserUser.UserId) {
				c.AbortWithStatusJSON(403, ResponseMsg{
					Code: 403,
					Msg:  "Forbidden, only admin or operator can operate",
					Data: "",
				})
				return
			}
		}
	}
}

func (srv *Service) userStatusMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		browserUser, _ := srv.getBrowserUserFromContext(c)
		user, err := srv.getUserById(browserUser.UserId)
		if err == gorm.ErrRecordNotFound {
			c.AbortWithStatusJSON(403, ResponseMsg{
				Code: 403,
				Msg:  "UserNotFound",
				Data: "",
			})
			return
		}
		if user.Status == UserStatusDisabled {
			c.AbortWithStatusJSON(200, ResponseMsg{
				Code: 4009,
				Msg:  "Your account has been disabled. Please contact customer service",
				Data: "",
			})
			return
		}
	}
}

func (srv *Service) getBrowserUser(ctx *gin.Context) (*BrowserUser, error) {
	var token string
	if cookieToken, err := ctx.Cookie("token"); err == nil {
		token = cookieToken
	} else {
		return nil, errors.New("cookie token error")
	}
	//decode jwt
	publicKey := srv.jwtConfig.PublicKey
	publicKey = strings.Replace(publicKey, "||", "\n", -1)
	pb, err := jwt.ParseRSAPublicKeyFromPEM([]byte(publicKey))
	if err != nil {
		return nil, err
	}
	ret, ok := parse(token, pb)
	if !ok {
		return nil, errors.New("token parse error:" + err.Error())
	}
	//parse to global struct
	jsonByte, _ := json.Marshal(ret)
	var browserUser BrowserUser
	err = json.Unmarshal(jsonByte, &browserUser)
	if err != nil {
		return nil, err
	}
	if time.Now().Unix() > browserUser.ExpireTime {
		return nil, errors.New("forbidden, token has expired, token is:" + token)
	}
	browserUser.Token = token
	return &browserUser, nil
}
