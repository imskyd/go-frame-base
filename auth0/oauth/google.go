package oauth

import "time"

type googleUser struct {
	Aud           string    `json:"aud"`
	Email         string    `json:"email"`
	EmailVerified bool      `json:"email_verified"`
	Exp           int       `json:"exp"`
	GivenName     string    `json:"given_name"`
	Iat           int       `json:"iat"`
	Iss           string    `json:"iss"`
	Locale        string    `json:"locale"`
	Name          string    `json:"name"`
	Nickname      string    `json:"nickname"`
	Picture       string    `json:"picture"`
	Sid           string    `json:"sid"`
	Sub           string    `json:"sub"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (u *googleUser) GetSub() string {
	return u.Sub
}

func (u *googleUser) GetChannel() string {
	return ChannelGoogle
}

func (u *googleUser) GetEmail() string {
	return u.Email
}

func (u *googleUser) GetEmailIsVerified() int {
	if u.EmailVerified {
		return EmailVerified
	}
	return EmailNotVerified
}

func (u *googleUser) GetName() string {
	return u.Name
}

func (u *googleUser) GetPicture() string {
	return u.Picture
}
