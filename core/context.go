package core

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type UserInfoContext interface {
	context.Context
	UserInfo() *userInfo
}

type Context struct {
	Request   *http.Request
	Params    map[string]interface{}
	TraceId   string
	Keys      map[string]interface{}
	LogFields map[string]interface{}
	Error     error
	// UserInfo will be deprecated in future
	UserInfo struct {
		AppId         int
		Uin           string
		SubAccountUin string
	}
	Reporter Reporter
	Action   string

	index       int64
	middlewares []Middleware
	mu          sync.RWMutex
}

// NewContext ...
func NewContext() *Context {
	return &Context{
		middlewares: make([]Middleware, 0, 16),
	}
}

type userInfo struct {
	AppId         int64
	Uin           string
	SubAccountUin string
}

type userInfoContext struct {
	ctx      *Context
	userInfo *userInfo
}

// NewUserInfoContext create userInfoContext with Context and AppId/Uin/SubAccountUin
func NewUserInfoContext(ctx *Context, appId int64, uin, subAccountUin string) *userInfoContext {
	return &userInfoContext{
		ctx,
		&userInfo{
			AppId:         appId,
			Uin:           uin,
			SubAccountUin: subAccountUin,
		},
	}
}

func (ctx *userInfoContext) raw() *Context {
	return ctx.ctx
}

func (ctx *userInfoContext) Deadline() (deadline time.Time, ok bool) {
	return ctx.raw().Deadline()
}

func (ctx *userInfoContext) Done() <-chan struct{} {
	return ctx.raw().Done()
}

func (ctx *userInfoContext) Err() error {
	return ctx.raw().Error
}

func (ctx *userInfoContext) Set(key string, value interface{}) {
	ctx.raw().Set(key, value)
}

func (ctx *userInfoContext) Get(key string) (interface{}, bool) {
	return ctx.raw().Get(key)
}

func (ctx *userInfoContext) Value(key interface{}) interface{} {
	return ctx.raw().Value(key)
}

// Cast will cast context.Context to *core.Context
// if c is *core.Context, return directly
// if c is *userInfoContext, return it's raw()
// else return nil means cast failed
func Cast(c context.Context) *Context {
	if ctx, ok := c.(*Context); ok {
		return ctx
	}

	if ctx, ok := c.(*userInfoContext); ok {
		return ctx.raw()
	}

	return nil
}

// WithUserInfo convert current context to userInfoContext, which implement UserInfoContext
// WithUserInfo load UserInfo from ctx.Params, and panic if LoadUserInfo failed.
func (ctx *Context) WithUserInfo() *userInfoContext {
	if err := ctx.LoadUserInfo(ctx.Params); err != nil {
		// LoadUserInfo failed, create userInfoContext will panic
		panic("can not convert current ctx to UserInfoContext")
	}
	userInfoCtx := &userInfoContext{ctx, &userInfo{
		AppId:         int64(ctx.UserInfo.AppId),
		Uin:           ctx.UserInfo.Uin,
		SubAccountUin: ctx.UserInfo.SubAccountUin,
	}}

	return userInfoCtx
}

// UserInfo return userInfo struct in ctx
func (ctx *userInfoContext) UserInfo() *userInfo {
	return ctx.userInfo
}

func (ctx *Context) Deadline() (deadline time.Time, ok bool) {
	// TODO(loganqian)
	return
}

func (ctx *Context) Done() <-chan struct{} {
	// TODO(loganqian)
	return nil
}

func (ctx *Context) Err() error {
	return ctx.Error
}

// LoadUserInfo 自动识别 AppId Uin SubAccountUin 并加载至 Context.UserInfo
// AppId Uin SubAccountUin 支持 string/int/int64，如果是其他格式将返回错误
func (ctx *Context) LoadUserInfo(data map[string]interface{}) error {

	if rawAppId, ok := data["AppId"]; ok {
		switch appId := rawAppId.(type) {
		case string:
			if intAppId, err := strconv.Atoi(appId); err != nil {
				return err
			} else {
				ctx.UserInfo.AppId = intAppId
			}
		case int:
			ctx.UserInfo.AppId = appId
		case int64:
			ctx.UserInfo.AppId = int(appId)
		case float64:
			ctx.UserInfo.AppId = int(appId)
		case json.Number:
			int64AppId, err := appId.Int64()
			if err != nil {
				return err
			}

			ctx.UserInfo.AppId = int(int64AppId)
		default:
			return errors.New("unknown AppId type")
		}
	} else {
		return errors.New("AppId not found")
	}

	if rawUin, ok := data["Uin"]; ok {
		switch uin := rawUin.(type) {
		case string:
			ctx.UserInfo.Uin = uin
		case int:
			ctx.UserInfo.Uin = strconv.Itoa(uin)
		case int64:
			ctx.UserInfo.Uin = strconv.Itoa(int(uin))
		case json.Number:
			ctx.UserInfo.Uin = uin.String()
		default:
			return errors.New("unknown Uin type")
		}
	} else {
		return errors.New(" Uin not found")
	}

	if rawSubAccountUin, ok := data["SubAccountUin"]; ok {
		switch subAccountUin := rawSubAccountUin.(type) {
		case string:
			ctx.UserInfo.SubAccountUin = subAccountUin
		case int:
			ctx.UserInfo.SubAccountUin = strconv.Itoa(subAccountUin)
		case int64:
			ctx.UserInfo.SubAccountUin = strconv.Itoa(int(subAccountUin))
		case json.Number:
			ctx.UserInfo.SubAccountUin = subAccountUin.String()
		default:
			return errors.New("unknown SubAccountUin type")
		}
	} else {
		return errors.New("SubAccountUin not found")
	}

	return nil
}

func (ctx *Context) Set(key string, value interface{}) {
	ctx.mu.Lock()
	if ctx.Keys == nil {
		ctx.Keys = make(map[string]interface{})
	}

	ctx.Keys[key] = value
	ctx.mu.Unlock()
}

func (ctx *Context) Get(key string) (interface{}, bool) {
	ctx.mu.RLock()
	value, exists := ctx.Keys[key]
	ctx.mu.RUnlock()
	return value, exists
}

func (ctx *Context) Value(key interface{}) interface{} {
	if key, ok := key.(string); ok {
		val, _ := ctx.Get(key)
		return val
	}
	return nil
}

// Use add middleware to request context
func (ctx *Context) Use(mw ...Middleware) {
	ctx.middlewares = append(ctx.middlewares, mw...)
}

// Next call next middleware recursively
func (ctx *Context) Next() error {
	if int(ctx.index) >= len(ctx.middlewares) {
		return nil
	}
	ctx.index++
	return ctx.middlewares[ctx.index-1].Run(ctx)
}
