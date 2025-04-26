package conf

import (
	"fmt"
	"os"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/spf13/viper"
)

// 服务配置
type ServerConfig struct {
	Name     string        `mapstructure:"name"`
	Host     string        `mapstructure:"host"`
	Port     int           `mapstructure:"port"`
	Mode     string        `mapstructure:"mode"`
	Timeout  time.Duration `mapstructure:"timeout"`
	MaxConns int           `mapstructure:"max_conns"`
}

// MySQL配置
type MySQLConfig struct {
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	User        string `mapstructure:"user"`
	Password    string `mapstructure:"password"`
	Database    string `mapstructure:"database"`
	MaxIdleConn int    `mapstructure:"max_idle_conn"`
	MaxOpenConn int    `mapstructure:"max_open_conn"`
}

// Redis配置
type RedisConfig struct {
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	Password    string `mapstructure:"password"`
	DB          int    `mapstructure:"db"`
	MaxIdle     int    `mapstructure:"max_idle"`
	MaxActive   int    `mapstructure:"max_active"`
	IdleTimeout int    `mapstructure:"idle_timeout"`
}

// Nacos配置
type NacosConfig struct {
	Host      string `mapstructure:"host"`
	Port      uint64 `mapstructure:"port"`
	Namespace string `mapstructure:"namespace"`
	User      string `mapstructure:"user"`
	Password  string `mapstructure:"password"`
	DataId    string `mapstructure:"data_id"`
	Group     string `mapstructure:"group"`
}

// Zipkin配置
type ZipkinConfig struct {
	Endpoint    string  `mapstructure:"endpoint"`
	ServiceName string  `mapstructure:"service_name"`
	SampleRate  float64 `mapstructure:"sample_rate"`
}

// 认证配置
type AuthConfig struct {
	JWTSecret          string        `mapstructure:"jwt_secret"`
	JWTExpiration      time.Duration `mapstructure:"jwt_expiration"`
	RefreshExpiration  time.Duration `mapstructure:"refresh_expiration"`
	OAuth2RedirectURL  string        `mapstructure:"oauth2_redirect_url"`
	GithubClientID     string        `mapstructure:"github_client_id"`
	GithubClientSecret string        `mapstructure:"github_client_secret"`
	GoogleClientID     string        `mapstructure:"google_client_id"`
	GoogleClientSecret string        `mapstructure:"google_client_secret"`
}

// 日志配置
type LogConfig struct {
	Level      string `mapstructure:"level"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
}

// 应用配置
type Config struct {
	Server *ServerConfig `mapstructure:"server"`
	MySQL  *MySQLConfig  `mapstructure:"mysql"`
	Redis  *RedisConfig  `mapstructure:"redis"`
	Nacos  *NacosConfig  `mapstructure:"nacos"`
	Zipkin *ZipkinConfig `mapstructure:"zipkin"`
	Auth   *AuthConfig   `mapstructure:"auth"`
	Log    *LogConfig    `mapstructure:"log"`
}

var (
	// GlobalConfig 全局配置
	GlobalConfig *Config
)

// Init 初始化配置
func Init(configFile string) (*Config, error) {
	// 设置默认配置
	setDefaultConfig()

	// 检查环境变量
	if env := os.Getenv("APP_ENV"); env != "" {
		hlog.Infof("APP_ENV: %s", env)
		viper.SetConfigName(fmt.Sprintf("config.%s", env))
	} else {
		viper.SetConfigName("config")
	}

	// 设置配置路径
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.AddConfigPath("config")
		viper.AddConfigPath(".")
	}

	viper.SetConfigType("yaml")

	// 允许环境变量覆盖
	viper.AutomaticEnv()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		hlog.Errorf("读取配置文件失败: %v", err)
		return nil, err
	}

	// 解析配置
	if err := viper.Unmarshal(&GlobalConfig); err != nil {
		hlog.Errorf("解析配置失败: %v", err)
		return nil, err
	}

	hlog.Infof("配置加载成功: %s", viper.ConfigFileUsed())
	return GlobalConfig, nil
}

// 设置默认配置
func setDefaultConfig() {
	// 服务配置
	viper.SetDefault("server.name", "user-service")
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.mode", "release")
	viper.SetDefault("server.timeout", "30s")
	viper.SetDefault("server.max_conns", 1000)

	// MySQL配置
	viper.SetDefault("mysql.host", "localhost")
	viper.SetDefault("mysql.port", 3306)
	viper.SetDefault("mysql.user", "root")
	viper.SetDefault("mysql.password", "")
	viper.SetDefault("mysql.database", "user_service")
	viper.SetDefault("mysql.max_idle_conn", 10)
	viper.SetDefault("mysql.max_open_conn", 100)

	// Redis配置
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.max_idle", 10)
	viper.SetDefault("redis.max_active", 100)
	viper.SetDefault("redis.idle_timeout", 300)

	// 日志配置
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.filename", "logs/app.log")
	viper.SetDefault("log.max_size", 500)
	viper.SetDefault("log.max_backups", 10)
	viper.SetDefault("log.max_age", 30)
	viper.SetDefault("log.compress", true)

	// 认证配置
	viper.SetDefault("auth.jwt_expiration", "24h")
	viper.SetDefault("auth.refresh_expiration", "720h") // 30 days
}
