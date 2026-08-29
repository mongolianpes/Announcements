package embedding

import (
	"database/sql"
	"fmt"
	"math"
	"strconv"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

type updMessagesInfo struct {
	usersIDs         []int
	announcementsIDs []int
}

func connectToDBForTest() *sql.DB {
	host := "localhost"
	port := "5432"
	user := "postgres"
	password := "123"
	dbname := "project_farm"

	// host := os.Getenv("DB_HOST")
	// port := os.Getenv("DB_PORT")
	// user := os.Getenv("DB_USER")
	// password := os.Getenv("DB_PASSWORD")
	// dbname := os.Getenv("DB_NAME")

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

func updateTestUsers(db *sql.DB) (updMessagesInfo, error) {
	updMessages := updMessagesInfo{}
	if _, err := db.Exec("DELETE FROM users"); err != nil {
		return updMessages, err
	}

	rowsUserIDs, err := db.Query(`INSERT INTO users (login, name, password) VALUES
		(1, '1', '1234567890'),
		(2, '2', '1234567890'),
		(3, '3', '1234567890')
		RETURNING user_id`)
	if err != nil {
		return updMessages, err
	}

	var userID1 int
	var userID2 int
	var userID3 int
	rowsUserIDs.Next()
	rowsUserIDs.Scan(&userID1)
	rowsUserIDs.Next()
	rowsUserIDs.Scan(&userID2)
	rowsUserIDs.Next()
	rowsUserIDs.Scan(&userID3)
	updMessages.usersIDs = append(updMessages.usersIDs, userID1)
	updMessages.usersIDs = append(updMessages.usersIDs, userID2)
	updMessages.usersIDs = append(updMessages.usersIDs, userID3)
	rowsUserIDs.Close()

	return updMessages, err
}

func updateTestAnnouncements(db *sql.DB, updMessages updMessagesInfo) (updMessagesInfo, error) {
	rowsAnnouncementIDs, err := db.Query(`INSERT INTO announcements (title, description, category, announcement_author_id) VALUES
	('1', '1', 'Участки', $1),
	('2', '2', 'Участки', $2),
	('22', '22', 'Участки', $2),
	('222', '222', 'Участки', $2)
	RETURNING announcement_id`, updMessages.usersIDs[0], updMessages.usersIDs[1])
	if err != nil {
		return updMessages, err
	}

	var announcementID1 int
	var announcementID2 int
	var announcementID3 int
	var announcementID4 int
	rowsAnnouncementIDs.Next()
	rowsAnnouncementIDs.Scan(&announcementID1)
	rowsAnnouncementIDs.Next()
	rowsAnnouncementIDs.Scan(&announcementID2)
	rowsAnnouncementIDs.Next()
	rowsAnnouncementIDs.Scan(&announcementID3)
	rowsAnnouncementIDs.Next()
	rowsAnnouncementIDs.Scan(&announcementID4)
	updMessages.announcementsIDs = append(updMessages.announcementsIDs, announcementID1)
	updMessages.announcementsIDs = append(updMessages.announcementsIDs, announcementID2)
	updMessages.announcementsIDs = append(updMessages.announcementsIDs, announcementID3)
	updMessages.announcementsIDs = append(updMessages.announcementsIDs, announcementID4)
	rowsAnnouncementIDs.Close()

	return updMessages, nil
}

func updateTestMessages(db *sql.DB) (updMessagesInfo, error) {
	db.Exec("DELETE FROM messages")

	updMessages, err := updateTestUsers(db)
	if err != nil {
		return updMessages, err
	}

	updMessages, err = updateTestAnnouncements(db, updMessages)
	if err != nil {
		return updMessages, err
	}

	messages := []struct {
		senderID, receivedID, announcementID int
		text                                 string
	}{
		{updMessages.usersIDs[0], updMessages.usersIDs[1], updMessages.announcementsIDs[0], "Привет по объявлению 10"},
		{updMessages.usersIDs[1], updMessages.usersIDs[0], updMessages.announcementsIDs[0], "Ответ по объявлению 10"},
		{updMessages.usersIDs[0], updMessages.usersIDs[1], updMessages.announcementsIDs[0], "Привет по объявлению 10 (2)"},
		{updMessages.usersIDs[0], updMessages.usersIDs[1], updMessages.announcementsIDs[0], "Привет по объявлению 10 (3)"},
		{updMessages.usersIDs[1], updMessages.usersIDs[0], updMessages.announcementsIDs[0], "Ответ по объявлению 10 (2)"},
		{updMessages.usersIDs[1], updMessages.usersIDs[0], updMessages.announcementsIDs[0], "Ответ по объявлению 10 (3)"},
		{updMessages.usersIDs[0], updMessages.usersIDs[1], updMessages.announcementsIDs[0], "Привет по объявлению 20"},
		{updMessages.usersIDs[0], updMessages.usersIDs[1], updMessages.announcementsIDs[0], "Прив по 3 для 2"},
		{updMessages.usersIDs[0], updMessages.usersIDs[2], updMessages.announcementsIDs[0], "Прив по 3 для 3"},
		{updMessages.usersIDs[0], updMessages.usersIDs[2], updMessages.announcementsIDs[0], "Прив по 4 для 3"},
		{updMessages.usersIDs[0], updMessages.usersIDs[2], updMessages.announcementsIDs[0], "Сообщение пользователю 3 по объявлению 10"},
		{updMessages.usersIDs[2], updMessages.usersIDs[0], updMessages.announcementsIDs[0], "Ответ пользователю 1 по объявлению 10"},
	}

	for _, m := range messages {
		if _, err := db.Exec(`INSERT INTO messages (sender_id, received_id, related_announcement_id, message)
							VALUES ($1, $2, $3, $4)`,
			m.senderID, m.receivedID, m.announcementID, m.text); err != nil {
			return updMessages, err
		}
		time.Sleep(50 * time.Millisecond) // задержка между вставками
	}

	return updMessages, nil
}

func TestDeleteAnnouncement(t *testing.T) {
	db := connectToDBForTest()

	upd, err := updateTestMessages(db)
	if err != nil {
		t.Fatal(err)
	}

	vector := make([]float64, 768)
	for i := range vector {
		vector[i] = 0.11111111
	}
	if _, err := db.Exec("UPDATE users SET embedding = $1::float8[]", vector); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("UPDATE announcements SET embedding = $1::float8[]", vector); err != nil {
		t.Fatal(err)
	}

	var userEmbeddingOld []float64
	var raw sql.NullString
	err = db.QueryRow("SELECT embedding FROM users WHERE user_id = $1", upd.usersIDs[0]).Scan(&raw)
	if err != nil {
		t.Fatal(err)
	}

	if raw.Valid {
		parts := strings.Trim(raw.String, "[]")
		nums := strings.Split(parts, ",")
		userEmbeddingOld = make([]float64, len(nums))
		for i, s := range nums {
			userEmbeddingOld[i], _ = strconv.ParseFloat(strings.TrimSpace(s), 64)
		}
	} else {
		t.Fatal("Вернул невалид эмбеддинг")
	}

	var announcementEmbedding []float64
	err = db.QueryRow("SELECT embedding FROM announcements WHERE announcement_id = $1", upd.announcementsIDs[0]).Scan(&raw)
	if err != nil {
		t.Fatal(err)
	}

	if raw.Valid {
		parts := strings.Trim(raw.String, "[]")
		nums := strings.Split(parts, ",")
		announcementEmbedding = make([]float64, len(nums))
		for i, s := range nums {
			announcementEmbedding[i], _ = strconv.ParseFloat(strings.TrimSpace(s), 64)
		}
	} else {
		t.Fatal("Вернул невалид эмбеддинг")
	}

	if err := UpdateUserEmbeddingAfterDeleteAnnouncement(db, int32(upd.usersIDs[0]), int32(upd.announcementsIDs[0])); err != nil {
		t.Error(err)
	}

	var userEmbeddingNew []float64
	err = db.QueryRow("SELECT embedding FROM users WHERE user_id = $1", upd.usersIDs[0]).Scan(&raw)
	if err != nil {
		t.Fatal(err)
	}

	if raw.Valid {
		parts := strings.Trim(raw.String, "[]")
		nums := strings.Split(parts, ",")
		userEmbeddingNew = make([]float64, len(nums))
		for i, s := range nums {
			userEmbeddingNew[i], _ = strconv.ParseFloat(strings.TrimSpace(s), 64)
		}
	} else {
		t.Fatal("Вернул невалид эмбеддинг")
	}

	const eps = 1e-6
	for i := range userEmbeddingOld {
		userEmbeddingOld[i] = float64(0.9)*(userEmbeddingOld)[i] + float64(0.1)*announcementEmbedding[i]

		if math.Abs(userEmbeddingOld[i]-userEmbeddingNew[i]) > eps {
			t.Errorf("Не соответствует ожидаемому. Должно быть: %v, получено: %v", userEmbeddingOld[i], userEmbeddingNew[i])
			break
		}
	}

	if rows, err := db.Query("SELECT announcement_id FROM announcements WHERE announcement_id = $1", upd.announcementsIDs[0]); err != sql.ErrNoRows && rows.Next() {
		t.Error(err)
	}

	if err := UpdateUserEmbeddingAfterDeleteAnnouncement(db, int32(upd.usersIDs[1]), int32(upd.announcementsIDs[1])); err != nil {
		t.Error(err)
	}
}
