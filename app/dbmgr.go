package app

import (
	"database/sql"
	"fmt"
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"sync"
)

type DbMgr struct {
	dbList map[string] *sql.DB
	Lock *sync.RWMutex
	dbConfigMap map[string]DbInst
}

// NewDbMgr struct
func NewDbMgr() *DbMgr {
	return &DbMgr{dbList: make(map[string]*sql.DB), Lock:&sync.RWMutex{}}
}

// InitializeDbList initialize Db List instance
func (dbMgr *DbMgr) InitializeDbList(dbConfig map[string]DbInst)  {
	dbMgr.dbConfigMap = dbConfig
	for dbKey, dbConf := range dbConfig {
		_, ok := dbMgr.dbList[dbKey]
		if ok { continue }
		log.Printf("initialize dbconnection %s \n", dbKey)
		dbClient, err := sql.Open(dbConf.Driver, dbConf.Dsn)
		ThrowErr(err)
		dbClient.SetMaxOpenConns(dbConf.MaxIdleConns)
		dbClient.SetMaxIdleConns(dbConf.MaxIdleConns)
		dbMgr.dbList[dbKey] = dbClient
	}
}

// GetDbByName get database instance by configure file name
func (dbMgr *DbMgr) GetDbByName(dbKey string) (*sql.DB, error) {
	dbClient, ok := dbMgr.dbList[dbKey]
	if !ok {
		return nil, errors.New(fmt.Sprintf("db 【%s】not exists！", dbKey))
	}
	dbErr := dbClient.Ping()
	if dbErr != nil { // connect one more time
		dbClient.Close()
		dbConf, ok := dbMgr.dbConfigMap[dbKey]
		if !ok {
			return nil, errors.New(fmt.Sprintf(" 配置【%s】不存在 ！", dbKey))
		}
		log.Println("Reconnect MySQL !")
		newDbClient, err := sql.Open(dbConf.Driver, dbConf.Dsn)
		fmt.Println(newDbClient)
		ThrowErr(err)
		//dbList[dbKey], dbClient = newDbClient, newDbClient
		dbClient = newDbClient
		dbMgr.dbList[dbKey] = newDbClient
	}
	return dbClient, nil
}

