package health

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Pinger — минимальный интерфейс для проверки зависимости.
type Pinger interface {
	Ping(ctx context.Context) error
}

type Handler struct {
	dependencies map[string]Pinger
}

func NewHandler(deps map[string]Pinger) *Handler {
	return &Handler{dependencies: deps}
}

// Live — liveness probe. Только «процесс жив», без проверки зависимостей.
func (h *Handler) Live(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Ready — readiness probe. Проверяет все зависимости параллельно.
func (h *Handler) Ready(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	var (
		mu      sync.Mutex
		wg      sync.WaitGroup
		results = make(map[string]string, len(h.dependencies))
		ready   = true
	)

	for name, dep := range h.dependencies {
		name, dep := name, dep
		wg.Add(1)
		go func() {
			defer wg.Done()

			err := dep.Ping(ctx)

			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				results[name] = "unreachable"
				ready = false
				return
			}
			results[name] = "ok"
		}()
	}

	wg.Wait()

	if !ready {
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
			"status": "not ready",
			"checks": results,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
		"checks": results,
	})
}
