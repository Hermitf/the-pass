package configs

type Configuration struct {
	Server   ServerConfig   `mapstructure:"server" json:"server" yaml:"server"`
	Database DatabaseConfig `mapstructure:"database" json:"database" yaml:"database"`
	JWT      JWTConfig      `mapstructure:"jwt" json:"jwt" yaml:"jwt"`
}
type ServerConfig struct {
	Port int `mapstructure:"port" json:"port" yaml:"port"`
}
type DatabaseConfig struct {
	Host     string `mapstructure:"host" json:"host" yaml:"host"`
	Port     int    `mapstructure:"port" json:"port" yaml:"port"`
	Username string `mapstructure:"username" json:"username" yaml:"username"`
	Password string `mapstructure:"password" json:"password" yaml:"password"`
	DbName   string `mapstructure:"db_name" json:"db_name" yaml:"db_name"`
}
