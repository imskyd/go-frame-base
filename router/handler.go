package router

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/CoinSummer/go-notify"
	"github.com/CoinSummer/go-notify/email"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/imroc/req"
	"github.com/imskyd/go-frame-base/actions"
	auth02 "github.com/imskyd/go-frame-base/auth0"
	oauth2 "github.com/imskyd/go-frame-base/auth0/oauth"
	"github.com/jinzhu/gorm"
	"github.com/raven-ruiwen/go-helper/auth0"
	"net/http"
	"strings"
	"time"
)

func (srv *Service) login(ctx *gin.Context) {
	state, _ := auth0.GenerateRandomState()
	_auth0Authenticator, _ := srv.newAuth0Authenticator(ctx)

	loginUrl := _auth0Authenticator.GetAuthCodeURL(state)
	// Save the state inside the session.
	session := sessions.Default(ctx)
	session.Set("state", state)
	if err := session.Save(); err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	srv.markUserLoginRedirectUrl(ctx, state)
	ctx.Redirect(http.StatusTemporaryRedirect, loginUrl)
}

func (srv *Service) teamInvite(ctx *gin.Context) {
	code := ctx.Query("code")
	teamUser, err := srv.getTeamUserByCode(code)
	if err != nil {
		ReturnErrorWithMsg(ctx, InternalError, "no invitation")
		return
	}
	if time.Now().Unix() > teamUser.ExpireAt {
		ReturnErrorWithMsg(ctx, InternalError, "invitation expired")
		return
	}
	srv.login(ctx)
}

func (srv *Service) loginLocal(ctx *gin.Context) {
	//set cookie
	token, err := srv.createToken(4, 1)
	if err != nil {
		srv.logger.Errorf("createToken error: %s", err.Error())

		msg := "Login failed, please try again"
		ctx.Redirect(http.StatusMovedPermanently, "/404?errormsg="+base64.StdEncoding.EncodeToString([]byte(msg)))
		return
	}
	ctx.SetCookie("token", token, 86400*3, "/", "", false, false)
	ctx.Redirect(http.StatusMovedPermanently, "/loginredirect")
	return
}

