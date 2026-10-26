package loki

import (
	"net"
	"testing"
	"time"
)

// lokiReachable reports whether a Loki instance is available at the given
// address. The driver's methods perform live HTTP calls, so without a running
// Loki this test would fail in CI / offline environments.
func lokiReachable(addr string) bool {
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func TestLoki(t *testing.T) {
	if !lokiReachable("localhost:3100") {
		t.Skip("loki not reachable at localhost:3100, skipping integration test")
	}
	d, _, err := NewDriver(`{"url":"http://localhost:3100","header":{"Content-Type":"application/json","X-Scope-OrgID":"tenant1"}}`)
	if err != nil {
		t.Fatalf("failed to create driver: %v", err)
	}
	logCountResult, err := d.LogCount("apinto", nil, 720, "strategy")
	if err != nil {
		t.Fatalf("failed to get log count: %v", err)
	}
	t.Log(logCountResult)
	logs, count, err := d.Logs("apinto", map[string]string{"strategy": "03899736-5d79-4f26-bd6a-c312a5880780"}, time.Now().Add(-time.Hour*24), time.Now(), 1, 1)
	if err != nil {
		t.Fatalf("failed to get logs: %v", err)
	}
	t.Log(logs, count)
	info, err := d.LogInfo("apinto", "c9f6b19c-7dfe-496b-9b39-4d049232fe95")
	if err != nil {
		t.Fatalf("failed to get log info: %v", err)
	}
	t.Log(info)
}

//func TestLokiLog(t *testing.T) {
//
//	headers := make(map[string]string)
//	headers["Content-Type"] = "application/json"
//	headers["X-Scope-OrgID"] = "tenant1"
//	queries := url.Values{}
//	queries.SetMCPServer("query", "{cluster=\"apinto\"} | json | request_id = `c9f6b19c-7dfe-496b-9b39-4d049232fe95`")
//	now := time.Now()
//	start := now.Add(-time.Hour * 24 * 30)
//	queries.SetMCPServer("start", strconv.FormatInt(start.UnixNano(), 10))
//	queries.SetMCPServer("end", strconv.FormatInt(now.UnixNano(), 10))
//	queries.SetMCPServer("limit", "100")
//	a := time.Now()
//	result, err := send[LogInfo](http.MethodGet, "http://localhost:3100/loki/api/v1/query_range", headers, queries, "")
//	if err != nil {
//		t.Fatalf("failed to send request: %v", err)
//	}
//	t.LogItem(time.Now().Sub(a))
//	data, err := json.Marshal(result)
//	if err != nil {
//		t.Fatalf("failed to marshal data: %v", err)
//	}
//	t.LogItem(string(data))
//}
//
//func TestLokiLogCount(t *testing.T) {
//	headers := make(map[string]string)
//	headers["Content-Type"] = "application/json"
//	headers["X-Scope-OrgID"] = "tenant1"
//	queries := url.Values{}
//	//queries.SetMCPServer("query", "sum(count_over_time({cluster=\"apinto\"}[24h])) by (strategy)")
//	queries.SetMCPServer("query", "sum(count_over_time({cluster=\"apinto\"}[24h]))")
//	result, err := send[LogCount](http.MethodGet, "http://localhost:3100/loki/api/v1/query", headers, queries, "")
//	if err != nil {
//		t.Fatalf("failed to send request: %v", err)
//	}
//	data, err := json.Marshal(result)
//	if err != nil {
//		t.Fatalf("failed to marshal data: %v", err)
//	}
//	t.LogItem(string(data))
//}
//
//func TestLokiLogs(t *testing.T) {
//	headers := make(map[string]string)
//	headers["Content-Type"] = "application/json"
//	headers["X-Scope-OrgID"] = "tenant1"
//	queries := url.Values{}
//	queries.SetMCPServer("query", "{cluster=\"apinto\"} | json | strategy=\"03899736-5d79-4f26-bd6a-c312a5880780\"")
//	now := time.Now()
//	start := now.Add(-time.Hour * 24 * 30)
//	queries.SetMCPServer("start", strconv.FormatInt(start.UnixNano(), 10))
//	queries.SetMCPServer("end", strconv.FormatInt(now.UnixNano(), 10))
//	queries.SetMCPServer("limit", "1")
//	now = time.Now()
//	result, err := send[map[string]interface{}](http.MethodGet, "http://localhost:3100/loki/api/v1/query_range", headers, queries, "")
//	t.LogItem(time.Now().Sub(now))
//	if err != nil {
//		t.Fatalf("failed to send request: %v", err)
//	}
//	data, err := json.Marshal(result)
//	if err != nil {
//		t.Fatalf("failed to marshal data: %v", err)
//	}
//	t.LogItem(string(data))
//}
