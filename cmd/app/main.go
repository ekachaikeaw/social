package main

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, err := sql.Open("mysql", "root:password@tcp(localhost:3306)/ecomm?parseTime=true")
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}
	// cfg := config.Config{
	// 	Addr: env.GetString("ADDR", ":8080"),
	// 	Db: config.DbConfig{
	// 		Addr:        env.GetString("DB_ADDR", "postgres://user:password@localhost:5432/socialnetwork?sslmode=disable"),
	// 		MaxOpenConn: env.GetInt("DB_MAX_OPEN_CONN", 30),
	// 		MaxIdleConn: env.GetInt("DB_MAX_IDLE_CONN", 30),
	// 		MaxIdleTime: env.GetString("DB_MAX_IDLE_TIME", "15m"),
	// 	},
	// }
	// fmt.Println(cfg)
	// db, err := db.New(
	// 	cfg.Db.Addr,
	// 	cfg.Db.MaxOpenConn,
	// 	cfg.Db.MaxIdleConn,
	// 	cfg.Db.MaxIdleTime,
	// )
	// if err != nil {
	// 	log.Panic(err)
	// }
	// defer db.Close()
	// log.Println("database connection pool established")

	// storage := store.NewStorage(db)
	// app := &config.Application{
	// 	Config: cfg,
	// 	Store:  storage,
	// }

	// mux := app.Mount()

	// log.Fatal(app.Run(mux))
}
