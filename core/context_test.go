package core

import (
	"github.com/google/uuid"
	"net/http"
	"sync"
	"testing"
)

const (
	AppId         = 123456789
	Uin           = "409339559"
	SubAccountUin = "409339559"
)

func TestContext_LoadUserInfo(t *testing.T) {
	type fields struct {
		Request   *http.Request
		Params    map[string]interface{}
		TraceId   string
		mu        sync.RWMutex
		Keys      map[string]interface{}
		LogFields map[string]interface{}
		Error     error
		UserInfo  struct {
			AppId         int
			Uin           string
			SubAccountUin string
		}
		Reporter Reporter
	}
	type args struct {
		data map[string]interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "test LoadUserInfo",
			args: args{
				data: map[string]interface{}{
					"Uin":           Uin,
					"SubAccountUin": SubAccountUin,
					"AppId":         AppId,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Context{
				Request: tt.fields.Request,
				Params:  tt.fields.Params,
				TraceId: tt.fields.TraceId,
				// mu:        tt.fields.mu,
				Keys:      tt.fields.Keys,
				LogFields: tt.fields.LogFields,
				Error:     tt.fields.Error,
				UserInfo:  tt.fields.UserInfo,
				Reporter:  tt.fields.Reporter,
			}
			if err := c.LoadUserInfo(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("LoadUserInfo() error = %v, wantErr %v", err, tt.wantErr)
			}
			if c.UserInfo.AppId != AppId ||
				c.UserInfo.Uin != Uin ||
				c.UserInfo.SubAccountUin != SubAccountUin {
				t.Error("LoadUserInfo result error")
			}
		})
	}
}

func TestNewUserInfoContext(t *testing.T) {
	traceId := uuid.New().String()
	ctx := &Context{TraceId: traceId}
	userInfoCtx := NewUserInfoContext(ctx, AppId, Uin, SubAccountUin)
	if userInfoCtx.ctx.TraceId != traceId {
		t.Error("traceId not equal")
	}

	userInfo := userInfoCtx.UserInfo()
	if userInfo.AppId != AppId || userInfo.Uin != Uin || userInfo.SubAccountUin != SubAccountUin {
		t.Error("user info not match")
	}
}
