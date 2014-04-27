package main

import (
	"github.com/jcla1/goquizduell"
	"github.com/jcla1/goquizduell/util"
)

func main() {
	c := util.PrepareClient(os.Getenv("QD_USERNAME"), os.Getenv("QD_PASSWORD"), os.Getenv("QD_COOKIE_FILE"))

}
