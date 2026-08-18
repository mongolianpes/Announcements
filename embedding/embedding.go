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

func InsertEmbedding(db *sql.DB, rowID int, text, insertCommand string) error {
	if ollamaHost == "" {
		return errors.New("Переменная OLLAMA_HOST должна иметь значение: адрес локальной нейросети ollama")
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
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New("Ollama вернул статус код не 200")
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