func (srv *Service) auth0CallBack(ctx *gin.Context) {
	var options Auth0CallbackOptions

	session := sessions.Default(ctx)
	if ctx.Query("state") != session.Get("state") {
		ctx.String(http.StatusBadRequest, "Invalid state parameter.")
		return
	}
	state := ctx.Query("state")

	//try decode by jwt
	publicKey := srv.jwtConfig.PublicKey
	publicKey = strings.Replace(publicKey, "||", "\n", -1)
	pb, err := jwt.ParseRSAPublicKeyFromPEM([]byte(publicKey))
	if err == nil {
		ret, ok := parse(state, pb)
		if ok {
			//parse to global struct
			jsonByte, _ := json.Marshal(ret)
			var invite InviteUser
			if err = json.Unmarshal(jsonByte, &invite); err == nil {
				//is invite user
				options.isInviteUser = true
				options.inviteUserId = invite.UserId
			}
		}
	}

	_auth0Authenticator, _ := srv.newAuth0Authenticator(ctx)
	code := ctx.Query("code")
	profile, _, err := _auth0Authenticator.GetProfileFromCode(ctx, code)
	if err != nil {

		msg := "Failed to get your information, please try again or use another account"
		ctx.Redirect(http.StatusMovedPermanently, "/404?errormsg="+base64.StdEncoding.EncodeToString([]byte(msg)))
		return
	}

	pj, _ := json.Marshal(profile)
	srv.logger.Infof("[User Login] data: %s", string(pj))
	oauthUser, err := oauth2.GetOauthUser(string(pj))
	if err != nil {
		srv.logger.Errorf("GetOauthUser error: %s, data: %s", err.Error(), string(pj))

		msg := "Failed to get your information, please try again or use another account"
		ctx.Redirect(http.StatusMovedPermanently, "/404?errormsg="+base64.StdEncoding.EncodeToString([]byte(msg)))
		return
	}
	if oauthUser.GetEmail() == "" {
		msg := "Your third-party account is not bound to an email address. Please log in or use another account after binding"
		ctx.Redirect(http.StatusMovedPermanently, "/404?errormsg="+base64.StdEncoding.EncodeToString([]byte(msg)))
		return
	}

	var u Users
	err = srv.mysql.Client.Model(&u).Where("email = ?", oauthUser.GetEmail()).Find(&u).Error
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			srv.logger.Errorf("Auth0CallBack Find user error: %s, data: %s", err.Error(), string(pj))

			msg := "Find your user account failed, Please contact the administrator"
			ctx.Redirect(http.StatusMovedPermanently, "/404?errormsg="+base64.StdEncoding.EncodeToString([]byte(msg)))
			return
		}
		//create user
		options.isNewUser = true
		u.Email = oauthUser.GetEmail()
		u.EmailVerified = oauthUser.GetEmailIsVerified()
		u.Name = oauthUser.GetName()
		u.Picture = oauthUser.GetPicture()
		now := time.Now()
		u.LastLoginAt = &now
		u.TimeFormat = UserTimeFormatRFC3339
		u.Status = UserStatusEnabled
		u.InviteCode = srv.generateInviteCode(&Users{})
		if options.isInviteUser {
			u.InviteUserId = options.inviteUserId
		}
		srv.mysql.Client.Omit("msg_monthly_used", "timezone", "msg_send_all").Create(&u)
		teamData := make(map[string]interface{})
		teamData["status"] = TeamUserStatusJoined
		teamData["user_id"] = u.Id
		teamData["invite_email"] = ""
		srv.mysql.Client.Model(TeamUser{}).Where("invite_email = ?", u.Email).Updates(teamData)
	}
	//create user sub info
	sub := UsersSub{
		UserIdMain:    u.Id,
		Channel:       oauthUser.GetChannel(),
		Sub:           oauthUser.GetSub(),
		EmailVerified: oauthUser.GetEmailIsVerified(),
	}
	srv.mysql.Client.Model(UsersSub{}).Where("user_id_main = ? and sub = ?", u.Id, oauthUser.GetSub()).FirstOrCreate(&sub)
	//set cookie
	token, err := srv.createToken(u.Id, sub.Id)
	if err != nil {
		srv.logger.Errorf("createToken error: %s", err.Error())

		msg := "Login failed, please try again"
		ctx.Redirect(http.StatusMovedPermanently, "/404?errormsg="+base64.StdEncoding.EncodeToString([]byte(msg)))
		return
	}
	ctx.SetCookie("token", token, 86400*3, "/", "", false, false)
	//set user id to mark user login action in middleware
	go func() {
		if options.isNewUser && options.isInviteUser && oauthUser.GetEmailIsVerified() == oauth2.EmailVerified {
			//
		}
		now := time.Now()
		srv.mysql.Client.Model(Users{}).Where("id = ?", u.Id).Update(&Users{LastLoginAt: &now})
		//if err != nil {
		//	srv.logger.Warnf("update last trigger at error: %s", err.Error())
		//}
		//
		//action := types.Login{}
		//action.Init()
		//action.SetOriginalDataToInit(u.Id, ctx)
		//detail := action.GetDetail()
		//
		//userLog := UsersOperationLogs{
		//	Method: detail.Method,
		//	Url:    detail.Url,
		//	UserId: detail.UserId,
		//	Action: detail.LogAction,
		//	Data:   detail.LogData,
		//	Remark: detail.LogRemark,
		//}
		//if err := srv.mysql.Client.Create(&userLog).Error; err != nil {
		//	srv.logger.Warnf("create user login log error: %s", err.Error())
		//}
	}()
	//determining user status
	if u.Status == UserStatusDisabled {
		msg := "Your account has been disabled. Please contact customer service"
		ctx.Redirect(http.StatusMovedPermanently, "/404?errormsg="+base64.StdEncoding.EncodeToString([]byte(msg)))
		return
	}
	if oauthUser.GetEmailIsVerified() == oauth2.EmailVerified {
		url := srv.getUserLoginRedirectUrl(ctx, state)
		fmt.Println("redirect to .." + url)
		ctx.Redirect(http.StatusMovedPermanently, url)
		return
	} else {
		ctx.Redirect(http.StatusMovedPermanently, "/account/verify")
		return
	}
}

