package app

func ThrowErr(err error) {
	if err != nil {
		panic(err)
	}
}

func CheckErr(err error)  {
	if err != nil {
		panic(err)
	}
}
