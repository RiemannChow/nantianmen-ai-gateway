package ldap

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/RiemannChow/nantianmen-ai-gateway/module/license"

	"github.com/eolinker/ap-account/auth_driver"
	"github.com/eolinker/ap-account/service/account"
	"github.com/eolinker/ap-account/service/role"
	"github.com/eolinker/ap-account/service/user"

	"github.com/eolinker/go-common/autowire"
	"github.com/eolinker/go-common/utils"

	"github.com/go-ldap/ldap/v3"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	name  = "ldap"
	title = "LDAP / SSO"
)

var _ auth_driver.IDriver = (*Driver)(nil)

func init() {
	d := &Driver{}
	auth_driver.Register(name, d)
}

// Driver implements auth_driver.IDriver for LDAP / SSO. It connects to an
// external LDAP directory, binds with a service account, searches for the
// logging-in user, performs a second simple-bind as that user to verify the
// credentials, then provisions (or updates) a local account. This is an
// enterprise (commercial) capability gated by the "sso" license module.
type Driver struct {
	isInit            bool
	accountService    account.IAccountService `autowired:""`
	userService       user.IUserService       `autowired:""`
	roleService       role.IRoleService       `autowired:""`
	roleMemberService role.IRoleMemberService `autowired:""`
}

func (d *Driver) Init() {
	if d.isInit {
		return
	}
	autowire.Autowired(&d.accountService)
	autowire.Autowired(&d.userService)
	autowire.Autowired(&d.roleService)
	autowire.Autowired(&d.roleMemberService)
	d.isInit = true
}

// FilterConfig masks the service-account password before the config is sent to
// the frontend, so the secret is never leaked in GET responses.
func (d *Driver) FilterConfig(config map[string]string) {
	delete(config, "bind_password")
}

func (d *Driver) Name() string  { return name }
func (d *Driver) Title() string { return title }

func (d *Driver) Delete(ctx context.Context, ids ...string) error {
	return d.accountService.OnRemoveUsers(ctx, ids...)
}

func (d *Driver) ThirdLogin(ctx context.Context, args map[string]string) (string, error) {
	// 企业版能力门禁：未授权 sso 模块时拒绝 LDAP 登录。
	if err := license.Check(ctx, "sso"); err != nil {
		return "", err
	}

	username, ok := args["username"]
	if !ok || username == "" {
		return "", fmt.Errorf("missing username parameter")
	}
	password, ok := args["password"]
	if !ok || password == "" {
		return "", fmt.Errorf("missing password parameter")
	}

	cfg, err := parseConfig(args)
	if err != nil {
		return "", err
	}

	l, err := connect(cfg)
	if err != nil {
		return "", err
	}
	defer l.Close()

	// 1. 使用服务账号绑定（匿名绑定亦允许时 bind_dn 可留空）。
	if cfg.BindDN != "" {
		if err := l.Bind(cfg.BindDN, cfg.BindPassword); err != nil {
			return "", fmt.Errorf("ldap bind service account error: %w", err)
		}
	}

	// 2. 在基准 DN 下搜索用户。
	searchReq := ldap.NewSearchRequest(
		cfg.BaseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		fmt.Sprintf(cfg.UserFilter, ldap.EscapeFilter(username)),
		[]string{cfg.UserDNAttr, cfg.AttrUsername, cfg.AttrName, cfg.AttrEmail, cfg.AttrMobile},
		nil,
	)
	result, err := l.Search(searchReq)
	if err != nil {
		return "", fmt.Errorf("ldap search error: %w", err)
	}
	if len(result.Entries) == 0 {
		return "", fmt.Errorf("ldap user not found")
	}
	entry := result.Entries[0]
	userDN := entry.DN

	// 3. 以该用户 DN 二次绑定，校验密码。
	if err := l.Bind(userDN, password); err != nil {
		return "", fmt.Errorf("ldap bind user error: invalid credentials")
	}

	// 4. 提取属性并落地本地账号。
	uid := entry.GetAttributeValue(cfg.AttrUsername)
	if uid == "" {
		uid = username
	}
	userName := entry.GetAttributeValue(cfg.AttrName)
	if userName == "" {
		userName = username
	}
	email := entry.GetAttributeValue(cfg.AttrEmail)
	mobile := entry.GetAttributeValue(cfg.AttrMobile)

	info, err := d.accountService.GetIdentifier(ctx, name, uid)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return "", err
		}
		newUid := uuid.NewString()
		err = d.accountService.Save(ctx, name, newUid, uid, utils.Md5(fmt.Sprintf("%s%s", newUid, uid)))
		if err != nil {
			return "", err
		}
		_, err = d.userService.Create(ctx, newUid, userName, email, mobile, name)
		if err != nil {
			return "", err
		}
		r, err := d.roleService.GetDefaultRole(ctx, role.SystemTarget())
		if err != nil {
			return "", err
		}
		err = d.roleMemberService.Add(ctx, &role.AddMember{
			Role:   r.Id,
			User:   newUid,
			Target: role.SystemTarget(),
		})
		if err != nil {
			return "", err
		}
		return newUid, nil
	}

	_, err = d.userService.Update(ctx, info.Uid, &userName, &email, &mobile)
	if err != nil {
		return "", err
	}
	return info.Uid, nil
}

