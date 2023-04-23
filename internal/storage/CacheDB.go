package storage

import (
	"context"
	"github.com/jackc/pgx/v5"
	"log"
	"time"
)

type structDB struct {
	LongURL string
	ID      string
}

func ReadDBCashe(DatabaseDsn string, st *URLStorage) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := pgx.Connect(ctx, DatabaseDsn)
	if err != nil {
		log.Fatal("Unable to connect to database:", err)
		return
	}
	defer db.Close(context.Background())

	_, err = db.Exec(ctx,
		`create table if not exists long_short_urls (
    			id_long_url serial,
				long_url varchar(256),
				short_url varchar(256),
				id_short_url varchar(32)
				);`)
	if err != nil {
		log.Fatal("create: ", err)
	}

	row, err := db.Query(ctx, "SELECT long_url, id_short_url FROM long_short_urls")
	if err != nil {
		log.Fatal("select: ", err)
	}
	defer row.Close()

	for row.Next() {
		var v structDB
		err = row.Scan(&v.LongURL, &v.ID)
		if err != nil {
			log.Fatal("scan:", err)
		}
		st.mutex.RLock()
		st.Urls[v.LongURL] = v.ID
		st.Urls[v.ID] = v.LongURL
		st.mutex.RUnlock()
	}

	err = row.Err()
	if err != nil {
		log.Fatal("Err: ", err)
	}
}
