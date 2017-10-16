package app

import (
	"database/sql"
	"fmt"
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

type DbMgr struct {
	Db *sql.DB
}

// db list instance
var dbList map[string] *sql.DB = make(map[string]*sql.DB)
var dbConfigMap map[string] DbInst = make(map[string]DbInst)

// NewDbMgr struct
func NewDbMgr() *DbMgr {
	return &DbMgr{}
}

// InitializeDbList initialize Db List instance
func (dbMgr *DbMgr) InitializeDbList(dbConfig map[string]DbInst)  {
	dbConfigMap = dbConfig
	for dbKey, dbConf := range dbConfig {
		_, ok := dbList[dbKey]
		if ok { continue }
		log.Printf("initialize dbconnection %s \n", dbKey)
		dbClient, err := sql.Open(dbConf.Driver, dbConf.Dsn)
		ThrowErr(err)
		dbList[dbKey] = dbClient
	}
}

// GetDbByName get database instance by configure file name
func (dbMgr *DbMgr) GetDbByName(dbKey string) (*DbMgr, error) {
	dbClient, ok := dbList[dbKey]
	if !ok {
		return nil, errors.New(fmt.Sprintf("db 【%s】not exists！", dbKey))
	}
	if dbClient.Ping() != nil { // connect one more time
		dbClient.Close()
		dbConf, ok := dbConfigMap[dbKey]
		if !ok {
			return nil, errors.New(fmt.Sprintf("db config【%s】 missing ！", dbKey))
		}
		log.Println("Reconnect MySQL !")
		newDbClient, err := sql.Open(dbConf.Driver, dbConf.Dsn)
		fmt.Println(newDbClient)
		ThrowErr(err)
		//dbList[dbKey], dbClient = newDbClient, newDbClient
		dbClient = newDbClient
		dbList[dbKey] = newDbClient
	}
	dbMgr.Db = dbClient
	return dbMgr, nil
}
