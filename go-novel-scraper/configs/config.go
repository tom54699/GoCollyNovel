package configs

import (
	"log"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type App struct {
	PrefixUrl     string
	ServerAddress string
	AppName       string

	LogSavePath string
	LogSaveName string
	LogFileExt  string

	MaxLogFiles int

	ImageStaticPath string
	ImageSavePath   string
}

type Server struct {
	RunMode      string
	HttpPort     int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type MongoDB struct {
	Host     string
	Port     int
	Name     string
	User     string
	Password string
}

type Redis struct {
	Addr     string
	Password string
	DB       int
}

type Jwt struct {
	Secret         string
	ExpirationDays int
}

type Config struct {
	App     App
	Server  Server
	MongoDB MongoDB
	Jwt     Jwt
	Redis   Redis
}

var C Config

// Setup initialize the configuration instance
func Setup() {
	viper.SetConfigType("toml")
	viper.SetConfigName("config")
	viper.AddConfigPath("./configs")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Fatalf("Config file not found")
		} else {
			log.Fatalf("Config file was found but another error was produced")
		}
	}

	viper.AutomaticEnv()

	err := viper.Unmarshal(&C)
	if err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}

	godotenv.Load()

	applyEnvVariables()
}

func applyEnvVariables() {
	C.App.PrefixUrl = viper.GetString("PREFIX_URL")
	C.App.ServerAddress = viper.GetString("SERVER_ADDRESS")
	C.App.AppName = viper.GetString("APP_NAME")
	C.Server.RunMode = viper.GetString("RUN_MODE")
	C.Server.HttpPort = viper.GetInt("HTTP_PORT")
	C.MongoDB.Host = viper.GetString("MONGODB_HOST")
	C.MongoDB.Port = viper.GetInt("MONGODB_PORT")
	C.MongoDB.Password = viper.GetString("MONGODB_PASSWORD")
	C.MongoDB.Name = viper.GetString("MONGODB_NAME")
	C.MongoDB.User = viper.GetString("MONGODB_USER")
	C.Redis.Addr = viper.GetString("REDIS_ADDR")
	C.Redis.Password = viper.GetString("REDIS_PASSWORD")
	C.Redis.DB = viper.GetInt("REDIS_DB")
	C.Jwt.Secret = viper.GetString("JWT_SECRET")
	C.Jwt.ExpirationDays = viper.GetInt("JWT_EXPIRATION_DAYS")
}
