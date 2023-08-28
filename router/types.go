package router

import (
	"github.com/deng00/go-base/db/mysql"
	"github.com/gin-gonic/gin"
	redisV8 "github.com/go-redis/redis/v8"
	"github.com/imskyd/go-frame-base/base"
	"github.com/raven-ruiwen/go-helper/auth0"
	"github.com/sirupsen/logrus"
	"strings"
)

const (
	RequestExpirationSeconds = 5
)

type ResponseMsg struct {
	Code ErrCode     `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type ParserMsg struct {
	Type     string `json:"type"`
	Selector string `json:"selector"`
	Abi      string `json:"abi"`
	Template string `json:"template"`
	TxMsg    string `json:"tx_msg"`
}

type ErrCode int

const (
	InvalidFormData ErrCode = 4000 + iota
	ChainError
	SignError
	AuthError
	InternalError
	ParseError
	ParseTxError
	ParserUnMatchErr
	TransactionIllegal
	HeaderError
	ExpiredRequest
	IllegalAccess
	ParseRequestError
	OperationFailed
	BotInMaintenance
)

var ErrorMsgMap = map[ErrCode]string{
	InvalidFormData:    "Invalid form data",
	ChainError:         "Unsupported chain",
	SignError:          "Sign error",
	AuthError:          "Auth error",
	InternalError:      "Internal error",
	ParseError:         "Parse function error",
	ParseTxError:       "Json parse error",
	TransactionIllegal: "TransactionIllegal",
	IllegalAccess:      "Illegal access",
	ExpiredRequest:     "Expired request",
	HeaderError:        "Invalid header",
	ParserUnMatchErr:   "Parser unMatched with tx",
	ParseRequestError:  "Parser request param error",
	OperationFailed:    "Operation Failed",
	BotInMaintenance:   "This Bot is currently under maintenance",
}

type MyError struct {
	Code ErrCode
}

func (e *MyError) Error() string {
	if ErrorMsgMap[e.Code] == "" {
		return "unknown error"
	}
	return ErrorMsgMap[e.Code]
}

func Error(code ErrCode) *MyError {
	return &MyError{Code: code}
}
func ReturnError(c *gin.Context, code ErrCode) {
	c.AbortWithStatusJSON(200, ResponseMsg{
		Code: code,
		Msg:  Error(code).Error(),
		Data: "",
	})
}

func ReturnErrorWithMsg(c *gin.Context, code ErrCode, msg string) {
	c.AbortWithStatusJSON(200, ResponseMsg{
		Code: code,
		Msg:  msg,
		Data: "",
	})
}

func ReturnSuccess(c *gin.Context, data interface{}) {
	c.AbortWithStatusJSON(200, ResponseMsg{
		Code: 0,
		Msg:  "success",
		Data: data,
	})
}

type PaginatorRequestParam struct {
	Page    int `form:"page"`
	PerPage int `form:"per_page" binding:"max=1000"`
}

type Paginator struct {
	Data        interface{} `json:"data"`
	CurrentPage int         `json:"current_page"`
	PerPage     string      `json:"per_page"`
	Total       int         `json:"total"`
}

func GetNewPaginator() *PaginatorRequestParam {
	return &PaginatorRequestParam{
		Page:    1,
		PerPage: 20,
	}
}

type Whitelist map[string][]string

func (w Whitelist) Add(apis string, ips string) {
	apiArr := strings.Split(apis, ",")
	ipArr := strings.Split(ips, ",")
	for _, api := range apiArr {
		w[api] = append(w[api], ipArr...)
	}
}
func (w Whitelist) Allow(api string, ip string) bool {
	ips, ok := w[api]
	if !ok {
		return false
	}
	for _, ipPattern := range ips {
		if ipPattern == ip {
			return true
		}
	}
	return false
}

type Service struct {
	mysql       *mysql.MySQL
	auth0Config *auth0.Config
	jwtConfig   *base.JwtConfig
	logger      *logrus.Logger
	redis       *redisV8.Client
	prefix      string
}

type UpdateUserParams struct {
	UserName string `gorm:"column:username" json:"username,omitempty" binding:"required,min=3,max=20,usernamevalidator"`
	Twitter  string `gorm:"column:twitter" json:"twitter,omitempty"`
	Telegram string `gorm:"column:telegram" json:"telegram,omitempty"`
	Discord  string `gorm:"column:discord" json:"discord,omitempty"`
	Bio      string `gorm:"column:bio" json:"bio,omitempty"`
}

type Auth0CallbackOptions struct {
	isNewUser    bool
	isInviteUser bool
	inviteUserId int64
}

type ParamDealTeamInvitation struct {
	Id       int64  `json:"id"`
	Response string `json:"response"`
}

type ParamAddTeamMember struct {
	TeamId int64  `json:"team_id" binding:"required"`
	Email  string `json:"email" binding:"required"`
	Role   Role   `json:"role" binding:"required"`
}

type ParamUpdateTeamMember struct {
	Role Role `json:"role,omitempty" binding:"required"`
}

type ParamContractViewMove struct {
	NewTeamId int64 `json:"new_team_id" binding:"required"`
}

type InviteUser struct {
	UserId int64 `json:"user_id"`
}

type TimeFormat string

const (
	UserTimeFormatRFC822   TimeFormat = "02 Jan 06 15:04 MST"
	UserTimeFormatRFC3339  TimeFormat = "2006-01-02 15:04:05Z07:00"
	UserTimeFormatUnixDate TimeFormat = "Mon Jan _2 15:04:05 MST 2006"
	UserTimeFormatMST      TimeFormat = "01/02/2006 15:04:05 MST"
)

type UserStatus string

const (
	UserStatusEnabled  UserStatus = "Enabled"
	UserStatusDisabled UserStatus = "Disabled"
)