func (srv *Service) getProfile(ctx *gin.Context) {
	browserUser, _ := srv.getBrowserUserFromContext(ctx)
	var u Users
	if err := srv.mysql.Client.Where("id = ?", browserUser.UserId).Find(&u).Error; err != nil {
		ReturnErrorWithMsg(ctx, InvalidFormData, "User not found")
		return
	}

	go func() {
		var sub UsersSub
		err := srv.mysql.Client.Where("user_id_main = ? and email_verified = ?", browserUser.UserId, oauth2.EmailNotVerified).First(&sub).Error
		if err != nil && err == gorm.ErrRecordNotFound {
			return
		}
		auth0User, err := auth02.GetUserInfo(srv.auth0Config.Domain, sub.Sub, srv.getAuth0AccessToken())
		if err != nil {
			srv.logger.Errorf("auth02.GetUserInfo error: %s", err.Error())
			return
		}
		if auth0User.EmailVerified == true {
			data := make(map[string]interface{})
			data["email_verified"] = oauth2.EmailVerified
			srv.mysql.Client.Model(UsersSub{}).Where("id = ?", sub.Id).Updates(data)
			srv.mysql.Client.Model(Users{}).Where("id = ?", browserUser.UserId).Updates(data)
			if u.InviteUserId != 0 {
				//
			}
		}
	}()
	//invite link
	u.InviteLink = fmt.Sprintf("https://%s/api/v1/invite?code=%s", ctx.Request.Host, u.InviteCode)
	//user registration
	var subs []UsersSub
	_ = srv.mysql.Client.Where("user_id_main = ?", browserUser.UserId).Find(&subs).Error
	for _, s := range subs {
		if s.Channel == UserSubChannelAuth0 {
			u.Registration = UserSubChannelAuth0
		}
	}
	if u.Registration == "" {
		u.Registration = subs[0].Channel
	}

	ReturnSuccess(ctx, u)
}

func (srv *Service) newAuth0Authenticator(ctx *gin.Context) (*auth0.Authenticator, error) {
	callbackUrl := fmt.Sprintf("https://%s/api/v1/login/callback", ctx.Request.Host)
	_auth0Authenticator, err := auth0.NewAuthenticator(srv.auth0Config.Domain, srv.auth0Config.ClientId, srv.auth0Config.ClientSecret, callbackUrl)
	return _auth0Authenticator, err
}

func (srv *Service) resetPassword(ctx *gin.Context) {
	browserUser, _ := srv.getBrowserUserFromContext(ctx)
	u, _ := srv.getUserById(browserUser.UserId)

	authApiUrl := "https://" + srv.auth0Config.Domain + "/dbconnections/change_password"
	header := req.Header{
		"Content-Type": "application/json",
	}
	data := make(map[string]string)
	data["client_id"] = srv.auth0Config.ClientId
	data["email"] = u.Email
	data["connection"] = srv.auth0Config.DbConnection

	j, _ := json.Marshal(data)

	resp, err := req.Post(authApiUrl, string(j), header)
	if err != nil {
		srv.logger.Errorf("send reset pwd email error:%s", err.Error())
		ReturnErrorWithMsg(ctx, InternalError, "")
		return
	}
	fmt.Println(resp.Response().Status)
	ReturnSuccess(ctx, struct{}{})
}

func (srv *Service) updateProfile(ctx *gin.Context) {
	browserUser, _ := srv.getBrowserUserFromContext(ctx)
	var u UpdateUserParams
	if err := ctx.ShouldBindWith(&u, binding.JSON); err != nil {
		ReturnErrorWithMsg(ctx, InvalidFormData, err.Error())
		return
	}
	if u.UserName == "" {
		ReturnErrorWithMsg(ctx, InvalidFormData, "Username can not be null")
		return
	}

	var f Users
	if err := srv.mysql.Client.Model(Users{}).Where("id != ? and username = ?", browserUser.UserId, u.UserName).First(&f).Error; err == nil {
		ReturnErrorWithMsg(ctx, InternalError, "Username already exist")
		return
	}
	data := make(map[string]interface{})
	if u.UserName != "" {
		data["username"] = u.UserName
	}
	if u.Twitter != "" {
		data["twitter"] = u.Twitter
	}
	if u.Telegram != "" {
		data["telegram"] = u.Telegram
	}
	if u.Discord != "" {
		data["discord"] = u.Discord
	}
	if u.Bio != "" {
		data["bio"] = u.Bio
	}
	base := srv.mysql.Client.Model(Users{}).Where("id = ?", browserUser.UserId)
	loginUser, _ := srv.getUserById(browserUser.UserId)
	if loginUser.UserName != "" {
		base = base.Omit("username")
	}
	err := base.Updates(data).Error
	if err != nil {
		ReturnErrorWithMsg(ctx, InternalError, "Operation failed")
		return
	}
	ReturnSuccess(ctx, struct{}{})
}

