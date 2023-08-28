package oauth

import "time"

type twitterUser struct {
	Aud       string    `json:"aud"`
	Email     string    `json:"email"`
	Exp       int       `json:"exp"`
	Iat       int       `json:"iat"`
	Iss       string    `json:"iss"`
	Name      string    `json:"name"`
	Nickname  string    `json:"nickname"`
	Picture   string    `json:"picture"`
	Sid       string    `json:"sid"`
	Sub       string    `json:"sub"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (u *twitterUser) GetSub() string {
	return u.Sub
}

func (u *twitterUser) GetChannel() string {
	return ChannelTwitter
}

func (u *twitterUser) GetEmail() string {
	return u.Email
}

func (u *twitterUser) GetEmailIsVerified() int {
	return EmailVerified
}

func (u *twitterUser) GetName() string {
	return u.Name
}

func (u *twitterUser) GetPicture() string {
	return u.Picture
}