// ldapConfig holds the resolved LDAP connection + mapping settings.
type ldapConfig struct {
	Host         string
	BaseDN       string
	BindDN       string
	BindPassword string
	UserFilter   string
	UserDNAttr   string
	AttrUsername string
	AttrName     string
	AttrEmail    string
	AttrMobile   string
	UseSSL       bool
}

func parseConfig(args map[string]string) (*ldapConfig, error) {
	host := args["host"]
	if host == "" {
		return nil, fmt.Errorf("missing host parameter")
	}
	userFilter := args["user_filter"]
	if userFilter == "" {
		userFilter = "(&(objectClass=person)(uid=%s))"
	}
	userDNAttr := args["user_dn_attr"]
	if userDNAttr == "" {
		userDNAttr = "uid"
	}
	return &ldapConfig{
		Host:         host,
		BaseDN:       args["base_dn"],
		BindDN:       args["bind_dn"],
		BindPassword: args["bind_password"],
		UserFilter:   userFilter,
		UserDNAttr:   userDNAttr,
		AttrUsername: defaultIfEmpty(args["attr_username"], "uid"),
		AttrName:     defaultIfEmpty(args["attr_name"], "cn"),
		AttrEmail:    defaultIfEmpty(args["attr_email"], "mail"),
		AttrMobile:   defaultIfEmpty(args["attr_mobile"], "mobile"),
		UseSSL:       strings.EqualFold(args["use_ssl"], "true") || args["use_ssl"] == "1",
	}, nil
}

func connect(cfg *ldapConfig) (*ldap.Conn, error) {
	addr := cfg.Host
	if !strings.Contains(addr, "://") {
		scheme := "ldap"
		if cfg.UseSSL {
			scheme = "ldaps"
		}
		addr = fmt.Sprintf("%s://%s", scheme, addr)
	}
	// 规范化端口缺失场景。
	if u, err := url.Parse(addr); err == nil && u.Port() == "" {
		if cfg.UseSSL {
			addr = fmt.Sprintf("%s://%s:636", u.Scheme, u.Host)
		} else {
			addr = fmt.Sprintf("%s://%s:389", u.Scheme, u.Host)
		}
	}

	l, err := ldap.DialURL(addr)
	if err != nil {
		return nil, fmt.Errorf("ldap dial error: %w", err)
	}

	// 明文 ldap:// 配合 UseSSL 时升级为 StartTLS。
	if cfg.UseSSL && strings.HasPrefix(addr, "ldap://") {
		if err := l.StartTLS(&tls.Config{InsecureSkipVerify: true}); err != nil {
			l.Close()
			return nil, fmt.Errorf("ldap starttls error: %w", err)
		}
	}
	return l, nil
}

func defaultIfEmpty(v, def string) string {
	if v == "" {
		return def
	}
	return v
}
