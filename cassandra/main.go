package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gocql/gocql"
)

func main() {
	cluster := gocql.NewCluster("127.0.0.1:9042")
	cluster.Keyspace = "store"
	session, err := cluster.CreateSession()
	if err != nil {
		panic(err)
	}
	defer session.Close()
	scanner := session.Query("SELECT * FROM store.shopping_cart WHERE userid = '9876'").Iter().Scanner()
	for scanner.Next() {
		var (
			id         string
			count      int
			lastUpdate time.Time
		)
		err = scanner.Scan(&id, &count, &lastUpdate)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(fmt.Sprintf("id: %s, count: %d, lastUpdate: %v", id, count, lastUpdate))
	}
}
