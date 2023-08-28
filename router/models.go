package router

import (
	"time"
)

type Users struct {
	Id            int64      `gorm:"column:id" json:"id,omitempty"`
	Email         string     `gorm:"column:email;not null" json:"email,omitempty"`
	EmailVerified int        `gorm:"column:email_verified;type:tinyint;not null" json:"email_verified"`
	Name          string     `gorm:"column:name;size:50;not null" json:"-"`
	UserName      string     `gorm:"column:username;size:50;default:'';not null" json:"username"`
	Picture       string     `gorm:"column:picture;type:text;not null" json:"picture,omitempty"`
	TimeZone      *int64     `gorm:"column:timezone;default:0;not null;type:int" json:"-"`
	TimeFormat    TimeFormat `gorm:"column:time_format;default:'RFC3339';not null" json:"-"`
	Status        UserStatus `gorm:"column:status;size:20;default:'Enabled';not null" json:"-"`
	InviteLink    string     `gorm:"-" json:"-"`
	InviteUserId  int64      `gorm:"column:invite_user_id;default:0;not null" json:"-"`
	InviteCode    string     `gorm:"column:invite_code;size:50;default:'';not null" json:"-"`
	Registration  string     `gorm:"-" json:"-"`
	Twitter       string     `gorm:"column:twitter;size:300;default:'';not null" json:"twitter"`
	Telegram      string     `gorm:"column:telegram;size:300;default:'';not null" json:"telegram"`
	Discord       string     `gorm:"column:discord;size:300;default:'';not null" json:"discord"`
	Bio           string     `gorm:"column:bio;size:1000;not null" json:"bio"`
	CreatedAt     *time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP;not null" json:"-"`
	UpdatedAt     *time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP;not null" json:"-"`
	DeletedAt     *time.Time `gorm:"column:deleted_at" json:"-"`
	LastLoginAt   *time.Time `gorm:"column:last_login_at;default:CURRENT_TIMESTAMP;not null" json:"-"`
}

func (m *Users) TableName() string {
	return "base_users"
}

func (srv *Service) getUserById(id int64) (Users, error) {
	var u Users
	err := srv.mysql.Client.Model(u).Where("id = ?", id).First(&u).Error
	return u, err
}

func (srv *Service) getUserByUsername(name string) (Users, error) {
	var u Users
	err := srv.mysql.Client.Model(u).Where("name = ?", name).First(&u).Error
	return u, err
}

func (srv *Service) getUserByEmail(email string) (Users, error) {
	var u Users
	err := srv.mysql.Client.Model(u).Where("email = ?", email).First(&u).Error
	return u, err
}

type PublicUser struct {
	Id       int64  `gorm:"column:id" json:"id"`
	UserName string `gorm:"column:username" json:"username"`
	Picture  string `gorm:"column:picture" json:"picture"`
	Email    string `gorm:"column:email" json:"email"`
}

func (m *PublicUser) TableName() string {
	return "base_users"
}

const (
	UserSubChannelAuth0   = "auth0"
	UserSubChannelGoogle  = "google-oauth2"
	UserSubChannelTwitter = "twitter"
)

type UsersSub struct {
	Id            int64  `gorm:"column:id" json:"id,omitempty"`
	UserIdMain    int64  `gorm:"column:user_id_main" json:"user_id_main,omitempty"`
	Channel       string `gorm:"column:channel" json:"-"`
	Sub           string `gorm:"column:sub" json:"-"`
	EmailVerified int    `gorm:"column:email_verified" json:"-"`
}

func (m *UsersSub) TableName() string {
	return "base_users_sub"
}

type UsersOperationLogs struct {
	Id          int64  `gorm:"column:id" json:"id"`
	Method      string `gorm:"column:method" json:"method"`
	Url         string `gorm:"column:url" json:"url"`
	UserId      int64  `gorm:"column:user_id" json:"user_id"`
	Action      string `gorm:"column:action" json:"action"`
	OperationId string `gorm:"column:operation_id" json:"operation_id"`
	Data        string `gorm:"column:data" json:"data"`
	Remark      string `gorm:"column:remark" json:"remark"`
	Ip          string `gorm:"column:ip" json:"ip"`
}

