package embedding

import (
	"database/sql"
	"fmt"
	"os"
	"time"
)

const addDefaultEmbeddingUser = `INSERT INTO users (
    user_id,
    login,
    name,
    password
) VALUES (
    0,
    '',
    '',
    ''
)`

const updateDefaultEmbeddingUser = `UPDATE users
SET embedding = sub.avg_embedding
FROM (
    SELECT AVG(embedding) AS avg_embedding
    FROM users WHERE user_id <> 0
) AS sub
WHERE user_id = 0`

func RunSetterDefaultEmbedding() error {
	for {
		db := connectToDB()
		var userID string
		if err := db.QueryRow("SELECT user_id FROM users WHERE user_id = 0").Scan(&userID); err != nil {
			if _, err := db.Exec(addDefaultEmbeddingUser); err != nil {
				return err
			}
		}

		if _, err := db.Exec(updateDefaultEmbeddingUser); err != nil {
			return err
		}

		if err := db.Close(); err != nil {
			return err
		}

		time.Sleep(time.Hour * 23)
	}
}

func connectToDB() *sql.DB {
	// host := "localhost"
	// port := "5432"
	// user := "postgres"
	// password := "123"
	// dbname := "project_farm"

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}

	return db
}
