package example

import "log"

func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
