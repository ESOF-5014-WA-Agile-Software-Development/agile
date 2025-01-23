package runtime

import (
	"context"

	"gorm.io/gorm"
)

type Runtime struct {
	Flags  *Flags
	Config *Config
	Mysql  *gorm.DB
	Redis  *Redis
	Email  *Email
}

func New() (*Runtime, error) {
	rt := &Runtime{}

	flags, err := parseFlags()
	if err != nil {
		return nil, err
	}
	rt.Flags = flags

	config, err := loadConfig(flags.ConfigFile)
	if err != nil {
		return nil, err
	}
	rt.Config = config

	db, err := initMysql(&config.Mysql)
	if err != nil {
		return nil, err
	}
	rt.Mysql = db

	redis, err := newRedis(rt)
	if err != nil {
		return nil, err
	}
	rt.Redis = redis

	email, err := newEmail(config)
	if err != nil {
		return nil, err
	}
	rt.Email = email

	return rt, nil
}

func (r *Runtime) Close(ctx context.Context) error {
	// TODO ctx
	defer ctx.Done()

	err := r.Redis.Close()
	if err != nil {
		return err
	}

	sqlDB, err := r.Mysql.DB()
	if err != nil {
		return err
	}
	if err := sqlDB.Close(); err != nil {
		return err
	}

	return nil
}
