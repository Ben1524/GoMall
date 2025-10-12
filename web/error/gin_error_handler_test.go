package error

import (
	"bytes"
	"encoding/json"
	"errors"
	"go-notify/model"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

type testStruct struct {
	Name string `json:"name" validate:"required,min=3,max=10"`
	Age  int    `json:"age" validate:"min=1,max=100"`
}

func setupRouterWithHandler() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(GinErrorHandler())
	return r
}

func TestGinErrorHandler_BindValidationError(t *testing.T) {
	r := setupRouterWithHandler()
	r.POST("/test", func(c *gin.Context) {
		var req testStruct
		if err := c.ShouldBindJSON(&req); err != nil {
			// Attach validator.ValidationErrors to gin context
			var ve validator.ValidationErrors
			if errors.As(err, &ve) {
				c.Error(ve).SetType(gin.ErrorTypeBind)
			} else {
				c.Error(err).SetType(gin.ErrorTypeBind)
			}
			c.Status(http.StatusBadRequest)
			return
		}
		c.Status(http.StatusOK)
	})

	// Missing required field "name"
	body := `{"age":0}`
	req, _ := http.NewRequest("POST", "/test", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp model.Error
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Contains(t, resp.ErrorDescription, "Field 'name' is required")
	assert.Contains(t, resp.ErrorDescription, "Field 'age' must be more or equal to 1")
}

func TestGinErrorHandler_BindNonValidationError(t *testing.T) {
	r := setupRouterWithHandler()
	r.GET("/test", func(c *gin.Context) {
		c.Error(errors.New("some bind error")).SetType(gin.ErrorTypeBind)
		c.Status(http.StatusBadRequest)
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp model.Error
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "some bind error", resp.ErrorDescription)
}

func TestGinErrorHandler_DefaultError(t *testing.T) {
	r := setupRouterWithHandler()
	r.GET("/test", func(c *gin.Context) {
		c.Error(errors.New("some other error")).SetType(gin.ErrorTypePrivate)
		c.Status(http.StatusInternalServerError)
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp model.Error
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "some other error", resp.ErrorDescription)
}

func TestGinErrorHandler_NoError(t *testing.T) {
	r := setupRouterWithHandler()
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Empty(t, w.Body.Bytes())
}
