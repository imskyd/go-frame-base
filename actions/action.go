package actions

import (
	"encoding/json"
	"github.com/CoinSummer/go-notify"
	"github.com/imroc/req"
	"github.com/imskyd/go-frame-base/types"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

func SetActionLogger(_logger *logrus.Logger) {
	logger = _logger
}

var smtpConfig *types.SmtpConfig

func SetSmtpConfig(_smtpConfig *types.SmtpConfig) {
	smtpConfig = _smtpConfig
}

type IAction interface {
	Type() ActionType
	Run() (string, error)
}

type Actions []*Action

func (a Actions) Init() {
	for _, act := range a {
		act.Init()
	}
}

type Action struct {
	Type         ActionType  `json:"type"`
	Param        interface{} `json:"params"`
	OriginalData interface{}
	ins          IAction
}

func (a *Action) Init() {
	a.ins = NewActionIns(a)
}

func (a *Action) Run() error {
	_, err := a.ins.Run()
	if err != nil {
		return err
	}
	return nil
}

func NewActionIns(action *Action) IAction {
	paramBytes, err := json.Marshal(action.Param)
	if err != nil {
		logger.Errorf("pusher param error")
		return nil
	}
	switch action.Type {
	case NotifyActionType:
		param := new(NotifyActionParam)
		err = json.Unmarshal(paramBytes, param)
		if err != nil {
			logger.Errorf("unmarshal to notify action param error: " + err.Error())
			return nil
		}
		p := notify.Platform(param.Platform)
		if p == "" {
			logger.Errorf("NewActionIns error: platform is trans to notify.Platform error, original platform data is: %s", param.Platform)
			return nil
		}
		config := &notify.Config{
			Platform: p,
			Token:    param.Token,
			Channel:  param.Channel,
			Source:   param.Source,
			Severity: param.Severity,
		}
		if p == notify.PlatformEmail {
			config.User = smtpConfig.User
			config.Password = smtpConfig.Password
			config.Host = smtpConfig.Host + ":" + smtpConfig.Port
		}
		_notifier := notify.NewNotify(config)
		return &NotifyAction{
			config:   config,
			msg:      param.Msg,
			notifier: _notifier,
		}
	case WebhookActionType:
		param := new(WebhookActionParam)
		err = json.Unmarshal(paramBytes, param)
		if err != nil {
			logger.Errorf("notify pusher param error")
			return nil
		}
		return &WebhookAction{
			url:  param.Url,
			data: param.Data,
		}
	default:
		logger.Errorf("error type:%s", action.Type)
		return nil
	}
}

func (f ActionType) IsValid() bool {
	return string(f) != ""
}

type ActionResult struct {
	Action ActionType `json:"pusher"`
	Status bool       `json:"status"`
	Remark string     `json:"remark"`
}

type Notifier interface {
	Send(msg string) error
}

type NotifyAction struct {
	msg      string
	notifier Notifier
	config   *notify.Config
}

func (n *NotifyAction) Run() (string, error) {
	defer func() {
		if err := recover(); err != nil {
			logger.Errorf("NotifyAction panic: %v", err)
		}
	}()
	err := n.notifier.Send(n.msg)
	return "", err
}

func (n *NotifyAction) Type() ActionType {
	return NotifyActionType
}

type WebhookAction struct {
	url  string
	data interface{}
}

func (n *WebhookAction) Type() ActionType {
	return EmailActionType
}

func (n *WebhookAction) Run() (string, error) {
	defer func() {
		if err := recover(); err != nil {
			logger.Errorf("WebhookAction panic")
		}
	}()
	resp, err := req.Post(n.url, req.BodyJSON(n.data))
	if err != nil {
		logger.Warnf("run webhook failed, %s", err.Error())
		return "", err
	}
	if resp.Response().StatusCode != 200 {
		logger.Warnf("run webhook status code!=200, %d", resp.Response().StatusCode)
		return "", ErrStatusCode
	}
	return "", nil
}
