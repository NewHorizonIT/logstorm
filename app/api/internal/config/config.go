package config

import "time"

// Config represents the root application configuration.
type Config struct {
	App        AppConfig        `mapstructure:"app" validate:"required"`
	Server     ServerConfig     `mapstructure:"server" validate:"required"`
	Database   DatabaseConfig   `mapstructure:"database" validate:"required"`
	ClickHouse ClickHouseConfig `mapstructure:"clickhouse" validate:"required"`
	Logging    LoggingConfig    `mapstructure:"logging" validate:"required"`
	Auth       AuthConfig       `mapstructure:"auth" validate:"required"`
	CORS       CORSConfig       `mapstructure:"cors" validate:"required"`
}

type AppConfig struct {
	Name        string `mapstructure:"name" validate:"required,min=3,max=50"`
	Environment string `mapstructure:"environment" validate:"required,oneof=development staging production test"`
	Version     string `mapstructure:"version"`
}

type ServerConfig struct {
	Protocol string `mapstructure:"protocol" validate:"required,oneof=http https"`
	Host     string `mapstructure:"host" validate:"required"`
	Port     int    `mapstructure:"port" validate:"required,min=1,max=65535"`
	BasePath string `mapstructure:"base_path" validate:"required,startswith=/"`
}

type DatabaseConfig struct {
	Driver          string        `mapstructure:"driver" validate:"required,oneof=postgres mysql sqlite"`
	Host            string        `mapstructure:"host" validate:"required"`
	Port            int           `mapstructure:"port" validate:"required,min=1,max=65535"`
	Username        string        `mapstructure:"username" validate:"required"`
	Password        string        `mapstructure:"password"`
	Name            string        `mapstructure:"name" validate:"required"`
	TLSEnabled      bool          `mapstructure:"tls_enabled"`
	MaxOpenConns    int           `mapstructure:"max_open_conns" validate:"gte=1"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns" validate:"gte=0"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime" validate:"gte=0"`
}

type ClickHouseConfig struct {
	Host        string `mapstructure:"host" validate:"required"`
	HTTPPort    int    `mapstructure:"http_port" validate:"required,min=1,max=65535"`
	NativePort  int    `mapstructure:"native_port" validate:"required,min=1,max=65535"`
	Username    string `mapstructure:"username" validate:"required"`
	Password    string `mapstructure:"password"`
	Database    string `mapstructure:"database" validate:"required"`
	Secure      bool   `mapstructure:"secure"`
	Compression bool   `mapstructure:"compression"`
}

type LoggingConfig struct {
	Level          string `mapstructure:"level" validate:"required,oneof=debug info warn error fatal panic"`
	Format         string `mapstructure:"format" validate:"required,oneof=json console"`
	Output         string `mapstructure:"output" validate:"required"`
	Environment    string `mapstructure:"environment" validate:"required,oneof=development staging production test"`
	ConsoleEnabled bool   `mapstructure:"console_enabled"`

	// File logging configuration
	FilePath    string `mapstructure:"file_path" validate:"required_if=Output file"`
	FileEnabled bool   `mapstructure:"file_enabled"`

	// Rotation configuration
	RotationEnabled    bool `mapstructure:"rotation_enabled"`
	RotationMaxSize    int  `mapstructure:"rotation_max_size" validate:"required_if=RotationEnabled true,gte=1"`
	RotationMaxAge     int  `mapstructure:"rotation_max_age" validate:"required_if=RotationEnabled true,gte=1"`
	RotationMaxBackups int  `mapstructure:"rotation_max_backups" validate:"required_if=RotationEnabled true,gte=0"`
	CompressionEnabled bool `mapstructure:"compression_enabled"`
}

type AuthConfig struct {
	JWTSecret       string        `mapstructure:"jwt_secret" validate:"required,min=32"`
	AccessTokenTTL  time.Duration `mapstructure:"access_token_ttl" validate:"required"`
	RefreshTokenTTL time.Duration `mapstructure:"refresh_token_ttl" validate:"required"`
}

type CORSConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins" validate:"required,min=1,dive,required"`
	AllowedMethods []string `mapstructure:"allowed_methods" validate:"required,min=1,dive,required"`
	AllowedHeaders []string `mapstructure:"allowed_headers" validate:"required,min=1,dive,required"`
}
