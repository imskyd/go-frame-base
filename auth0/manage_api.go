package auth0

import (
	"encoding/json"
	"errors"
	"github.com/imroc/req"
	"github.com/sirupsen/logrus"
	"time"
)

var (
	logger *logrus.Logger
)

func SetLogger(_logger *logrus.Logger) {
	logger = _logger
}

const CacheKey = "ChainBotApi:Auth0AccessKey"

type TokenResp struct {
	Info TokenInfo
	Err  error
}

type TokenInfo struct {
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`

	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	ErrorUri         string `json:"error_uri"`
}

func GetAccessToken(host string, clientId string, clientSecret string) TokenResp {
	url := "https://" + host + "/oauth/token"

	body := make(map[string]string)
	body["client_id"] = clientId
	body["client_secret"] = clientSecret
	body["audience"] = "https://" + host + "/api/v2/"
	body["grant_type"] = "client_credentials"

	dataJson, _ := json.Marshal(body)
	header := req.Header{
		"Content-Type": "application/json",
	}
	resp, err := req.Post(url, string(dataJson), header)

	if err != nil {
		logger.Errorf("auth0 GetAccessToken error: %s", err.Error())
		return TokenResp{Info: TokenInfo{}, Err: err}
	}

	var info TokenInfo
	if err := json.Unmarshal(resp.Bytes(), &info); err != nil {
		logger.Errorf("auth0 GetAccessToken json.Unmarshal error: %s", err.Error())
		return TokenResp{Info: TokenInfo{}, Err: err}
	}

	if info.Error != "" {
		infoJ, _ := json.Marshal(info)
		logger.Errorf("auth0 GetAccessToken info error: %s", string(infoJ))
		return TokenResp{Info: TokenInfo{}, Err: errors.New(info.Error)}
	}

	return TokenResp{Info: info, Err: nil}
}

type UserInfo struct {
	CreatedAt     time.Time `json:"created_at"`
	Email         string    `json:"email"`
	EmailVerified bool      `json:"email_verified"`
	Identities    []struct {
		UserId     string `json:"user_id"`
		Provider   string `json:"provider"`
		Connection string `json:"connection"`
		IsSocial   bool   `json:"isSocial"`
	} `json:"identities"`
	Name        string    `json:"name"`
	Nickname    string    `json:"nickname"`
	Picture     string    `json:"picture"`
	UpdatedAt   time.Time `json:"updated_at"`
	UserId      string    `json:"user_id"`
	LastIp      string    `json:"last_ip"`
	LastLogin   time.Time `json:"last_login"`
	LoginsCount int       `json:"logins_count"`
}

func GetUserInfo(host string, userId string, token string) (UserInfo, error) {
	url := "https://" + host + "/api/v2/users/" + userId

	header := req.Header{
		"Authorization": "Bearer " + token,
	}
	resp, err := req.Get(url, header)

	if err != nil {
		logger.Errorf("auth0 GetUserInfo error: %s", err.Error())
		return UserInfo{}, err
	}

	var info UserInfo
	if err := json.Unmarshal(resp.Bytes(), &info); err != nil {
		logger.Errorf("auth0 GetUserInfo json.Unmarshal error: %s", err.Error())
		return UserInfo{}, err
	}

	return info, nil
}
