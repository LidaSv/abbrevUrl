package storage

import (
	"context"
	"github.com/jackc/pgx/v5"
	"log"
	"strings"
	"time"
)

type structBD struct {
	LongURL string
	ID      string
}

func ReadDBCashe(DatabaseDsn string, st *URLStorage) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := pgx.Connect(ctx, DatabaseDsn)
	if err != nil {
		log.Println("Unable to connect to database:", err)
		return
	}
	defer db.Close(context.Background())

	_, err = db.Exec(ctx,
		`create table if not exists long_short_urls (
				long_url varchar(256),
				short_url varchar(256),
				id_short_url varchar(32)
				);`)
	if err != nil {
		log.Fatal("create: ", err)
	}

	row, err := db.Query(ctx, "SELECT long_url, id_short_url FROM long_short_urls")
	if err != nil {
		log.Println("select: ", err)
	}
	defer row.Close()

	for row.Next() {
		var v structBD
		err = row.Scan(&v.LongURL, &v.ID)
		if err != nil {
			log.Println("scan:", err)
		}
		st.mutex.RLock()
		st.Urls[v.LongURL] = v.ID
		st.Urls[v.ID] = v.LongURL
		st.mutex.RUnlock()
	}

	err = row.Err()
	if err != nil {
		log.Println("Err: ", err)
	}
}

func WriteDBCashe(DatabaseDsn string, st *URLStorage) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := pgx.Connect(ctx, DatabaseDsn)
	if err != nil {
		log.Println("Unable to connect to database:", err)
		return
	}
	defer db.Close(context.Background())

	_, err = db.Exec(ctx,
		`truncate table long_short_urls;`)
	if err != nil {
		log.Fatal("truncate: ", err)
	}

	cache := st.Urls
	for r, cc := range cache {
		if strings.HasPrefix(r, "https://") {
			_, err := db.Exec(ctx,
				`insert into long_short_urls 
				select 
				    $1 long_url,
					$2 short_url,
					$3 id_short_url
				;`, r, st.BaseURL+"/"+cc, cc)
			if err != nil {
				log.Fatal("create: ", err)
			}
		}
	}
}
