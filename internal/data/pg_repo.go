package data

import (
	"playwallet/internal/cfgs"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PGRepo struct {
	db *gorm.DB
}

func NewPGRepo(cfg cfgs.PGCfg) (*PGRepo, error) {
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	ret := &PGRepo{db}
	return ret, nil
}
