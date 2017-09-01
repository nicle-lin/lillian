package manager

import (
	"errors"

	log "github.com/Sirupsen/logrus"
	"github.com/astaxie/beego/session"
	"github.com/nicle-lin/lillian/helper/auth"
	"github.com/nicle-lin/lillian/model"
	"github.com/nicle-lin/mysql"
	"github.com/nicle-lin/redis"
	"net/http"
	"time"
)

const (
	tblNameConfig      = "config"
	tblNameEvents      = "events"
	tblNameAccounts    = "accounts"
	tblNameRoles       = "roles"
	tblNameServiceKeys = "service_keys"
	tblNameExtensions  = "extensions"
	storeKey           = "lillian"
	trackerHost        = "http://1001ai.com"
	NodeHealthUp       = "up"
	NodeHealthDown     = "down"
)

var (
	ErrLoginFailure               = errors.New("无效的用户名和密码")
	ErrAccountExists              = errors.New("账户已存在")
	ErrAccountDoesNotExist        = errors.New("账户不存在")
	ErrRoleDoesNotExist           = errors.New("角色不存在")
	ErrNodeDoesNotExist           = errors.New("节点不存在")
	ErrServiceKeyDoesNotExist     = errors.New("服务密钥不存在")
	ErrInvalidAuthToken           = errors.New("无效的认证令牌")
	ErrExtensionDoesNotExist      = errors.New("Extension 不存在")
	ErrWebhookKeyDoesNotExist     = errors.New("webhook key 不存在")
	ErrRegistryDoesNotExist       = errors.New("registry 不存在")
	ErrConsoleSessionDoesNotExist = errors.New("控制台session不存在")
)

type DefaultManager struct {
	authKey          string
	authenticator    auth.Authenticator
	disableUsageInfo bool
	globalSessions   *session.Manager
	mysql            *mysql.Mysql
	redis            *redis.RedisPool
}

type ScaleResult struct {
	Scaled []string
	Errors []string
}

type Manager interface {
	Store(w http.ResponseWriter, r *http.Request) session.Store
	Redis() *redis.RedisPool
	Mysql() *mysql.Mysql
	Accounts() ([]*auth.Account, error)
	Account(username string) (*auth.Account, error)
	Authenticate(username, password string) (bool, error)
	GetAuthenticator() auth.Authenticator
	SaveAccount(account *auth.Account) error
	DeleteAccount(account *auth.Account) error
	NewAuthToken(username string, userAgent string) (*auth.AuthToken, error)
	VerifyServiceKey(key string) error
	VerifyAuthToken(username, token string) error
	ChangePassword(username, password string) error
	SaveEvent(event *model.Event) error
	Events(limit int) ([]*model.Event, error)
	PurgeEvents() error
	LogEvent(eventType, message string, tags []string)
}

func NewManager(redis *redis.RedisPool, mysql *mysql.Mysql, globalSessions *session.Manager, disableUsageInfo bool,
	authenticator auth.Authenticator) (Manager, error) {

	m := &DefaultManager{
		authenticator:    authenticator,
		disableUsageInfo: disableUsageInfo,
		globalSessions:   globalSessions,
		redis:            redis,
		mysql:            mysql,
	}
	return m, nil
}

func (m DefaultManager) Store(w http.ResponseWriter, r *http.Request) session.Store {
	store, err := m.globalSessions.SessionStart(w, r)
	if err != nil {
		log.Debug(err)
		return nil
	}
	return store
}

func (m DefaultManager) Redis() *redis.RedisPool {
	return m.redis
}

func (m DefaultManager) Mysql() *mysql.Mysql {
	return m.mysql
}

func (m DefaultManager) Accounts() ([]*auth.Account, error) {
	return nil, nil
}

func (m DefaultManager) Account(username string) (*auth.Account, error) {
	return nil, nil
}

func (m DefaultManager) SaveAccount(account *auth.Account) error {
	return nil
}

func (m DefaultManager) DeleteAccount(account *auth.Account) error {
	return nil
}

func (m DefaultManager) Roles() ([]*auth.ACL, error) {
	roles := auth.DefaultACLs()
	return roles, nil
}

func (m DefaultManager) Role(name string) (*auth.ACL, error) {
	acls, err := m.Roles()
	if err != nil {
		return nil, err
	}

	for _, r := range acls {
		if r.RoleName == name {
			return r, nil
		}
	}

	return nil, nil
}

func (m DefaultManager) GetAuthenticator() auth.Authenticator {
	return m.authenticator
}

func (m DefaultManager) Authenticate(username, password string) (bool, error) {
	// only get the account to get the hashed password if using the builtin auth
	passwordHash := ""
	if m.authenticator.Name() == "builtin" {
		acct, err := m.Account(username)
		if err != nil {
			log.Error(err)
			return false, ErrLoginFailure
		}

		passwordHash = acct.Password
	}

	a, err := m.authenticator.Authenticate(username, password, passwordHash)
	if !a || err != nil {
		log.Error(ErrLoginFailure)
		return false, ErrLoginFailure
	}

	return true, nil
}

func (m DefaultManager) NewAuthToken(username string, userAgent string) (*auth.AuthToken, error) {
	return nil, nil
}

func (m DefaultManager) VerifyAuthToken(username, token string) error {
	return nil
}

func (m DefaultManager) VerifyServiceKey(key string) error {
	return nil
}

func (m DefaultManager) NewServiceKey(description string) (*auth.ServiceKey, error) {
	return nil, nil
}

func (m DefaultManager) ChangePassword(username, password string) error {
	return nil
}

func (m DefaultManager) SaveEvent(event *model.Event) error {

	return nil
}

func (m DefaultManager) Events(limit int) ([]*model.Event, error) {
	return nil, nil
}

func (m DefaultManager) PurgeEvents() error {
	return nil
}

func (m DefaultManager) LogEvent(eventType, message string, tags []string) {
	evt := &model.Event{
		Type:    eventType,
		Time:    time.Now(),
		Message: message,
		Tags:    tags,
	}
	if err := m.SaveEvent(evt); err != nil {
		log.Errorf("logging event error:%s\n", err)
	}
}
