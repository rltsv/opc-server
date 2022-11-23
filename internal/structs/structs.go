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
	Name      string   `json:"name"`
	Dpt       string   `json:"dpt"`
	Addresses []string `json:"addresses"`
}

type KNXListener struct {
	ServiceType    string        `json:"service_type"`
	ServiceAddress string        `json:"service_address"`
	Measurements   []Measurement `json:"measurement"`

	GaLogbook   map[string]bool
	GaTargetMap map[string]AddressTarget
	Client      KNXInterface

	msg interface{}

	Wg sync.WaitGroup
}
