package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/opc-server/internal/listener"
	"github.com/opc-server/internal/structs"
	"io"
	"log"
	"net/http"
	"os"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	//ОТКРЫВАЕМ ФАЙЛ С ШАБЛОНОМ СТРАНИЦЫ
	indexFile, err := os.Open("/Users/pavelryltsov/GolandProjects/opc-server/html/index.html")
	if err != nil {
		fmt.Println(err)
	}
	//ЧИТАЕМ ШАБЛОН СТРАНИЦЫ
	index, err := io.ReadAll(indexFile)
	if err != nil {
		fmt.Println(err)
	}
	//ОБРАБОТЧИК 2. ПОДКЛЮЧЕНИЕ WEBSOCKET
	http.HandleFunc("/websocket", func(w http.ResponseWriter, r *http.Request) {
		//ПОЛУЧАЕМ ВЕБСОКЕТ ПОДКЛЮЧЕНИЕ
		Conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("Подключение создано")

		var mainStruct structs.KNXListener

		//ЧИТАЕМ ФАЙЛ С СПИСКОМ ТЕГОМ ДЛЯ ОТСЛЕЖИВАНИЯ
		file, err := os.ReadFile("/Users/pavelryltsov/GolandProjects/opc-server/initial_data.json")
		if err != nil {
			log.Fatal("Что-то пошло не так")
		}

		//АНМАРШАЛИМ ФАЙЛ В СТРУКТУРУ
		err = json.Unmarshal(file, &mainStruct)
		if err != nil {
			log.Fatal("Анмаршал не прошел")
		}

		//СОСТАВЛЯЕМ МАПУ ДЛЯ КОНТРОЛЯ ЛЕТЯЩИХ СООБЩЕНИЙ ИЗ КАНАЛА ШЛЮЗА
		err = listener.Start(&mainStruct)
		if err != nil {
			log.Fatal("Ошибка при составлении мапы")
		}
		fmt.Print(mainStruct.ServiceType, "\n", mainStruct.ServiceAddress, "\n", mainStruct.GaTargetMap, "\n")

		//СОЗДАЕМ ПОДКЛЮЧЕНИЕ С ШЛЮЗОМ
		//TODO: Вынести в отдельную функцию
		err = listener.Connect(&mainStruct)
		if err != nil {
			fmt.Print("Подключения нет! ", err)
		}
		fmt.Print("Связь установлена!\n")

		//СОЗДАЕМ ГОУРУТИНУ ДЛЯ ЧТЕНИЯ СООБЩЕНИЙ ИЗ КАНАЛ И ОТПРАВКИ В ВЕБСОКЕТ ПОДКЛЮЧЕНИЕ НА ФРОНТ
		mainStruct.Wg.Add(1)
		go func() {
			listener.Listen(&mainStruct, Conn)
		}()
		mainStruct.Wg.Wait()

	})
	//ОБРАБОТЧИК 2. ВЫДАЕМ ГЛАВНУЮ СТРАНИЦУ
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, string(index))
	})

	//СЛУШАЕМ ПОРТ
	log.Fatal(http.ListenAndServe(":3000", nil))

}
