package health

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type fakePinger struct {
	err error
}

func (f fakePinger) Ping(ctx context.Context) error {
	return f.err
}

func TestLive(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHandler(map[string]Pinger{})
	router := gin.New()
	router.GET("/health", h.Live)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestReady_AllHealthy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHandler(map[string]Pinger{
		"database":   fakePinger{err: nil},
		"ml_service": fakePinger{err: nil},
	})
	router := gin.New()
	router.GET("/ready", h.Ready)

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestReady_OneUnhealthy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHandler(map[string]Pinger{
		"database":   fakePinger{err: nil},
		"ml_service": fakePinger{err: errors.New("down")},
	})
	router := gin.New()
	router.GET("/ready", h.Ready)

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want %d", w.Code, http.StatusServiceUnavailable)
	}
}
