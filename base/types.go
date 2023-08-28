package base

import "time"

type JwtConfig struct {
	PublicKey  string `mapstructure:"PublicKey"`
	PrivateKey string `mapstructure:"PrivateKey"`
}

type OauthConfig struct {
	User            string        `mapstructure:"user"`
	Pass            string        `mapstructure:"pass"`
	Host            string        `mapstructure:"host"`
	Db              string        `mapstructure:"db"`
	MaxIdleConn     int           `mapstructure:"max_idle_conn"`
	MaxOpenConn     int           `mapstructure:"max_open_conn"`
	MaxLifeTimeConn time.Duration `mapstructure:"max_lifetime_conn"`
	MaxIdleTimeConn time.Duration `mapstructure:"max_idletime_conn"`
}

type ApiKey struct {
	Name    string `mapstructure:"name"`
	ChainId int    `mapstructure:"chain_id"`
	Host    string `mapstructure:"host"`
	Key     string `mapstructure:"key"`
	Tps     int    `mapstructure:"tps"`
	Rpc     string `mapstructure:"rpc"`
}

type SmtpConfig struct {
	User              string `mapstructure:"user"`
	Password          string `mapstructure:"password"`
	Host              string `mapstructure:"host"`
	Port              string `mapstructure:"port"`
	ReplyEmailAddress string `mapstructure:"reply_email_address"`
}

type TelegramConfig struct {
	BotName  string `mapstructure:"bot_name"`
	BotToken string `mapstructure:"bot_token"`
}
