package embedding

import (
	"announcements/internal/db"
	"bytes"
	"context"
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

func InsertEmbedding(ctx context.Context, storage db.AnnouncementsStorage, rowID int, text string) error {
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
		storage.SaveEmbeddingText(ctx, rowID, text)

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

	err = storage.InsertEmbedding(ctx, rowID, result.Embedding)
	if err != nil {
		return err
	}

	return nil
}

func RetryInsertEmbeddings(storage db.AnnouncementsStorage) {
	for {
		runInsertSavedEmbeddings(storage)
		time.Sleep(time.Hour * 4)
	}
}

func runInsertSavedEmbeddings(storage db.AnnouncementsStorage) error {
	offset := 0
	limit := 5

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
	defer cancel()

	for {
		embeddings, err := storage.GetSavedEmbeddings(ctx, offset, limit)
		if err != nil {
			if err == sql.ErrNoRows {
				break
			}
			return err
		}

		for id, text := range embeddings {
			if err := InsertEmbedding(ctx, storage, id, text); err != nil {
				if err == cantConnectToOllamaError || err == envOSError {
					break
				}

				storage.DeleteSavedEmbedding(ctx, id)
				continue
			}
		}

		offset += 5
	}

	return nil
}
