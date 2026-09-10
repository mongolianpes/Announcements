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

const updateEmbeddingForTopPartnerOfAnnouncement = `
WITH top_partner AS (
    SELECT CASE WHEN sender_id = $1 THEN received_id ELSE sender_id END AS id
    FROM messages
    WHERE related_announcement_id = $2
    GROUP BY id
    ORDER BY COUNT(*) DESC
    LIMIT 1
),
announcement AS (
    SELECT announcement_author_id, embedding
    FROM announcements
    WHERE announcement_id = $2
)
UPDATE users u
SET embedding = (
    u.embedding * array_fill(0.9::real, ARRAY[vector_dims(u.embedding)])::vector
  + a.embedding * array_fill(0.1::real, ARRAY[vector_dims(a.embedding)])::vector
)
FROM announcement a, top_partner tp
WHERE u.user_id = tp.id;`

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

func UpdateUserEmbeddingAfterDeleteAnnouncement(db *sql.DB, authorID, announcementID int64) error {
	if _, err := db.Exec(updateEmbeddingForTopPartnerOfAnnouncement, authorID, announcementID); err != nil {
		return err
	}

	if _, err := db.Exec("DELETE FROM announcements WHERE announcement_id = $1", announcementID); err != nil {
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