func (m *UsersOperationLogs) TableName() string {
	return "cb_users_operation_logs"
}

type Team struct {
	Id          int64      `gorm:"column:id" json:"id"`
	DisplayName string     `gorm:"column:display_name" json:"display_name" binding:"required"`
	Handle      string     `gorm:"column:handle" json:"handle" binding:"required"`
	UserId      int64      `gorm:"column:user_id" json:"user_id"`
	Twitter     string     `gorm:"column:twitter" json:"twitter"`
	Telegram    string     `gorm:"column:telegram" json:"telegram"`
	Discord     string     `gorm:"column:discord" json:"discord"`
	Bio         string     `gorm:"column:bio" json:"bio"`
	CreatedAt   *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   *time.Time `gorm:"column:updated_at" json:"-"`
	DeletedAt   *time.Time `gorm:"column:deleted_at" json:"-"`
}

func (m *Team) TableName() string {
	return "base_teams"
}

func (srv *Service) getTeamById(id int64) (Team, error) {
	var r Team
	err := srv.mysql.Client.Model(r).Where("id = ?", id).First(&r).Error
	return r, err
}

type TeamUser struct {
	Id          int64          `gorm:"column:id" json:"id,omitempty"`
	TeamId      int64          `gorm:"column:team_id" json:"team_id" binding:"required"`
	Team        Team           `json:"team,omitempty" gorm:"foreignKey:TeamId;references:Id"`
	UserId      int64          `gorm:"column:user_id" json:"user_id,omitempty" binding:"required"`
	User        PublicUser     `json:"user,omitempty" gorm:"foreignKey:UserId;references:Id"`
	Role        Role           `gorm:"column:role" json:"role" binding:"required"`
	Status      TeamUserStatus `gorm:"column:status" json:"status,omitempty"`
	InviteEmail string         `gorm:"column:invite_email" json:"-"`
	InviteCode  string         `gorm:"column:invite_code" json:"-"`
	ExpireAt    int64          `gorm:"column:expire_at" json:"-"`
	CreatedAt   *time.Time     `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   *time.Time     `gorm:"column:updated_at" json:"-"`
	DeletedAt   *time.Time     `gorm:"column:deleted_at" json:"-"`
}

func (srv *Service) getTeamUserById(id int64) (TeamUser, error) {
	var r TeamUser
	err := srv.mysql.Client.Model(r).Where("id = ?", id).First(&r).Error
	return r, err
}

func (srv *Service) getTeamUserByCode(code string) (TeamUser, error) {
	var r TeamUser
	err := srv.mysql.Client.Model(r).Where("invite_code = ?", code).First(&r).Error
	return r, err
}

func (srv *Service) getTeamUser(teamId int64, userId int64) (TeamUser, error) {
	var r TeamUser
	err := srv.mysql.Client.Model(r).Where("team_id = ? and user_id = ?", teamId, userId).First(&r).Error
	return r, err
}

func (srv *Service) isUserHaveOperateAccess(teamId int64, userId int64) bool {
	if err := srv.mysql.Client.Model(TeamUser{}).Where("team_id = ? and user_id = ? and (role = ? or role = ?)", teamId, userId, RoleAdmin, RoleOperator).First(&TeamUser{}).Error; err != nil {
		return false
	}
	return true
}

func (srv *Service) isUserHaveBasicAccess(teamId int64, userId int64) bool {
	if err := srv.mysql.Client.Model(TeamUser{}).Where("team_id = ? and user_id = ?", teamId, userId).First(&TeamUser{}).Error; err != nil {
		return false
	}
	return true
}

type TeamUserStatus string
type Role string

const (
	RoleAdmin    Role = "Admin"
	RoleOperator Role = "Operator"
	RoleViewer   Role = "Viewer"
)

const (
	TeamUserStatusPending TeamUserStatus = "Pending"
	TeamUserStatusJoined  TeamUserStatus = "Joined"
	TeamUserStatusReject  TeamUserStatus = "Reject"
	TeamUserStatusExpired TeamUserStatus = "Expired"
)

const (
	InvitationAccept = "Accept"
	InvitationReject = "Reject"
)

func (m *TeamUser) TableName() string {
	return "base_teams_users"
}
