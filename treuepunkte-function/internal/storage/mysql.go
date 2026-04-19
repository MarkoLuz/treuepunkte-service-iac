package storage

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	_ "embed"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
)

//go:embed certs/global-bundle.pem
var rdsCABundle []byte

func OpenMySQL(appEnv, user, pass, host, port, db string) (*sql.DB, error) {
	var dsn string

	if appEnv == "local" {
		dsn = fmt.Sprintf(
			"%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci",
			user, pass, host, port, db,
		)
	} else {
		rootCertPool := x509.NewCertPool()

		if ok := rootCertPool.AppendCertsFromPEM(rdsCABundle); !ok {
			return nil, fmt.Errorf("failed to append CA cert")
		}

		err := mysql.RegisterTLSConfig("rds", &tls.Config{
			RootCAs:    rootCertPool,
			MinVersion: tls.VersionTLS12,
		})
		if err != nil {
			return nil, fmt.Errorf("register tls config: %w", err)
		}

		dsn = fmt.Sprintf(
			"%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci&tls=rds",
			user, pass, host, port, db,
		)
	}

	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	conn.SetMaxOpenConns(10)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(30 * time.Minute)

	if err := conn.Ping(); err != nil {
		return nil, err
	}

	return conn, nil
}