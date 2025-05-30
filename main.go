package main

import (
	"fmt"
	"github.com/luckyhut/gator/internal/config"
)

func main() {
	// read config file
	conf := config.Read()
	//	fmt.Println(conf.DbUrl)
	conf.SetUser("ed")
	conf = config.Read()
	fmt.Println(conf)
}
