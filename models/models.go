// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package models

import (
	"fmt"
	"os"
	"path"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/lunny/xorm"
	// _ "github.com/mattn/go-sqlite3"

	"github.com/gogits/gogs/modules/base"
)

var (
	orm    *xorm.Engine
	tables []interface{}

	HasEngine bool

	DbCfg struct {
		Type, Host, Name, User, Pwd, Path, SslMode string
	}

	UseSQLite3 bool
)

func init() {
	tables = append(tables, new(User), new(PublicKey), new(Repository), new(Watch),
		new(Action), new(Access), new(Issue), new(Comment), new(Oauth2))
}

func LoadModelsConfig() {
	DbCfg.Type = base.Cfg.MustValue("database", "DB_TYPE")
	if DbCfg.Type == "sqlite3" {
		UseSQLite3 = true
	}
	DbCfg.Host = base.Cfg.MustValue("database", "HOST")
	DbCfg.Name = base.Cfg.MustValue("database", "NAME")
	DbCfg.User = base.Cfg.MustValue("database", "USER")
	DbCfg.Pwd = base.Cfg.MustValue("database", "PASSWD")
	DbCfg.SslMode = base.Cfg.MustValue("database", "SSL_MODE")
	DbCfg.Path = base.Cfg.MustValue("database", "PATH", "data/gogs.db")
}

func NewTestEngine(x *xorm.Engine) (err error) {
	switch DbCfg.Type {
	case "mysql":
		x, err = xorm.NewEngine("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8",
			DbCfg.User, DbCfg.Pwd, DbCfg.Host, DbCfg.Name))
	case "postgres":
		x, err = xorm.NewEngine("postgres", fmt.Sprintf("user=%s password=%s dbname=%s sslmode=%s",
			DbCfg.User, DbCfg.Pwd, DbCfg.Name, DbCfg.SslMode))
	// case "sqlite3":
	// 	os.MkdirAll(path.Dir(DbCfg.Path), os.ModePerm)
	// 	x, err = xorm.NewEngine("sqlite3", DbCfg.Path)
	default:
		return fmt.Errorf("Unknown database type: %s", DbCfg.Type)
	}
	if err != nil {
		return fmt.Errorf("models.init(fail to conntect database): %v", err)
	}
	return x.Sync(tables...)
}

func SetEngine() (err error) {
	switch DbCfg.Type {
	case "mysql":
		orm, err = xorm.NewEngine("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8",
			DbCfg.User, DbCfg.Pwd, DbCfg.Host, DbCfg.Name))
	case "postgres":
		orm, err = xorm.NewEngine("postgres", fmt.Sprintf("user=%s password=%s dbname=%s sslmode=%s",
			DbCfg.User, DbCfg.Pwd, DbCfg.Name, DbCfg.SslMode))
	case "sqlite3":
		os.MkdirAll(path.Dir(DbCfg.Path), os.ModePerm)
		orm, err = xorm.NewEngine("sqlite3", DbCfg.Path)
	default:
		return fmt.Errorf("Unknown database type: %s", DbCfg.Type)
	}
	if err != nil {
		return fmt.Errorf("models.init(fail to conntect database): %v", err)
	}

	// WARNNING: for serv command, MUST remove the output to os.stdout,
	// so use log file to instead print to stdout.
	execDir, _ := base.ExecDir()
	logPath := execDir + "/log/xorm.log"
	os.MkdirAll(path.Dir(logPath), os.ModePerm)

	f, err := os.Create(logPath)
	if err != nil {
		return fmt.Errorf("models.init(fail to create xorm.log): %v", err)
	}
	orm.Logger = f

	orm.ShowSQL = true
	orm.ShowDebug = true
	orm.ShowErr = true
	return nil
}

func NewEngine() (err error) {
	if err = SetEngine(); err != nil {
		return err
	}
	if err = orm.Sync(tables...); err != nil {
		return fmt.Errorf("sync database struct error: %v\n", err)
	}
	return nil
}

type Statistic struct {
	Counter struct {
		User, PublicKey, Repo, Watch, Action, Access int64
	}
}

func GetStatistic() (stats Statistic) {
	stats.Counter.User, _ = orm.Count(new(User))
	stats.Counter.PublicKey, _ = orm.Count(new(PublicKey))
	stats.Counter.Repo, _ = orm.Count(new(Repository))
	stats.Counter.Watch, _ = orm.Count(new(Watch))
	stats.Counter.Action, _ = orm.Count(new(Action))
	stats.Counter.Access, _ = orm.Count(new(Access))
	return
}
