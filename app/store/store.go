package store

import (
  "time"
  "encoding/json"
  jbolt "github.com/umputun/updater/app/store/jbolt"
)

const TOKEN_KEY = "jtrw/secret"

type Store struct {
	StorePath string
	JBolt jbolt.Bolt
}

type JSON map[string]interface{}

type Message struct {
	Key     string
	Bucket  string
	Exp     time.Time
	Type    string
	Data    string
	DataJson JSON
	//Data    []byte
	PinHash string
	Errors  int
}

func (s Store) NewStore() jbolt.Bolt {
    bolt := jbolt.Open(s.StorePath)

    return *bolt
}

func (s Store) Get(bucket, key string) string {
    return jbolt.Get(s.JBolt.DB, bucket, key)
}

func (s Store) Set(bucket, key, value string) {
    jbolt.Set(s.JBolt.DB, bucket, key, value)
}

func (s Store) Save(msg *Message) {
    jdata, jerr := json.Marshal(msg)
    if jerr != nil {
        return
        //return jerr
    }
    jbolt.Set(s.JBolt.DB, msg.Bucket, msg.Key, string(jdata))
}

func (s Store) Load(bucket, key string) (result *Message, err error) {
    val := jbolt.Get(s.JBolt.DB, bucket, key)

    result = &Message{}

    errMarshal := json.Unmarshal([]byte(val), result)

    return result, errMarshal
}

func (msg Message) ToJson() ([]byte, error) {
    return json.Marshal(msg.Data)
}
