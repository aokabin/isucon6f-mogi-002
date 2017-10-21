package mem

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/bradfitz/gomemcache/memcache"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type Token struct {
	ID        int64     `db:"id"`
	CSRFToken string    `db:"csrf_token"`
	CreatedAt time.Time `db:"created_at"`
}

type Point struct {
	ID       int64   `json:"id" db:"id"`
	StrokeID int64   `json:"stroke_id" db:"stroke_id"`
	X        float64 `json:"x" db:"x"`
	Y        float64 `json:"y" db:"y"`
}

type Stroke struct {
	ID        int64     `json:"id" db:"id"`
	RoomID    int64     `json:"room_id" db:"room_id"`
	Width     int       `json:"width" db:"width"`
	Red       int       `json:"red" db:"red"`
	Green     int       `json:"green" db:"green"`
	Blue      int       `json:"blue" db:"blue"`
	Alpha     float64   `json:"alpha" db:"alpha"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	Points    []Point   `json:"points" db:"points"`
}

type Room struct {
	ID           int64     `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	CanvasWidth  int       `json:"canvas_width" db:"canvas_width"`
	CanvasHeight int       `json:"canvas_height" db:"canvas_height"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	Strokes      []Stroke  `json:"strokes"`
	StrokeCount  int       `json:"stroke_count"`
	WatcherCount int       `json:"watcher_count"`
}

func Sample() error {
	fmt.Println("sample")

	mc := memcache.New("localhost:11211")

	_ = mc.FlushAll()

	// token := Token{
	// 	ID:        10,
	// 	CSRFToken: "hogehoge",
	// 	CreatedAt: time.Now(),
	// }

	host := os.Getenv("MYSQL_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("MYSQL_PORT")
	if port == "" {
		port = "3306"
	}
	_, err := strconv.Atoi(port)
	if err != nil {
		log.Fatalf("Failed to read DB port number from an environment variable MYSQL_PORT.\nError: %s", err.Error())
	}
	user := os.Getenv("MYSQL_USER")
	if user == "" {
		user = "isucon"
	}
	password := os.Getenv("MYSQL_PASS")
	if password == "" {
		password = "isucon"
	}
	dbname := "isuketch"

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		user,
		password,
		host,
		port,
		dbname,
	)

	dbx, err := sqlx.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %s.", err.Error())
	}
	defer dbx.Close()

	query := "SELECT `id`, `stroke_id`, `x`, `y` FROM `points` ORDER BY `id` ASC"
	points := []Point{}
	err = dbx.Select(&points, query)
	if err != nil {
		return err
	}

	for _, p := range points {
		buf := bytes.NewBuffer(nil)
		gob.NewEncoder(buf).Encode(&p)

		mc.Set(&memcache.Item{Key: "Point:" + strconv.Itoa(int(p.ID)), Value: buf.Bytes()})
	}

	it, err := mc.Get("Point:4")
	if err != nil {
		return err
	}

	pbuf := bytes.NewBuffer([]byte(it.Value))

	var point Point
	gob.NewDecoder(pbuf).Decode(&point)

	fmt.Println(point)

	return nil
}
