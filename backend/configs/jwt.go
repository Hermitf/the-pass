package configs

type JWTConfig struct {
	SecretKey string `mapstructure:"secret_key" json:"secret_key" yaml:"secret_key"`
	ExpiresIn int64  `mapstructure:"expires_in" json:"expires_in" yaml:"expires_in"`
}