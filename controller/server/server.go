package server

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/astaxie/beego/session"
	_ "github.com/astaxie/beego/session/redis"
	"github.com/codegangsta/cli"
	"github.com/go-ini/ini"
	"github.com/nicle-lin/lillian/controller/api"
	"github.com/nicle-lin/lillian/controller/manager"
	"github.com/nicle-lin/lillian/helper/auth/builtin"
	"github.com/nicle-lin/lillian/version"
	"github.com/nicle-lin/mysql"
	"github.com/nicle-lin/redis"
	"os"
	"path/filepath"
)

const configPath = "config/config.ini"

var (
	cfg *ini.File
)

func init() {
	curWorkPath, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	configAbsPath := filepath.Join(curWorkPath, configPath)
	cfg, err = ini.Load(configAbsPath)
	if err != nil {
		log.Fatal(err)
	}
}

func Server(c *cli.Context) {
	disableUsageInfo := c.Bool("disable-usage-info")
	log.Infof("lillian CRM version: %s", version.Version)

	// default to builtin auth
	authenticator := builtin.NewAuthenticator("defaultlillian")

	globalSessions := Session()
	redis := redisSession()
	mysql := mysqlSession()

	controllerManager, err := manager.NewManager(redis,mysql,globalSessions, disableUsageInfo, authenticator)
	if err != nil {
		log.Fatal(err)
	}

	listenAddr := GetKeyValueString("app","host")
	apiConfig := api.ApiConfig{
		ListenAddr: listenAddr,
		Manager:    controllerManager,
	}

	lillianApi := api.NewApi(apiConfig)

	if err := lillianApi.Run(); err != nil {
		log.Fatal(err)
	}
}

func Session() *session.Manager {
	log.Debug("setting up session")

	cookiename := GetKeyValueString("session", "cookiename")
	gclifetime := GetKeyValueInt("session", "gclifetime")
	maxpoolsize := GetKeyValueString("session", "maxpoolsize")
	host := GetKeyValueString("session", "host")
	port := GetKeyValueString("session", "port")
	password := GetKeyValueString("session", "password")

	if host == "" || port == "" || password == "" {
		log.Debug("未配置session")
		return nil
	}

	if cookiename == ""{
		cookiename = "lilliansessionid"
	}
	if gclifetime == 0{
		gclifetime = 3600
	}
	if maxpoolsize == ""{
		maxpoolsize = "100"
	}

	cfg := &session.ManagerConfig{
		CookieName:    cookiename,
		Gclifetime:     int64(gclifetime),
		ProviderConfig: fmt.Sprintf("%s:%s,%s,%s",host,port,maxpoolsize,password),

	}
	globalSessions, err := session.NewManager("redis", cfg)
	if err != nil {
		log.Fatal(err)
	}
	go globalSessions.GC()
	return globalSessions
}

func redisSession() *redis.RedisPool {
	host := GetKeyValueString("redis", "host")
	port := GetKeyValueString("redis", "port")
	password := GetKeyValueString("redis", "password")

	if host == "" || port == "" || password == "" {
		log.Debug("未配置redis")
		return nil
	}
	return redis.NewRedisPool(host, port, password)
}

func mysqlSession() *mysql.Mysql {
	user := GetKeyValueString("mysql", "user")
	password := GetKeyValueString("mysql", "password")
	host := GetKeyValueString("mysql", "host")
	port := GetKeyValueString("mysql", "port")
	dbname := GetKeyValueString("mysql", "dbname")
	charset := GetKeyValueString("mysql", "charset")

	if host == "" || port == "" || user == "" || password == "" || dbname == "" {
		log.Debug("未配置mysql")
		return nil
	}

	if dbname == "" {
		dbname = "utf8"
	}

	mysqlConnStr := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s",
		user, password, host, port, dbname, charset)
	return mysql.NewMysql(mysqlConnStr)
}

func GetKeyValueString(section, key string) string {
	return cfg.Section(section).Key(key).String()
}

func GetKeyValueInt(section, key string) int {
	result, err := cfg.Section(section).Key(key).Int()
	if err != nil {
		log.Debug(err)
		return 0
	}
	return result
}
