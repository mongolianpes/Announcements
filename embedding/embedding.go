package embedding

import (
	"bytes"
	"database/sql"
	"errors"
	"io"
	"net/http"
	"os"
	"time"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

var client = &http.Client{
	Timeout: 30 * time.Second,
}

const userAdaptationRate = 0.9

type ollamaJsonResponse struct {
	Embedding []float64
}

type ollamaJsonRequest struct {
	Model  string
	Prompt string
}

var ollamaHost = os.Getenv("OLLAMA_HOST")

var cantConnectToOllamaError = errors.New("Невозможно подключиться к Ollama")
var envOSError = errors.New("Переменная OLLAMA_HOST должна иметь значение: адрес локальной нейросети ollama")

const insertCommand = "UPDATE announcements SET embedding = $1::float8[] WHERE announcement_id = $2"

func InsertEmbedding(db *sql.DB, rowID int, text string) error {
	if ollamaHost == "" {
		return envOSError
	}

	req := ollamaJsonRequest{
		Model:  "nomic-embed-text",
		Prompt: text,
	}

	jsonBody, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := client.Post(ollamaHost+"/api/embeddings", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		saveEmbeddingText(db, rowID, text)

		return cantConnectToOllamaError
	}
	if resp.StatusCode != http.StatusOK {
		return cantConnectToOllamaError
	}

	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	var result ollamaJsonResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	_, err = db.Exec(insertCommand, result.Embedding, rowID)
	if err != nil {
		return err
	}

	return nil
}

func UpdateUserEmbedding(db *sql.DB, userEmbedding, announcementEmbedding *[]float32, userID int32) error {
	for i := range *userEmbedding {
		(*userEmbedding)[i] = float32(userAdaptationRate)*(*userEmbedding)[i] + float32(1-userAdaptationRate)*(*announcementEmbedding)[i]
	}

	if _, err := db.Exec("UPDATE users SET embedding = $1::float8[] WHERE user_id = $2", userEmbedding, userID); err != nil {
		return err
	}

	return nil
}

func saveEmbeddingText(db *sql.DB, rowID int, text string) error {
	if _, err := db.Exec("INSERT INTO embeddings_announcements (embedding_id, text) VALUES ($1, $2)", rowID, text); err != nil {
		return err
	}

	return nil
}

func RetryInsertEmbeddings(db *sql.DB) {
	for {
		runInsertSavedEmbeddings(db)
		time.Sleep(time.Hour * 4)
	}
}

func runInsertSavedEmbeddings(db *sql.DB) error {
	offset := 0
	limit := 5

	for {
		rows, err := db.Query("SELECT announcement_id, text FROM embeddings_announcements OFFSET $1 LIMIT $2", offset, limit)
		if err != nil {
			if err == sql.ErrNoRows {
				break
			}
			return err
		}

		var id int
		var text string
		for rows.Next() {
			if err := rows.Scan(&id, &text); err != nil {
				return err
			}

			if err := InsertEmbedding(db, id, text); err != nil {
				if err == cantConnectToOllamaError || err == envOSError {
					break
				}

				db.Exec("DELETE FROM embeddings_announcements WBERE announcement_id = $1", id)
				continue
			}
		}

		offset += 5
	}

	return nil
}
