package jobs

import (
	"net/http"

	"github.com/beego/beego/context"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/pt_af_logic/types"
)

const systemUserName = "system"

func getSystemUser() (*object.User, error) {
	return &object.User{
		Owner:         types.BuiltInOrgCode,
		Name:          systemUserName,
		IsGlobalAdmin: true,
	}, nil
}

func createCtx() *context.Context {
	ctx := context.NewContext()
	ctx.Request = &http.Request{
		Header: make(http.Header),
	}
	return ctx
}
