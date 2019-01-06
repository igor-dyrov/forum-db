package getters

import (
	_ "github.com/lib/pq"
)

func PanicIfError(err error) {
	if err != nil {
		panic(err)
	}
}
