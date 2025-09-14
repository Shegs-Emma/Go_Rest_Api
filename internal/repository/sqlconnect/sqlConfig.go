package sqlconnect

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func ConnectDb() (*sql.DB, error ){
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	dbport := os.Getenv("DB_PORT")
	host := os.Getenv("HOST")

	// connectionString := "root:M@rch041992@tcp(127.0.0.l:3306)/" + dbname
	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", user, password, host, dbport, dbname)
	db, err := sql.Open("mysql", connectionString)

	if err != nil {
		// panic(err)
		return nil, err
	}
	fmt.Println("Connected to mariadb...")
	return db, nil
}