package routes

import (
	"context"
	"encoding/json"
	"net/http"
)

type ServiceHandleFunc func(c IContext)

type HTTPContext struct {
	w http.ResponseWriter
	r *http.Request
}

type IContext interface {
	Query(name string) string
	Param(key string) string

	JSON(code int, obj any)
	Bind(obj any) error
	Get(key string) any
	Set(key string, value any)
	GetSession() string
}

func (c *HTTPContext) Query(name string) string {
	return c.r.URL.Query().Get(name)
}

func (c *HTTPContext) GetSession() string {
	return c.w.Header().Get(XSession)
}

func (c *HTTPContext) Get(key string) any {
	return c.r.Context().Value(ContextKey(key))
}

func (c *HTTPContext) Set(key string, value any) {
	c.r = c.r.WithContext(context.WithValue(c.r.Context(), ContextKey(key), value))
}

func (c *HTTPContext) Param(key string) string {
	v := c.r.Context().Value(ContextKey(key))
	var result string
	switch v := v.(type) {
	case string:
		result = v
	}

	c.r = c.r.WithContext(context.WithValue(c.r.Context(), ContextKey(key), nil))
	return result
}

func (c *HTTPContext) Bind(obj any) error {
	decoder := json.NewDecoder(c.r.Body)
	decoder.UseNumber()
	decoder.DisallowUnknownFields()
	return decoder.Decode(obj)
}

func NewMyContext(w http.ResponseWriter, r *http.Request) IContext {
	return &HTTPContext{w, r}
}

func (c *HTTPContext) JSON(code int, obj any) {
	c.w.Header().Set("Content-Type", "application/json; charset=UTF8")
	c.w.WriteHeader(code)
	json.NewEncoder(c.w).Encode(obj)
}
