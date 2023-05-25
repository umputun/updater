package store

import (
 // "fmt"
  "log"
  "time"
  "github.com/nilBora/bolt"
)

type Bolt struct {
	DB *bolt.DB
}

func Open(file string) *Bolt {
  db, err := bolt.Open(file, 0600, &bolt.Options{Timeout: 1 * time.Second})
  if err != nil {
    //handle error
    log.Fatal(err)
  }
  var boltDB *Bolt
  boltDB = new(Bolt)
  boltDB.DB = db

  return boltDB
}

func Set(db *bolt.DB, bucket, key, value string) {
  db.Update(func(tx *bolt.Tx) error {
    b, _ := tx.CreateBucketIfNotExists([]byte(bucket))
    err := b.Put([]byte(key), []byte(value))
    return err
  })
}

func Get(db *bolt.DB, bucket, key string) string {
  val := ""
  db.View(func(tx *bolt.Tx) error {
    bucket := tx.Bucket([]byte(bucket))
    v := bucket.Get([]byte(key))
    if v == nil {
        log.Printf("[INFO] not found %s", key)
        return nil
    }
    val = string(v)
    return nil
  })
  return val
}

func Del(db *bolt.DB, bucket, key string) {
  db.Update(func(tx *bolt.Tx) error {
    b := tx.Bucket([]byte(bucket))
    b.Delete([]byte(key))
    return nil
  })
}