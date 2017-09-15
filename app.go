package go_followfix

import "github.com/cihub/seelog"

func Run()  {
	defer func() {
		if r := recover(); r != nil {
			seelog.Error(r)
		}
	}()
	_, err := NewAppConf().Load("./conf/app.toml", false)
	if err != nil {
		panic(err)
	}

	
}
