package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config 应用程序配置
type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`

	Server   ServerConfig   `mapstructure:"server"`
	OAuth2   OAuth2Config   `mapstructure:"oauth2"`
	Admin    AdminConfig    `mapstructure:"admin"`
}

// AppConfig 应用配置
type AppConfig struct {
	Name        string `mapstructure:"name"`
	Version     string `mapstructure:"version"`
	Environment string `mapstructure:"environment"`
	Debug       bool   `mapstructure:"debug"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}


// ServerConfig 服务器配置
type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Host string `mapstructure:"host"`
}

// OAuth2Config OAuth2配置
type OAuth2Config struct {
	ClientID                string `mapstructure:"client_id"`
	ClientSecret            string `mapstructure:"client_secret"`
	AuthURL                 string `mapstructure:"auth_url"`
	TokenURL                string `mapstructure:"token_url"`
	UserInfoURL             string `mapstructure:"userinfo_url"`
	RedirectURI             string `mapstructure:"redirect_uri"`
	LogoutURL               string `mapstructure:"logout_url"`
	LogoutRedirectURIParam  string `mapstructure:"logout_redirect_uri_param"`
}

// AdminConfig 管理控制台配置
type AdminConfig struct {
	ConsoleURL string `mapstructure:"console_url"`
}

// Load 加载配置
func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("/etc/fly-print-cloud")

	// 设置环境变量前缀
	viper.SetEnvPrefix("FLY_PRINT")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 设置默认值
	setDefaults()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("read config file error: %w", err)
		}
		// 配置文件不存在时使用默认值和环境变量
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unmarshal config error: %w", err)
	}

	return &config, nil
}

// GetOAuth2UserInfoURL 获取 OAuth2 UserInfo URL
func GetOAuth2UserInfoURL() string {
	return viper.GetString("oauth2.userinfo_url")
}

// setDefaults 设置默认配置值
func setDefaults() {
	// App 默认值
	viper.SetDefault("app.name", "fly-print-cloud")
	viper.SetDefault("app.version", "0.1.0")
	viper.SetDefault("app.environment", "development")
	viper.SetDefault("app.debug", true)

	// Database 默认值
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "postgres")
	viper.SetDefault("database.dbname", "fly_print_cloud")
	viper.SetDefault("database.sslmode", "disable")

	// Redis 默认值
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)

	// Server 默认值
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)

	// OAuth2 默认值
	viper.SetDefault("oauth2.client_id", "")
	viper.SetDefault("oauth2.client_secret", "")
	viper.SetDefault("oauth2.auth_url", "")
	viper.SetDefault("oauth2.token_url", "")
	viper.SetDefault("oauth2.userinfo_url", "")
	viper.SetDefault("oauth2.redirect_uri", "")
	viper.SetDefault("oauth2.logout_url", "")
	viper.SetDefault("oauth2.logout_redirect_uri_param", "post_logout_redirect_uri")
	viper.SetDefault("admin.console_url", "http://localhost:3000")

	// Admin 创建配置
	viper.SetDefault("create_default_admin", "false")
	viper.SetDefault("default_admin_password", "")
}

// GetDSN 获取数据库连接字符串
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

// GetRedisAddr 获取Redis地址
func (c *RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// GetServerAddr 获取服务器地址
func (c *ServerConfig) GetServerAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}