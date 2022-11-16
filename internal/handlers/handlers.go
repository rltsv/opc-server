package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
	"time"
)

var upgrader = websocket.Upgrader{}

func H1(w http.ResponseWriter, r *http.Request) {
	//Апгрейдим подключение до вебсокета
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println("Начало сеанса связи!")

	type Person struct {
		Name string
		Age  int
	}

	newPerson := Person{
		Name: "Jack",
		Age:  10,
	}

	for {
		time.Sleep(2 * time.Second)
		if newPerson.Age < 40 {
			myjson, err := json.Marshal(newPerson)
			if err != nil {
				log.Println("Ошибка маршала")
				return
			}
			err = conn.WriteMessage(1, myjson)
			if err != nil {
				log.Println("Ошибка вывода")
			}
			newPerson.Age = +1
		} else {
			conn.WriteMessage(1, []byte("Сейчас соединение закроется"))
			conn.Close()
			break
		}

	}

}

func H2(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hello from a HandleFunc #2!\n")
}
