package ginext

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	Key_SITE_ADMIN     = "SITE_ADMIN"
	Key_SITE_NAME      = "SITE_NAME"
	Key_SITE_LINK      = "SITE_LINK"
	Key_NOTIFY_BOT_URL = "NOTIFY_BOT_URL"
)

const (
	DBField     = "ginext_db"
	ConfigField = "ginext_cfg"
)

// GinExt Config and core info
type GinExt struct {
	AppDir   string `json:"-"`
	AssetDir string `json:"-"`
	ConfDir  string `json:"-"`
	ConfFile string `json:"-"`
	LogFile  string `json:"log_file"`

	PasswordSalt  string `json:"password_salt"`
	SessionSecret string `json:"session_secret"`
	SessionStore  string `json:"session_store"`
	SessionName   string `json:"session_name"`

	DbDriver  string `json:"db_driver"`
	DbDSN     string `json:"db_dsn"`
	ServeAddr string `json:"serve_addr"`
	RedisAddr string `json:"redis_addr"`

	DbInstance   *gorm.DB       `json:"-"`
	sessionStore sessions.Store `json:"-"`
	LogWriter    io.Writer      `json:"-"`
}

func HintRootDir(conf string) string {
	cwd, _ := os.Getwd()
	dirs := []string{cwd, "~", "..", "../.."}
	for _, dir := range dirs {
		testFileName := filepath.Join(os.ExpandEnv(dir), conf)
		st, err := os.Stat(testFileName)

		if err == nil && !st.IsDir() {
			val, _ := filepath.Abs(os.ExpandEnv(dir))
			return val
		}
	}
	return cwd
}

func NewGinExt(appDir string) *GinExt {
	if len(appDir) <= 0 {
		appDir, _ = os.Getwd()
	}
	log.Println("Rootdir:", appDir)

	cfg := &GinExt{
		AppDir:        appDir,
		AssetDir:      filepath.Join(appDir, "assets"),
		ConfDir:       filepath.Join(appDir, "conf"),
		ConfFile:      filepath.Join(appDir, "conf/settings.json"),
		LogFile:       "",
		PasswordSalt:  "",
		SessionSecret: "ginext-session-secret",
		SessionStore:  "cookie",
		SessionName:   "ginsession",
		DbDriver:      "sqlite",
		DbDSN:         "file::memory:",
		ServeAddr:     ":8080",
		RedisAddr:     "127.0.0.1:6379",
		LogWriter:     os.Stdout,
	}

	withMysql := os.Getenv("WITH_MYSQL")
	if len(withMysql) > 0 {
		/*
			docker run -ti --rm -e MYSQL_ALLOW_EMPTY_PASSWORD=1 -e MYSQL_DATABASE=testdb -p 12306:3306 mysql:8.0.23
			export WITH_MYSQL="root@tcp(127.0.0.1:12306)/testdb?charset=utf8mb4&parseTime=True&loc=Local"
		*/
		cfg.DbDSN = withMysql
		cfg.DbDriver = "mysql"
	}
	return cfg
}

func (c *GinExt) Session() *GinExt {
	v := *c
	v.DbInstance = c.DbInstance.Session(&gorm.Session{})
	return &v
}
func (c *GinExt) FilePath(path string) string {
	return filepath.Join(c.AppDir, path)
}

func (c *GinExt) LoadSettings(isMigrate bool) {
	log.Default().SetFlags(log.Lshortfile | log.Ltime | log.Ldate)

	confData, err := ioutil.ReadFile(c.ConfFile)
	if err != nil {
		log.Printf("load conf fail %s - %v", c.ConfFile, err)
		return
	}
	err = json.Unmarshal(confData, c)
	if err != nil {
		log.Printf("load conf fail %s - %v", c.ConfFile, err)
	} else {
		log.Println("Load done ", c.ConfFile)
	}

	if len(c.LogFile) > 0 && !isMigrate {
		gin.DisableConsoleColor()
		logFile, err := os.OpenFile(c.LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Panic(err)
		}
		c.LogWriter = io.MultiWriter(logFile)
		log.SetOutput(logFile)
		gin.DefaultWriter = c.LogWriter
	}
}

