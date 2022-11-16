package listener

import (
	"encoding/json"
	"fmt"
	"github.com/vapourismo/knx-go/knx"
	"github.com/vapourismo/knx-go/knx/dpt"
	"log"
	"os"
	"reflect"
	"sync"
)

type KNXInterface interface {
	Inbound() <-chan knx.GroupEvent
	Close()
}

type addressTarget struct {
	measurement string
	datapoint   dpt.DatapointValue
}

type Measurement struct {
	Name      string   `json:"name"`
	Dpt       string   `json:"dpt"`
	Addresses []string `json:"addresses"`
}

type KNXListener struct {
	ServiceType    string        `json:"service_type"`
	ServiceAddress string        `json:"service_address"`
	Measurements   []Measurement `json:"measurement"`

	gaLogbook   map[string]bool
	gaTargetMap map[string]addressTarget
	client      KNXInterface

	msg interface{}

	wg sync.WaitGroup
}

func Start(ms *KNXListener) error {

	ms.gaTargetMap = make(map[string]addressTarget)

	for _, m := range ms.Measurements {

		for _, ga := range m.Addresses {

			if _, ok := ms.gaTargetMap[ga]; ok {
				return fmt.Errorf("duplicate specification of address %q", ga)
			}
			d, ok := dpt.Produce(m.Dpt)
			if !ok {
				return fmt.Errorf("cannot create datapoint-type %q for address %q", m.Dpt, ga)
			}
			ms.gaTargetMap[ga] = addressTarget{m.Name, d}
		}
	}
	return nil
}

func Connect(ms *KNXListener) error {
	c, err := knx.NewGroupTunnel(ms.ServiceAddress, knx.DefaultTunnelConfig)
	if err != nil {
		return err
	}
	ms.client = &c
	return nil
}

func Listen(ms *KNXListener) {
	ms.gaLogbook = make(map[string]bool)

	for msg := range ms.client.Inbound() {
		// Match GA to DataPointType and measurement name
		ga := msg.Destination.String()
		target, ok := ms.gaTargetMap[ga]
		if !ok {
			if !ms.gaLogbook[ga] {

				ms.gaLogbook[ga] = true
			}
			continue
		}
		//ИЗВЛЕКАЕМ ЗНАЧЕНИЯ ИЗ ПАКЕТА
		err := target.datapoint.Unpack(msg.Data)
		if err != nil {
			//fmt.Printf("Unpacking data failed: %v\n", err)
			continue
		}
		//fmt.Printf("Matched GA %q to measurement %q with value %v\n", ga, target.measurement, target.datapoint)

		// Convert the DatapointValue interface back to its basic type again
		// as otherwise telegraf will not push out the metrics and eat it
		// silently.
		var value interface{}
		vi := reflect.Indirect(reflect.ValueOf(target.datapoint))
		switch vi.Kind() {
		case reflect.Bool:
			value = vi.Bool()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			value = vi.Int()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			value = vi.Uint()
		case reflect.Float32, reflect.Float64:
			value = vi.Float()
		default:
			fmt.Printf("Type conversion %v failed for address %q", vi.Kind(), ga)
			continue
		}

		// Compose the actual data to be pushed out
		fields := map[string]interface{}{"value": value}
		tags := map[string]string{
			"groupAddress": ga,
			"unit":         target.datapoint.(dpt.DatapointMeta).Unit(),
			"source":       msg.Source.String(),
		}
		fmt.Println(target.measurement, fields, tags)
	}
}

// ЗАПУСК ПРОГРАММЫ
func main() {

	var mainStruct KNXListener
	//ЧИТАЕМ JSON ИЗ ФАЙЛА
	file, err := os.ReadFile("initial_data.json")
	if err != nil {
		log.Fatal("Что-то пошло не так")
	}

	err = json.Unmarshal(file, &mainStruct)
	if err != nil {
		log.Fatal("Анмаршал не прошел")
	}
	// СОСТАВЛЯЕМ МАПУ
	err = Start(&mainStruct)
	if err != nil {
		log.Fatal("Ошибка при составлении мапы")
	}
	fmt.Print(mainStruct.ServiceType, "\n", mainStruct.ServiceAddress, "\n", mainStruct.gaTargetMap, "\n")

	//ПОДКЛЮЧАЕМСЯ К ШЛЮЗУ
	err = Connect(&mainStruct)
	if err != nil {
		fmt.Print("Подключения нет! ", err)
	}
	fmt.Print("Связь установлена!\n")

	//НАЧИНАЕМ ЧИТАТЬ
	mainStruct.wg.Add(1)
	go func() {
		Listen(&mainStruct)
	}()
	mainStruct.wg.Wait()

}
