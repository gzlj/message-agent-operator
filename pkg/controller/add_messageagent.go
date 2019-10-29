package controller

import (
	"github.com/gzlj/message-agent-operator/pkg/controller/messageagent"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, messageagent.Add)
}
