package main

import (
	"Supernove2024/svr"
	"flag"
	"log"
)

func main() {
	configFile := flag.String("config", "mini-router.yaml", "")
	address := flag.String("address", "mini-router.yaml", "")
	flag.Parse()
	srv, err := svr.NewConfigSvr(*configFile)
	if err != nil {
		log.Fatalln(err)
	}
	if err = srv.Serve(*address); err != nil {
		log.Fatalln(err)
	}
}
