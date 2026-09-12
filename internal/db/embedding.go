package db

import "context"

const (
	updateEmbeddingForTopPartnerOfAnnouncement = `
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
	insertCommand = "UPDATE announcements SET embedding = $1::float8[] WHERE announcement_id = $2"
)

func (s *PostgresStorage) UpdateTopPartnerEmbeddingBeforeDeleteAnnouncement(ctx context.Context, authorID, announcementID int) error {
	if _, err := s.db.ExecContext(ctx, updateEmbeddingForTopPartnerOfAnnouncement, authorID, announcementID); err != nil {
		return err
	}
	return nil
}

func (s *PostgresStorage) InsertEmbedding(ctx context.Context, rowID int, embedding []float64) error {
	_, err := s.db.Exec(insertCommand, embedding, rowID)
	if err != nil {
		return err
	}
	return nil
}

func (s *PostgresStorage) SaveEmbeddingText(ctx context.Context, rowID int, text string) error {
	if _, err := s.db.Exec("INSERT INTO embeddings_announcements (announcement_id, text) VALUES ($1, $2)", rowID, text); err != nil {
		return err
	}

	return nil
}

func (s *PostgresStorage) GetSavedEmbeddings(ctx context.Context, offset, limit int) (map[int]string, error) {
	rows, err := s.db.Query("SELECT announcement_id, text FROM embeddings_announcements OFFSET $1 LIMIT $2", offset, limit)
	if err != nil {
		return nil, err
	}

	result := make(map[int]string)
	var id int
	var text string
	for rows.Next() {
		result[id] = text
	}

	return result, err
}

func (s *PostgresStorage) DeleteSavedEmbedding(ctx context.Context, id int) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM embeddings_announcements WBERE announcement_id = $1", id)
	return err
}
