package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"net/http"
	"os"
	"time"
)

type Person struct {
	Name string
	Age  int
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	indexFile, err := os.Open("/Users/pavelryltsov/GolandProjects/OPCServer_Project1/html/index.html")
	if err != nil {
		fmt.Println(err)
	}
	index, err := io.ReadAll(indexFile)
	if err != nil {
		fmt.Println(err)
	}
	http.HandleFunc("/websocket", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Client subscribed")

		myPerson := Person{
			Name: "Bill",
			Age:  0,
		}

		for {
			time.Sleep(2 * time.Second)
			if myPerson.Age < 40 {
				myJson, err := json.Marshal(myPerson)
				if err != nil {
					fmt.Println(err)
					return
				}
				err = conn.WriteMessage(websocket.TextMessage, myJson)
				if err != nil {
					fmt.Println(err)
					break
				}
				myPerson.Age += 2
			} else {
				conn.Close()
				break
			}
		}
		fmt.Println("Client unsubscribed")

	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, string(index))
	})
	http.ListenAndServe(":3000", nil)
}
