package main

import (
	"database/sql"
	"log"

	"github.com/kamilwrzyszcz/go_example/api"
	db "github.com/kamilwrzyszcz/go_example/db/sqlc"
	"github.com/kamilwrzyszcz/go_example/session"
	"github.com/kamilwrzyszcz/go_example/util"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config: ", err)
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db: ", err)
	}

	store := db.NewStore(conn)
	sessionClient, err := session.NewRedisClient(config.RedisAddress, config.RedisPassword)
	if err != nil {
		log.Fatal("cannot connect to session client: ", err)
	}

	server, err := api.NewServer(config, store, sessionClient)
	if err != nil {
		log.Fatal("cannot create server: ", err)
	}

	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("cannot start server: ", err)
	}
}
