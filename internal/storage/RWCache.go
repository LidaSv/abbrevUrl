package storage

import (
	"encoding/json"
	"github.com/caarlos0/env/v6"
	"github.com/spf13/pflag"
	"io"
	"log"
	"os"
)

type strucRW struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type consumer struct {
	file    *os.File
	decoder *json.Decoder
}

func NewConsumer(filename string) (*consumer, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		return nil, err
	}

	return &consumer{
		file:    file,
		decoder: json.NewDecoder(file),
	}, nil
}

func (c *consumer) ReadEvent() (*strucRW, error) {
	event := &strucRW{}
	if err := c.decoder.Decode(&event); err != nil {
		return nil, err
	}
	return event, nil
}

func (c *consumer) Close() error {
	return c.file.Close()
}

type producer struct {
	file    *os.File
	encoder *json.Encoder
}

func NewProducer(fileName string) (*producer, error) {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		return nil, err
	}
	return &producer{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}
func (p *producer) WriteEvent(cache *strucRW) error {
	return p.encoder.Encode(&cache)
}
func (p *producer) Close() error {
	return p.file.Close()
}

func WriterCache(fileName string, st *URLStorage) {

	producer, err := NewProducer(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer producer.Close()

	cache := st.Urls
	for r, cc := range cache {
		l := []*strucRW{
			{
				Key:   cc,
				Value: r,
			},
		}
		if err := producer.WriteEvent(l[0]); err != nil {
			log.Fatal(err)
		}
	}
}

func ReadCache(st *URLStorage) string {

	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	flag := pflag.FlagSet{}
	FlagFileStoragePath := flag.String("f", cfg.FileStoragePath, "a string")
	flag.Parse(os.Args[1:])

	filepath, exist := os.LookupEnv("FILE_STORAGE_PATH")

	var fileName string
	if exist {
		fileName = filepath
	} else {
		fileName = *FlagFileStoragePath
	}

	consumer, err := NewConsumer(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer consumer.Close()
	for {
		readEvent, err := consumer.ReadEvent()
		if err == io.EOF {
			break
		}
		st.mutex.RLock()
		st.Urls[readEvent.Key] = readEvent.Value
		st.mutex.RUnlock()
	}

	return fileName
}
