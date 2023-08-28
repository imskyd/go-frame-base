package oauth

import (
	"encoding/json"
	"errors"
	"strings"
)

type User interface {
	GetChannel() string
	GetSub() string
	GetEmail() string
	GetEmailIsVerified() int
	GetName() string
	GetPicture() string
}

type Base struct {
	Sub string `json:"sub"`
}

const (
	ChannelGoogle  = "google-oauth2"
	ChannelTwitter = "twitter"
	ChannelAuth0   = "auth0"
)

const (
	EmailNotVerified = iota
	EmailVerified
)

var ErrorNoSub = errors.New("parse error: info does not have column [sub]")

func GetOauthUser(info string) (User, error) {
	var base Base
	_ = json.Unmarshal([]byte(info), &base)

	if base.Sub == "" {
		return nil, ErrorNoSub
	}

	split := strings.Split(base.Sub, "|")

	switch split[0] {
	case ChannelGoogle:
		var user googleUser
		_ = json.Unmarshal([]byte(info), &user)
		return &user, nil
	case ChannelTwitter:
		var user twitterUser
		_ = json.Unmarshal([]byte(info), &user)
		return &user, nil
	case ChannelAuth0:
		var user auth0User
		_ = json.Unmarshal([]byte(info), &user)
		return &user, nil
	default:
		return nil, ErrorNoSub
	}
}
