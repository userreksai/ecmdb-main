package ginx

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestWrapDoesNotAppendJSONAfterDirectWrite(t *testing.T) {
	gin.SetMode(gin.TestMode)
	server := gin.New()
	server.GET("/file", Wrap(func(ctx *gin.Context) (Result, error) {
		ctx.Data(http.StatusOK, "application/octet-stream", []byte("binary-file"))
		return Result{}, nil
	}))

	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/file", nil))

	assert.Equal(t, "binary-file", recorder.Body.String())
}

func TestWrapBodyDoesNotAppendJSONAfterDirectWrite(t *testing.T) {
	type request struct {
		Name string `json:"name"`
	}

	gin.SetMode(gin.TestMode)
	server := gin.New()
	server.POST("/file", WrapBody[request](func(ctx *gin.Context, _ request) (Result, error) {
		ctx.Data(http.StatusOK, "application/octet-stream", []byte("binary-file"))
		return Result{}, nil
	}))

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/file", strings.NewReader(`{"name":"test"}`))
	req.Header.Set("Content-Type", "application/json")
	server.ServeHTTP(recorder, req)

	assert.Equal(t, "binary-file", recorder.Body.String())
}
