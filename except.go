package go_followfix

import "github.com/cihub/seelog"

func ThrowErr(err error) {
	if err != nil {
		panic(err)
	}
}

func CheckErr(err error)  {
	if err != nil {
		seelog.Error(err)
	}
}
