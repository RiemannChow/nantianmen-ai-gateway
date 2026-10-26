package ai_provider_local

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-contrib/gzip"

	"github.com/eolinker/eosc/log"

	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
)

// newTestRouter 组装本地模型管理的调试路由（与包内 HTTP 处理器一致）。
func newTestRouter() *gin.Engine {
	r := gin.New()
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.GET("/stream", streamHandler)
	r.GET("/stop", stopPull)
	r.GET("/models", models)
	return r
}

// TestStreamHandlerPullFailure 验证模型拉取失败时 /stream 快速返回 500，
// 而不是挂起（回归防护：此前的调试用 r.Run 会永久阻塞，导致 go test 超时）。
func TestStreamHandlerPullFailure(t *testing.T) {
	srv := httptest.NewServer(newTestRouter())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/stream?model=definitely-not-a-real-model-xyz")
	if err != nil {
		t.Fatalf("GET /stream: %v", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("GET /stream with invalid model: status = %d, want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

// TestStopHandler 验证 /stop 对未在拉取的模型也能正常响应。
func TestStopHandler(t *testing.T) {
	srv := httptest.NewServer(newTestRouter())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/stop?model=definitely-not-a-real-model-xyz")
	if err != nil {
		t.Fatalf("GET /stop: %v", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET /stop: status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func streamHandler(c *gin.Context) {
	// 创建一个通道，用于监测客户端关闭连接的信号
	model := c.Query("model")
	p, err := PullModel(model, uuid.NewString(), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	done := make(chan struct{})
	// 启动一个 goroutine 监听客户端关闭连接
	go func() {
		select {
		case <-c.Writer.CloseNotify():
			log.Info("client closed connection,close pipeline")
			taskExecutor.ClosePipeline(model, p.id)
		case <-done:
		}
	}()

	c.Stream(func(w io.Writer) bool {
		msg, ok := <-p.channel
		if !ok {
			return false
		}
		_, err := w.Write([]byte(fmt.Sprintf("%s\n", msg.Msg)))
		if err != nil {
			log.Error("write message error: %v", err)
			return false
		}
		return true
	})
	done <- struct{}{}
}

func stopPull(c *gin.Context) {
	model := c.Query("model")
	StopPull(model)
	c.JSON(http.StatusOK, gin.H{"message": "stop pull model"})
}

func models(c *gin.Context) {
	ms, err := ModelsInstalled()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"models": ms})
}
