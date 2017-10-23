package app
import (
	"path/filepath"
	"strings"
	"github.com/BurntSushi/toml"
	"os"
)

// application configure
type AppConf struct {
	Name string
	Version string
	Hotfix bool
	DbConf map[string] DbInst
	RedisConf map[string] RedisInst
}

// database instance configure
type DbInst struct {
	Driver string
	Dsn string
	MaxOpenConns int
	MaxIdleConns int
}

// redis instance configure
type RedisInst  struct {
	Host string
	Port int
	Db int
	Password string
	Socket string
}

// isLoad the flag of appConf is load or not
var isLoad bool = false

func NewAppConf() *AppConf {
	return &AppConf{}
}

// Load load configure from filename
func (appConf *AppConf) Load(filename string, forceReload bool)  (*AppConf, error){
	// appConf is load and force reload is false return direction
	if isLoad && !forceReload{
		return appConf, nil
	}
	filename, err := appConf.getEnvConfigPath(filename)
	if  err != nil {
		return nil, err
	}
	_, err = toml.DecodeFile(filename, appConf)
	if err != nil {
		return  nil, err
	}
	// mark appConf as loaded
	isLoad = true
	return appConf, nil
}

// getEnvConfigPath  get file name by environment
// if APP_ENV environment variable not set . this will return filename without suffix
// else will return filename{APP_ENV}
func (appConfig *AppConf) getEnvConfigPath(filename string) (string, error){
	isAbsPath := filepath.IsAbs(filename)
	var err error
	if !isAbsPath {
		filename, err = filepath.Abs(filename)
		if err != nil {
			return "", err
		}
	}
	env := os.Getenv("APP_ENV")
	if env == "" {
		return filename, nil
	}
	suffixIndex := strings.LastIndex(filename, ".")
	return filename[:suffixIndex+1] + env + filename[suffixIndex:], nil
}
