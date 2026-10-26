package common

import "testing"

func TestIsMatchString(t *testing.T) {
	cases := []struct {
		pattern RegexpPattern
		value   string
		wantErr bool
	}{
		{EnglishOrNumber_, "abc_123", false},
		{EnglishOrNumber_, "1abc", true},
		{EnglishOrNumber_, "abc-123", true},
		{AnyEnglishOrNumber_, "abc_123", false},
		{AnyEnglishOrNumber_, "abc-123", true},
		{AnyEnglishOrNumber_, "", true},
		{UUIDExp, "abc-123_x", false},
		{UUIDExp, "abc.123", true},
	}
	for _, c := range cases {
		err := IsMatchString(c.pattern, c.value)
		if (err != nil) != c.wantErr {
			t.Errorf("IsMatchString(%q, %q) error = %v, wantErr %v", c.pattern, c.value, err, c.wantErr)
		}
	}
}

func TestVariable(t *testing.T) {
	if !IsMatchVariable("${abc}") {
		t.Error("IsMatchVariable(${abc}) = false, want true")
	}
	if IsMatchVariable("abc") {
		t.Error("IsMatchVariable(abc) = true, want false")
	}
	if IsMatchVariable("${1abc}") {
		t.Error("IsMatchVariable(${1abc}) = true, want false")
	}
	if got := GetVariableKey("${my_var}"); got != "my_var" {
		t.Errorf("GetVariableKey = %q, want %q", got, "my_var")
	}
}

func TestFilterAppKey(t *testing.T) {
	if !IsMatchFilterAppKey("appkey{myapp}") {
		t.Error("IsMatchFilterAppKey(appkey{myapp}) = false, want true")
	}
	if IsMatchFilterAppKey("appkey{}") {
		t.Error("IsMatchFilterAppKey(appkey{}) = true, want false")
	}
	if got := GetFilterAppKey("appkey{myapp}"); got != "myapp" {
		t.Errorf("GetFilterAppKey = %q, want %q", got, "myapp")
	}
	if got := SetFilterAppKey("myapp"); got != "appkey{myapp}" {
		t.Errorf("SetFilterAppKey = %q, want %q", got, "appkey{myapp}")
	}
}

func TestIsMatchDomainPort(t *testing.T) {
	cases := []struct {
		value string
		want  bool
	}{
		{"example.com", true},
		{"example.com:8080", true},
		{"sub.example.com", true},
		// 单段主机名（如 localhost / 内网服务名）按当前正则同样合法
		{"example", true},
		{"http://example.com", false},
		{"-example.com", false},
	}
	for _, c := range cases {
		if got := IsMatchDomainPort(c.value); got != c.want {
			t.Errorf("IsMatchDomainPort(%q) = %v, want %v", c.value, got, c.want)
		}
	}
}

func TestIsMatchIpPort(t *testing.T) {
	cases := []struct {
		value string
		want  bool
	}{
		{"127.0.0.1:8080", true},
		{"192.168.1.100:65535", true},
		{"256.1.1.1:80", false},
		{"127.0.0.1", false},
	}
	for _, c := range cases {
		if got := IsMatchIpPort(c.value); got != c.want {
			t.Errorf("IsMatchIpPort(%q) = %v, want %v", c.value, got, c.want)
		}
	}
}

func TestIsMatchSchemeIpPort(t *testing.T) {
	cases := []struct {
		value string
		want  bool
	}{
		{"http://127.0.0.1:8080", true},
		{"https://10.0.0.1:443", true},
		{"http://127.0.0.1", false},
		{"127.0.0.1:8080", false},
	}
	for _, c := range cases {
		if got := IsMatchSchemeIpPort(c.value); got != c.want {
			t.Errorf("IsMatchSchemeIpPort(%q) = %v, want %v", c.value, got, c.want)
		}
	}
}

func TestIsMatchCIDRIpv4(t *testing.T) {
	cases := []struct {
		value string
		want  bool
	}{
		{"192.168.1.0", true},
		{"192.168.1.0/24", true},
		{"10.0.0.0/32", true},
		{"10.0.0.0/33", false},
		{"256.0.0.1", false},
		{"192.168.1.0/", false},
	}
	for _, c := range cases {
		if got := IsMatchCIDRIpv4(c.value); got != c.want {
			t.Errorf("IsMatchCIDRIpv4(%q) = %v, want %v", c.value, got, c.want)
		}
	}
}

func TestRestfulPath(t *testing.T) {
	if !IsRestfulPath("/api/{id}/detail") {
		t.Error("IsRestfulPath(/api/{id}/detail) = false, want true")
	}
	if IsRestfulPath("/api/detail") {
		t.Error("IsRestfulPath(/api/detail) = true, want false")
	}
	if !IsRestfulParam("{id}") {
		t.Error("IsRestfulParam({id}) = false, want true")
	}
	if IsRestfulParam("id") {
		t.Error("IsRestfulParam(id) = true, want false")
	}
	if got := ReplaceRestfulPath("/api/{id}/x/{name}", "*"); got != "/api/*/x/*" {
		t.Errorf("ReplaceRestfulPath = %q, want %q", got, "/api/*/x/*")
	}
}

func TestCheckPathContainsIPPort(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{"http://127.0.0.1:8080/api", true},
		{"127.0.0.1:8080/api", true},
		{"/api/v1/users", false},
	}
	for _, c := range cases {
		if got := CheckPathContainsIPPort(c.path); got != c.want {
			t.Errorf("CheckPathContainsIPPort(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}

func TestModelNameValid(t *testing.T) {
	cases := []struct {
		name string
		want bool
	}{
		{"gpt-4", true},
		{"gpt-4o-mini", true},
		{"qwen2.5:72b", true},
		// 模型 ID 允许 / 与 :（如 openrouter 风格的 vendor/model）
		{"openai/gpt-4", true},
		{"gpt 4", false},
		{"gpt@4", false},
	}
	for _, c := range cases {
		if got := ModelNameValid(c.name); got != c.want {
			t.Errorf("ModelNameValid(%q) = %v, want %v", c.name, got, c.want)
		}
	}
}
