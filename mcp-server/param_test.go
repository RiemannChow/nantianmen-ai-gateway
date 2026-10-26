package mcp_server

import "testing"

func TestBodyParamEncodeJSON(t *testing.T) {
	p := NewBodyParam(string(ContentTypeJSON))
	p.Set("name", "world")
	p.Set("count", 2)

	got, err := p.Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	// JSON 对象键顺序不固定，验证字段存在即可
	for _, want := range []string{`"name":"world"`, `"count":2`} {
		if !contains(got, want) {
			t.Errorf("Encode() = %q, want to contain %q", got, want)
		}
	}
}

func TestBodyParamEncodeForm(t *testing.T) {
	p := NewBodyParam(string(ContentTypeForm))
	p.Set("a", "1")
	p.Set("b", "x y")

	got, err := p.Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	if got != "a=1&b=x+y" {
		t.Errorf("Encode() = %q, want %q", got, "a=1&b=x+y")
	}
}

func TestBodyParamEncodeUnsupported(t *testing.T) {
	p := NewBodyParam("application/octet-stream")
	if _, err := p.Encode(); err == nil {
		t.Error("Encode with unsupported content type should fail")
	}
}

func TestNewBodyParamDefaultsToJSON(t *testing.T) {
	p := NewBodyParam("")
	p.Set("k", "v")
	got, err := p.Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	if got != `{"k":"v"}` {
		t.Errorf("Encode() = %q, want %q", got, `{"k":"v"}`)
	}
}

func TestNewParam(t *testing.T) {
	p := NewParam(PositionHeader, true, "api key")
	if p.Position() != PositionHeader {
		t.Errorf("Position() = %q, want %q", p.Position(), PositionHeader)
	}
	if !p.Required() {
		t.Error("Required() = false, want true")
	}
	if p.Description() != "api key" {
		t.Errorf("Description() = %q, want %q", p.Description(), "api key")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
