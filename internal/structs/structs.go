package structs

import (
	"github.com/vapourismo/knx-go/knx"
	"github.com/vapourismo/knx-go/knx/dpt"
	"sync"
)

type KNXInterface interface {
	Inbound() <-chan knx.GroupEvent
	Close()
}

type AddressTarget struct {
	Measurement string
	Datapoint   dpt.DatapointValue
}

type Measurement struct {
	Name      string   `json:"name"`      //НАИМЕНОВАНИЕ УСТРОЙСТВА
	Dpt       string   `json:"dpt"`       //ТИП ДАННЫХ
	Addresses []string `json:"addresses"` //ГРУППОВОЙ АДРЕС
}

type KNXListener struct {
	ServiceType    string        `json:"service_type"`    //ТИП ПОДКЛЮЧЕНИЯ: TUNNEL И Т.Д.
	ServiceAddress string        `json:"service_address"` //IP АДРЕС И ПОРТ
	Measurements   []Measurement `json:"measurement"`

	GaLogbook   map[string]bool
	GaTargetMap map[string]AddressTarget
	Client      KNXInterface

	msg interface{}

	Wg sync.WaitGroup
}

//type Tag struct {
//	GroupAddress string `json:"Групповой адрес"`
//	Unit         string `json:"Ед. изм."`
//	Source       string `json:"Источник"`
//}

type ListenOutput struct {
	TargetMeas   string      `json:"Наименование устройства"`
	Source       string      `json:"Источник"`
	GroupAddress string      `json:"Групповой адрес"`
	Fields       interface{} `json:"Значение"`
	Unit         string      `json:"Ед. изм."`
}
