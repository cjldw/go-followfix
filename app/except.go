package app

import (
	"log"
)

func ThrowErr(err error) {
	if err != nil {
		panic(err)
	}
}

func CheckErr(err error)  {
	if err != nil {
		log.Fatalf("catch exception %v", err.Error())
	}
}
