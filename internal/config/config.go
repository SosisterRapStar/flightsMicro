package config

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type AppConfig struct {
	Server     Server     `yaml:"server"`
	API        API        `yaml:"api"`
	Kafka      Kafka      `yaml:"kafka"`
	Prometheus Prometheus `yaml:"prometheus"`
	Repository Repository `yaml:"repository"`
}

type Server struct {
	Address string `yaml:"address" env:"SERVER_ADDRESS" env_default:":8080"`
}

type API struct {
	Timeout           time.Duration `yaml:"timeout" env:"API_TIMEOUT" env_default:"30s"`
	ReadHeaderTimeout time.Duration `yaml:"read_header_timeout" env:"API_READ_HEADER_TIMEOUT" env_default:"5s"`
}

type Kafka struct {
	URL      string   `yaml:"url" env:"KAFKA_URL" env_default:"localhost:9092"`
	Brokers  []string `yaml:"brokers" env:"KAFKA_BROKERS" env_default:"localhost:9092"`
	Version  string   `yaml:"version" env:"KAFKA_VERSION" env_default:""`
	Topics   []string `yaml:"topics" env:"KAFKA_TOPICS"`
	GroupID  string   `yaml:"group_id" env:"KAFKA_GROUP_ID" env_default:"default-consumer-group"`
	Producer Producer `yaml:"producer"`
	Consumer Consumer `yaml:"consumer"`
}

type Producer struct {
	AckPolicy int16 `yaml:"ack_policy" env:"KAFKA_ACK_POLICY" env_default:"1"`
	RetryMax  int   `yaml:"retry_max" env:"KAFKA_RETRY_MAX" env_default:"5"`
}

type Consumer struct {
	AutoCommitEnable bool          `yaml:"auto_commit_enable" env:"KAFKA_AUTO_COMMIT" env_default:"false"`
	MaxWaitTime      time.Duration `yaml:"max_wait_time" env:"KAFKA_MAX_WAIT_TIME" env_default:"500ms"`
}

type Prometheus struct {
	Host string `yaml:"host" env:"PROM_HOST" env_default:"localhost:2112"`
}

type Repository struct {
	Hosts    string `yaml:"hosts" env:"DB_HOSTS" env_default:"localhost:5432"`
	Dbname   string `yaml:"dbname" env:"DB_NAME" env_default:"orders_db"`
	User     string `yaml:"user" env:"DB_USER" env_default:"postgres"`
	Password string `yaml:"password" env:"DB_PASSWORD" env_default:"postgres"`
	Schema   string `yaml:"schema" env:"DB_SCHEMA"`

	MaxIdleConn int `yaml:"max_idle_conn" env:"DB_MAX_IDLE_CONN" env_default:"5"`
	MaxOpenConn int `yaml:"max_open_conn" env:"DB_MAX_OPEN_CONN" env_default:"10"`

	MaxIdleLifetime time.Duration `yaml:"max_idle_lifetime" env:"DB_MAX_IDLE_LIFETIME" env_default:"5m"`
	MaxOpenLifetime time.Duration `yaml:"max_open_lifetime" env:"DB_MAX_OPEN_LIFETIME" env_default:"30m"`
}

func Load(path string) (*AppConfig, error) {
	var cfg AppConfig
	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		return nil, fmt.Errorf("reading config %s: %w", path, err)
	}
	return &cfg, nil
}

func MustLoad(path string) *AppConfig {
	cfg, err := Load(path)
	if err != nil {
		log.Fatalf("loading config: %v", err)
	}
	return cfg
}

func (r *Repository) DSN() string {
	splitHosts := strings.Split(r.Hosts, ",")

	dsnHosts := make([]string, 0, len(splitHosts))
	dsnPorts := make([]string, 0, len(splitHosts))

	for _, host := range splitHosts {
		splitHost := strings.Split(host, ":")
		dsnHosts = append(dsnHosts, splitHost[0])
		dsnPorts = append(dsnPorts, splitHost[1])
	}

	connStr := fmt.Sprintf(
		"port=%s host=%s user=%s password=%s dbname=%s sslmode=disable target_session_attrs=read-write",
		strings.Join(dsnPorts, ","),
		strings.Join(dsnHosts, ","),
		r.User,
		r.Password,
		r.Dbname,
	)
	if r.Schema != "" {
		connStr += fmt.Sprintf(" search_path=%s", r.Schema)
	}
	return connStr
}
