package db

import (
	"github.com/SosisterRapStar/LETI-PaperTestMicroservices/internal/config"
	"github.com/jmoiron/sqlx"
)

const (
	pgxDriverName = "pgx"
)

type Postgres struct {
	DB *sqlx.DB
}

func NewPostgres(cfg *config.Repository) (*sqlx.DB, error) {
	db, err := sqlx.Connect(pgxDriverName, cfg.DSN())
	if err != nil {
		return nil, err
	}
	db.SetMaxIdleConns(cfg.MaxIdleConn)
	db.SetMaxOpenConns(cfg.MaxOpenConn)

	db.SetConnMaxIdleTime(cfg.MaxIdleLifetime)
	db.SetConnMaxLifetime(cfg.MaxOpenLifetime)

	return db, nil
}
