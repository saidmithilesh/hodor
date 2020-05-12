package main

import (
	"github.com/saidmithilesh/hodor/server"
	"github.com/saidmithilesh/hodor/utils"
)

func main() {
	config := utils.LoadConfig()
	utils.BuildLogger(&config)

	srv := server.HTTPServer{}
	srv.Build(&config)
	srv.Start(&config)
}
