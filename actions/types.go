package actions

type ActionType string

const (
	NotifyActionType  ActionType = "Notify"
	WebhookActionType ActionType = "Webhook"
	EmailActionType   ActionType = "Email"
)

type NotifyActionParam struct {
	Msg      string `json:"msg"`
	GroupId  int    `json:"group_id"`
	Platform string `json:"platform"`
	Token    string `json:"token"`
	Channel  string `json:"channel"`
	Source   string `json:"source"`
	Severity string `json:"severity"`
}

type WebhookActionParam struct {
	Url  string      `json:"url"`
	Data interface{} `json:"data"`
}