func (c *GinExt) Init() (err error) {

	if len(c.DbDSN) > 0 {
		err = c.initDB()
		if err != nil {
			return nil
		}
	}

	if len(c.SessionStore) > 0 {
		err = c.initSession()
	}
	return err
}

func (c *GinExt) IsUnitTest() bool {
	return c.DbDriver == "sqlite" && strings.Contains(c.DbDSN, "memory")
}

func (c *GinExt) initDB() (err error) {
	newLogger := logger.New(
		log.New(c.LogWriter, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  logger.Warn, // Log level
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
			Colorful:                  false,       // Disable color
		},
	)

	cfg := &gorm.Config{
		Logger:                 newLogger,
		SkipDefaultTransaction: true,
	}

	err = c.createDatabaseInstance(cfg)
	if err != nil {
		log.Panicf("connect db fail %v", err)
	}
	err = c.DbInstance.AutoMigrate(&GinExtConfig{})
	if err != nil {
		log.Panicf("Migrate GinExtConfig Fail %v", err)
	}
	return nil
}

func (c *GinExt) initSession() (err error) {
	if c.SessionStore == "memstore" {
		c.sessionStore = memstore.NewStore([]byte(c.SessionSecret))
	} else if c.SessionStore == "redis" {
		store, err := redis.NewStore(10, "tcp", c.RedisAddr, "", []byte(c.SessionSecret))
		if err != nil {
			log.Printf("redis session fail (fallback cookie) -%v", err)
			c.sessionStore = cookie.NewStore([]byte(c.SessionSecret))
		} else {
			c.sessionStore = store
		}
	} else {
		c.sessionStore = cookie.NewStore([]byte(c.SessionSecret))
	}
	return nil
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent) // 204
			return
		}

		c.Next()
	}
}

// Init Gin middleware
func (cfg *GinExt) WithGinExt(r *gin.Engine) {
	gin.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, nuHandlers int) {}
	r.Use(sessions.Sessions(cfg.SessionName, cfg.sessionStore))
	r.Use(CORSMiddleware())

	r.Use(func(c *gin.Context) {
		c.Set(ConfigField, cfg)
		c.Set(DBField, cfg.DbInstance)
		c.Next()
	})

	if gin.Mode() != gin.ReleaseMode {
		registerDocHandler(r)
	}
}

func GetValueEx(db *gorm.DB, key string) string {
	var v GinExtConfig
	newKey := strings.ToUpper(key)
	result := db.Where("key", newKey).Take(&v)
	if result.Error != nil {
		return ""
	}
	return v.Value
}

func GetInt64ValueEx(db *gorm.DB, key string, defaultVal int64) int64 {
	v := GetValueEx(db, key)
	if v == "" {
		return defaultVal
	}
	val, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return defaultVal
	}
	return val
}

func GetIntValueEx(db *gorm.DB, key string, defaultVal int) int {
	v := GetValueEx(db, key)
	if v == "" {
		return defaultVal
	}
	val, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return defaultVal
	}
	return int(val)
}

func SetValueEx(db *gorm.DB, key, value string) {
	var v GinExtConfig
	newKey := strings.ToUpper(key)
	result := db.Where("key", newKey).Take(&v)
	if result.Error != nil {
		newV := &GinExtConfig{
			Key:   newKey,
			Value: value,
		}
		db.Create(&newV)
		return
	}
	db.Model(&GinExtConfig{}).Where("key", newKey).UpdateColumn("value", value)
}

func (cfg *GinExt) GetValue(key string) string {
	return GetValueEx(cfg.DbInstance, key)
}

func (cfg *GinExt) CheckValue(key, defaultValue string) {
	if len(cfg.GetValue(key)) <= 0 {
		cfg.SetValue(key, defaultValue)
	}
}

func (cfg *GinExt) SetValue(key, value string) {
	SetValueEx(cfg.DbInstance, key, value)
}
