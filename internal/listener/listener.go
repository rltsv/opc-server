package listener

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/opc-server/internal/structs"
	"github.com/vapourismo/knx-go/knx"
	"github.com/vapourismo/knx-go/knx/dpt"
	"reflect"
)

func Start(ms *structs.KNXListener) error {

	ms.GaTargetMap = make(map[string]structs.AddressTarget)

	for _, m := range ms.Measurements {

		for _, ga := range m.Addresses {

			if _, ok := ms.GaTargetMap[ga]; ok {
				return fmt.Errorf("duplicate specification of address %q", ga)
			}
			d, ok := dpt.Produce(m.Dpt)
			if !ok {
				return fmt.Errorf("cannot create datapoint-type %q for address %q", m.Dpt, ga)
			}
			ms.GaTargetMap[ga] = structs.AddressTarget{m.Name, d}
		}
	}
	return nil
}

func Connect(ms *structs.KNXListener) error {
	c, err := knx.NewGroupTunnel(ms.ServiceAddress, knx.DefaultTunnelConfig)
	if err != nil {
		return err
	}
	ms.Client = &c
	return nil
}

func Listen(ms *structs.KNXListener, Conn *websocket.Conn) {
	ms.GaLogbook = make(map[string]bool)

	for msg := range ms.Client.Inbound() {
		// Match GA to DataPointType and measurement name
		ga := msg.Destination.String()
		target, ok := ms.GaTargetMap[ga]
		if !ok {
			if !ms.GaLogbook[ga] {

				ms.GaLogbook[ga] = true
			}
			continue
		}
		//ИЗВЛЕКАЕМ ЗНАЧЕНИЯ ИЗ ПАКЕТА
		err := target.Datapoint.Unpack(msg.Data)
		if err != nil {
			//fmt.Printf("Unpacking data failed: %v\n", err)
			continue
		}
		//fmt.Printf("Matched GA %q to measurement %q with value %v\n", ga, target.measurement, target.datapoint)

		//Convert the DatapointValue interface back to its basic type again
		//as otherwise telegraf will not push out the metrics and eat it
		//silently.
		var value interface{}
		vi := reflect.Indirect(reflect.ValueOf(target.Datapoint))
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

		//Compose the actual data to be pushed out
		//fields := map[string]interface{}{"value": value}
		//tags := map[string]string{
		//	"groupAddress": ga,
		//	"unit":         target.Datapoint.(dpt.DatapointMeta).Unit(),
		//	"source":       msg.Source.String(),
		//}

		OutputStruct := structs.ListenOutput{
			TargetMeas:   target.Measurement,
			Fields:       value,
			GroupAddress: ga,
			Unit:         target.Datapoint.(dpt.DatapointMeta).Unit(),
			Source:       msg.Source.String(),
		}

		//fmt.Println(target.Measurement, fields, tags)

		myJson, err := json.Marshal(OutputStruct)
		if err != nil {
			fmt.Println(err)
			return
		}

		var oldJson []byte

		if !bytes.Equal(oldJson, myJson) {
			err = Conn.WriteMessage(websocket.TextMessage, myJson)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
		oldJson = myJson
	}
}
