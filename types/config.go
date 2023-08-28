package types

type JwtConfig struct {
	PublicKey  string `mapstructure:"PublicKey"`
	PrivateKey string `mapstructure:"PrivateKey"`
}

type SmtpConfig struct {
	User              string `mapstructure:"user"`
	Password          string `mapstructure:"password"`
	Host              string `mapstructure:"host"`
	Port              string `mapstructure:"port"`
	ReplyEmailAddress string `mapstructure:"reply_email_address"`
}
