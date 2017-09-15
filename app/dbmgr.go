package app
import (
	"database/sql"
	"fmt"
	"errors"
	_ "github.com/go-sql-driver/mysql"
)

type DbMgr struct {
	Db *sql.DB
}

// db list instance
var dbList map[string] *sql.DB = make(map[string]*sql.DB)

// NewDbMgr struct
func NewDbMgr() *DbMgr {
	return &DbMgr{}
}

// InitializeDbList initialize Db List instance
func (dbMgr *DbMgr) InitializeDbList(dbConfigMap map[string]DbInst)  {
	for dbKey, dbConf := range dbConfigMap {
		_, ok := dbList[dbKey]
		if ok { continue }
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
	fmt.Println("----------------------")
	fmt.Println(dbClient)
	fmt.Println("----------------------")
	/*
	if dbClient.Ping() != nil {
		return nil, errors.New(fmt.Sprintf("db 【%s】 go away！", dbKey))
	}
	*/
	dbMgr.Db = dbClient
	return dbMgr, nil
}
