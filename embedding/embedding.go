package embedding

import (
	"bytes"
	"database/sql"
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

type Response struct {
	Embedding []float64
}

var ollamaHost = os.Getenv("OLLAMA_HOST")

func InsertEmbedding(db *sql.DB, rowID int, text, insertCommand string) error {
	jsonBody := []byte(`{"model":  "nomic-embed-text", "prompt": "` + text + `"}`)

	resp, err := client.Post(ollamaHost+"/api/embeddings", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}

	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	var result Response
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
