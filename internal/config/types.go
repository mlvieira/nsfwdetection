package config

type DBConfig struct {
	Username string `toml:"username"`
	Password string `toml:"password"`
	Host     string `toml:"host"`
	Database string `toml:"database"`
	Port     int    `toml:"port"`
}

type FileHandlingConfig struct {
	UploadDir     string `toml:"upload_dir"`
	TempUploadDir string `toml:"temp_upload_dir"`
	MaxFileSizeMB int64  `toml:"max_file_size_mb"`
}

type ModelConfig struct {
	ModelPath string `toml:"model_path"`
}

type SecurityConfig struct {
	JWTSecretKey string `toml:"jwt_secret_key"`
}

type ServerConfig struct {
	Port       int     `toml:"port"`
	ReqPerSec  float64 `toml:"req_per_sec"`
	Burst      int     `toml:"burst"`
	DomainName string  `toml:"domain_name"`
}

type RedisConfig struct {
	Addr     string `toml:"addr"`
	Password string `toml:"password"`
	DB       int    `toml:"db"`
}
