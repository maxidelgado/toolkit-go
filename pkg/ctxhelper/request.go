package ctxhelper

import (
	"context"
)

const Key = "RequestMetadata"

type ContextHelper interface {
	GetRequestId() string
	SetRequestId(string)
	GetHeader(string) interface{}
	SetHeader(string, interface{})
	GetUser() User
	SetUser(User)
	GetApiKey() ApiKey
	SetApiKey(ApiKey)
}

type User struct {
	Id      string
	GroupId string
}

type ApiKey struct {
	Id       string
	Platform string
}

type RequestHelper struct {
	requestId string
	apiKey    ApiKey
	user      User
	headers   map[string]interface{}
}

func New() ContextHelper {
	rh := RequestHelper{}

	return &rh
}

func WithContext(ctx context.Context) ContextHelper {
	if rh, ok := ctx.Value(Key).(ContextHelper); ok {
		return rh
	}
	return New()
}

func (r *RequestHelper) SetRequestId(rid string) {
	r.requestId = rid
}

func (r *RequestHelper) GetRequestId() string {
	return r.requestId
}

func (r *RequestHelper) GetHeader(key string) interface{} {
	return r.headers[key]
}

func (r *RequestHelper) SetHeader(key string, value interface{}) {
	r.headers[key] = value
}

func (r *RequestHelper) GetUser() User {
	return r.user
}

func (r *RequestHelper) SetUser(user User) {
	r.user = user
}

func (r *RequestHelper) GetApiKey() ApiKey {
	return r.apiKey
}

func (r *RequestHelper) SetApiKey(apiKey ApiKey) {
	r.apiKey = apiKey
}
