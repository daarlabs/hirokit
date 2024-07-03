package esquel

import (
	"fmt"
	"os"
	"strings"
	"time"
	
	_ "github.com/lib/pq"
)

type Config interface{}

type config struct {
	configType int
	value      any
}

const (
	DefaultTimout = 30 * time.Second
)

const (
	configTypeDriver = iota
	configTypeHost
	configTypePort
	configTypeDbname
	configTypeUser
	configTypePassword
	configTypeSsl
	configTypeCertPath
	configTypeLog
	configTypeTimeout
)

const (
	sslDisable    = "disable"
	sslAllow      = "allow"
	sslPrefer     = "prefer"
	sslRequire    = "require"
	sslVerifyCa   = "verify-ca"
	sslVerifyFull = "verify-full"
)

func Connect(configs ...Config) (*DB, error) {
	driver, dataSource, log, timeout, err := createConnectionDataSource(configs...)
	if err != nil {
		return nil, err
	}
	db, err := Open(driver, dataSource)
	db.log = log
	db.timeout = timeout
	return db, err
}

func MustConnect(configs ...Config) *DB {
	db, err := Connect(configs...)
	if err != nil {
		panic(err)
	}
	return db
}

func createConnectionDataSource(configs ...Config) (string, string, bool, time.Duration, error) {
	var driver string
	var log bool
	var timeout time.Duration
	props := make([]string, 0)
	for _, item := range configs {
		c, ok := item.(config)
		if !ok {
			continue
		}
		switch c.configType {
		case configTypeLog:
			log = c.value.(bool)
		case configTypeTimeout:
			timeout = c.value.(time.Duration)
		case configTypeDriver:
			driver = fmt.Sprintf("%v", c.value)
		case configTypeHost:
			props = append(props, fmt.Sprintf("host=%v", c.value))
		case configTypePort:
			props = append(props, fmt.Sprintf("port=%v", c.value))
		case configTypeDbname:
			props = append(props, fmt.Sprintf("dbname=%v", c.value))
		case configTypeUser:
			props = append(props, fmt.Sprintf("user=%v", c.value))
		case configTypePassword:
			props = append(props, fmt.Sprintf("password=%v", c.value))
		case configTypeSsl:
			props = append(props, fmt.Sprintf("sslmode=%v", c.value))
		case configTypeCertPath:
			v := fmt.Sprintf("%v", c.value)
			dir, err := os.Getwd()
			if err != nil {
				return "", "", false, 0, err
			}
			if strings.HasSuffix(dir, "/") {
				dir = strings.TrimSuffix(dir, "/")
			}
			if !strings.HasPrefix(v, "/") {
				v = "/" + v
			}
			props = append(props, fmt.Sprintf("sslrootcert=%s", dir+v))
		}
	}
	if timeout == 0 {
		timeout = DefaultTimout
	}
	return driver, strings.Join(props, " "), log, timeout, nil
}

func WithLog(log bool) Config {
	return config{
		configType: configTypeLog,
		value:      log,
	}
}

func WithTimeout(duration time.Duration) Config {
	return config{
		configType: configTypeLog,
		value:      log,
	}
}

func WithPostgres() Config {
	return config{
		configType: configTypeDriver,
		value:      Postgres,
	}
}

func WithMysql() Config {
	return config{
		configType: configTypeDriver,
		value:      Mysql,
	}
}

func WithDriver(driver string) Config {
	return config{
		configType: configTypeDriver,
		value:      driver,
	}
}

func WithHost(host string) Config {
	return config{
		configType: configTypeHost,
		value:      host,
	}
}

func WithPort(port int) Config {
	return config{
		configType: configTypePort,
		value:      port,
	}
}

func WithDbname(dbname string) Config {
	return config{
		configType: configTypeDbname,
		value:      dbname,
	}
}

func WithUser(user string) Config {
	return config{
		configType: configTypeUser,
		value:      user,
	}
}

func WithPassword(password string) Config {
	return config{
		configType: configTypePassword,
		value:      password,
	}
}

func WithSsl(sslmode string) Config {
	return config{
		configType: configTypeSsl,
		value:      sslmode,
	}
}

func WithSslDisable() Config {
	return config{
		configType: configTypeSsl,
		value:      sslDisable,
	}
}

func WithSslAllow() Config {
	return config{
		configType: configTypeSsl,
		value:      sslAllow,
	}
}

func WithSslPrefer() Config {
	return config{
		configType: configTypeSsl,
		value:      sslPrefer,
	}
}

func WithSslRequire() Config {
	return config{
		configType: configTypeSsl,
		value:      sslRequire,
	}
}

func WithSslVerifyCa() Config {
	return config{
		configType: configTypeSsl,
		value:      sslVerifyCa,
	}
}

func WithSslVerifyFull() Config {
	return config{
		configType: configTypeSsl,
		value:      sslVerifyFull,
	}
}

func WithCertPath(certpath string) Config {
	return config{
		configType: configTypeCertPath,
		value:      certpath,
	}
}
