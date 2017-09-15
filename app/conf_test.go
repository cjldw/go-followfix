package app
import (
	"testing"
	"os"
)

func TestNewAppConf(t *testing.T) {
	env := os.Getenv("APP_ENV")
	t.Log(env)
	appConf := NewAppConf()
	appConf, err := appConf.Load("./conf/app.toml", false)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(appConf)
}
