package actions

import "fmt"

var (
	ErrStatusCode  = fmt.Errorf("status code not 200")
	ErrUnSupported = fmt.Errorf("unSupported event type")
)
