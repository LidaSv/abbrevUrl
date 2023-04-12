package storage

import (
	"encoding/json"
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

//func createFile(fileName string) (string, error) {
//	if string(fileName[0]) == "/" {
//		fileName = fileName[1:]
//	}
//	s := strings.Split(fileName, "/")
//	st := "/" + s[len(s)-1]
//	dir := strings.ReplaceAll(fileName, st, "")
//
//	if _, err := os.Stat(dir); os.IsNotExist(err) {
//		err = os.MkdirAll(dir, 0777)
//		if err != nil {
//			return "", err
//		}
//		return fileName, nil
//	}
//	return fileName, nil
//}

func NewConsumer(fileName string) (*consumer, error) {
	//fileNewName, err := createFile(fileName)
	//if err != nil {
	//	return nil, err
	//}
	fileNewName := fileName

	file, err := os.OpenFile(fileNewName, os.O_RDONLY|os.O_CREATE, 0777)
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
	//fileNewName, err := createFile(fileName)
	//if err != nil {
	//	return nil, err
	//}
	fileNewName := fileName

	file, err := os.OpenFile(fileNewName, os.O_WRONLY|os.O_CREATE, 0777)
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

func ReadCache(fileName string, st *URLStorage) {

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
}
