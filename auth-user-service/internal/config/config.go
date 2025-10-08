package config

type Config struct {
	ServerPort string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	RedisHost  string
	RedisPort  string
	JWTSecret  string
}

func Load() *Config {
	return &Config{
		ServerPort: "8080",
		DBHost:     "postgres",
		DBPort:     "5432",
		DBUser:     "postgres",
		DBPassword: "password",
		DBName:     "auth_user_db",
		RedisHost:  "redis",
		RedisPort:  "6379",
		JWTSecret:  "your-secret-key",
	}
}

func (c *Config) GetDBConnection() string {
	return "postgres://" + c.DBUser + ":" + c.DBPassword + "@" + c.DBHost + ":" + c.DBPort + "/" + c.DBName + "?sslmode=disable"
}
