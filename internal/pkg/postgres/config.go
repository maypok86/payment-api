package postgres

import "fmt"

type ConnectionConfig struct {
	host     string
	port     string
	dbname   string
	username string
	password string
	sslmode  string
}

func NewConnectionConfig(host, port, dbname, username, password, sslmode string) ConnectionConfig {
	return ConnectionConfig{
		host:     host,
		port:     port,
		dbname:   dbname,
		username: username,
		password: password,
		sslmode:  sslmode,
	}
}

func (cc ConnectionConfig) getDSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cc.username,
		cc.password,
		cc.host,
		cc.port,
		cc.dbname,
		cc.sslmode,
	)
}
