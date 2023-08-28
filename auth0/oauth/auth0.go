package oauth

import "time"

type auth0User struct {
	Aud           string    `json:"aud"`
	Email         string    `json:"email"`
	EmailVerified bool      `json:"email_verified"`
	Exp           int       `json:"exp"`
	Iat           int       `json:"iat"`
	Iss           string    `json:"iss"`
	Name          string    `json:"name"`
	Nickname      string    `json:"nickname"`
	Picture       string    `json:"picture"`
	Sid           string    `json:"sid"`
	Sub           string    `json:"sub"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (u *auth0User) GetSub() string {
	return u.Sub
}

func (u *auth0User) GetChannel() string {
	return ChannelAuth0
}

func (u *auth0User) GetEmail() string {
	return u.Email
}

func (u *auth0User) GetEmailIsVerified() int {
	if u.EmailVerified {
		return EmailVerified
	}
	return EmailNotVerified
}

func (u *auth0User) GetName() string {
	return u.Name
}

func (u *auth0User) GetPicture() string {
	return u.Picture
}