func (srv *Service) userDelete(ctx *gin.Context) {
	browserUser, _ := srv.getBrowserUserFromContext(ctx)
	srv.mysql.Client.Model(Users{}).Where("id = ?", browserUser.UserId).Delete(Users{})
	ReturnSuccess(ctx, struct{}{})
}

func (srv *Service) logout(ctx *gin.Context) {
	scheme := "https"
	redirectUrl := scheme + "://" + ctx.Request.Host
	url := fmt.Sprintf("https://%s/v2/logout?client_id=%s&returnTo=%s", srv.auth0Config.Domain, srv.auth0Config.ClientId, redirectUrl)
	ctx.Redirect(http.StatusTemporaryRedirect, url)
}

//func (srv *Service) ResendVerifyEmail(ctx *gin.Context) {
//	browserUser, _ := getBrowserUserFromContext(ctx)
//	u, _ := GetUserById(browserUser.UserId)
//	if u.EmailVerified == 1 {
//		ReturnErrorWithMsg(ctx, InvalidFormData, "You have been verified")
//		return
//	}
//	var sub UsersSub
//	err := srv.mysql.Client.Model(UsersSub{}).Where("user_id_main = ? and channel = ?", browserUser.UserId, UserSubChannelAuth0).First(&sub).Error
//	if err != nil {
//		ReturnErrorWithMsg(ctx, InternalError, "Failed to find the user. Please check to see if you have a password to register the user")
//		return
//	}
//
//	auth0ApiToken, _ := redisClient.Get(context.Background(), auth02.CacheKey).Result()
//
//	authApiUrl := "https://" + srv.auth0Config.Domain + "/api/v2/jobs/verification-email"
//	header := req.Header{
//		"Content-Type":  "application/json",
//		"Authorization": "Bearer " + auth0ApiToken,
//	}
//	data := make(map[string]string)
//	data["client_id"] = srv.auth0Config.ClientId
//	data["user_id"] = sub.Sub
//
//	j, _ := json.Marshal(data)
//
//	resp, err := req.Post(authApiUrl, string(j), header)
//	if err != nil {
//		srv.logger.Errorf("send reset pwd email error:%s", err.Error())
//		ReturnErrorWithMsg(ctx, InternalError, "")
//		return
//	}
//	fmt.Println(resp.Response().StatusCode)
//
//	if resp.Response().StatusCode == 201 {
//		ReturnSuccess(ctx, struct{}{})
//		return
//	} else {
//		ReturnErrorWithMsg(ctx, InternalError, "Send email failed  ")
//		return
//	}
//}

func (srv *Service) sendEmail(ctx *gin.Context, emailAddr string, code string) {
	host := ctx.Request.Host
	emailInfo := email.Info{
		Subject: fmt.Sprintf("[ %s Invitation ]", srv.appName),
		Content: "Please click the link to join the group: " + host + "/api/v1/team/invite?code=" + code,
	}
	j, _ := json.Marshal(emailInfo)

	param := make(map[string]string)
	param["platform"] = notify.PlatformEmail
	param["token"] = emailAddr
	param["channel"] = emailAddr
	param["msg"] = string(j)

	a := actions.Action{
		Type:  "Notify",
		Param: param,
	}
	a.Init()
	err := a.Run()
	if err != nil && param["platform"] != notify.PlatformDiscord {
		fmt.Println(err.Error())
	}
}
