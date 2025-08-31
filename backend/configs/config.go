package configs

type Configuration struct {
	Server   ServerConfig   `mapstructure:"server" json:"server" yaml:"server"`
	Database DatabaseConfig `mapstructure:"database" json:"database" yaml:"database"`
	JWT      JWTConfig      `mapstructure:"jwt" json:"jwt" yaml:"jwt"`
}

type CORSConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins" json:"allowed_origins" yaml:"allowed_origins"`
	AllowedMethods []string `mapstructure:"allowed_methods" json:"allowed_methods" yaml:"allowed_methods"`
}

type ServerConfig struct {
	Port int        `mapstructure:"port" json:"port" yaml:"port"`
	CORS CORSConfig `mapstructure:"cors" json:"cors" yaml:"cors"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host" json:"host" yaml:"host"`
	Port     int    `mapstructure:"port" json:"port" yaml:"port"`
	Username string `mapstructure:"username" json:"username" yaml:"username"`
	Password string `mapstructure:"password" json:"password" yaml:"password"`
	DbName   string `mapstructure:"db_name" json:"db_name" yaml:"db_name"`
}

type JWTConfig struct {
	SecretKey string `mapstructure:"secret_key" json:"secret_key" yaml:"secret_key"`
	ExpiresIn int64  `mapstructure:"expires_in" json:"expires_in" yaml:"expires_in"`
}
