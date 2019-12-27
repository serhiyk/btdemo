package btdemo

/*
#cgo CFLAGS: -I .
#cgo LDFLAGS: -L . -lbt

#include <stdlib.h>
#include <stdint.h>
#include <stdbool.h>
#include "wpc.h"

bool onDataReceived_cgo(uint8_t *data,
                        uint8_t num_bytes,
                        uint32_t src_addr,
                        uint32_t dst_addr,
                        unsigned int qos,
                        uint8_t src_ep,
                        uint8_t dst_ep,
                        uint32_t travel_time,
                        uint8_t hop_count,
                        unsigned long long timestamp_ms_epoch);
*/
import "C"
import (
	"context"
	"fmt"
	"time"
	"unsafe"
	"math/rand"
	"reflect"
	"bytes"
    "encoding/binary"

	// "github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/core/trigger"
)

var g_handlers []trigger.Handler

type heartbeat_t struct {
	opcode uint8
	euid [6]uint8
	role uint8
	state uint8
	addr uint32
	session uint32
}

var session uint32
var pdu_id C.uint16_t = 0

func send_data(data *C.uint8_t, num_bytes C.uint8_t, addr C.uint32_t) int {
	pdu_id = (pdu_id + 1) % 0xffff
	return int(C.WPC_send_data(data, num_bytes, pdu_id, addr, 1, 50, 50, nil, 0))
}

func send_heartbeat() {
	heartbeat := heartbeat_t{session: session}
	buf := &bytes.Buffer{}
	binary.Write(buf, binary.LittleEndian, heartbeat)
	send_data((*C.uint8_t)(unsafe.Pointer(&buf.Bytes()[0])), C.uint8_t(buf.Len()), 0xffffffff)
	fmt.Println("heartbeat sent", buf.Bytes())
}

func on_new_node(euid [6]uint8, addr uint32) {
	out := &Output{
		euid: euid,
		addr: addr,
	}
	handlers := g_handlers
	for _, handler := range handlers {
		_, err := handler.Handle(context.Background(), out)
		if err != nil {
			fmt.Println("Error running handler: ", err.Error())
		}
	}
}

//export onDataReceived
func onDataReceived(data *C.uint8_t, num_bytes C.uint8_t, src_addr C.uint32_t, dst_addr C.uint32_t, qos C.uint, src_ep C.uint8_t, dst_ep C.uint8_t, travel_time C.uint32_t, hop_count C.uint8_t, timestamp_ms_epoch C.ulonglong) C.bool {
	var msg []uint8
    sliceHeader := (*reflect.SliceHeader)((unsafe.Pointer(&msg)))
    sliceHeader.Cap = int(num_bytes)
    sliceHeader.Len = int(num_bytes)
    sliceHeader.Data = uintptr(unsafe.Pointer(data))
	fmt.Printf("onDataReceived = %d\n", num_bytes)
	// fmt.Println("data ", msg)
	if (msg[0] == 0) {
		var node_session uint32 = uint32(msg[13]) + (uint32(msg[14]) << 8) + (uint32(msg[15]) << 16) + (uint32(msg[16]) << 24)
		if (node_session != session) {
			fmt.Printf("session %X != %X\n", node_session, session)
			send_heartbeat()
		} else if (msg[8] == 0) {
			var node_addr uint32 = uint32(msg[9]) + (uint32(msg[10]) << 8) + (uint32(msg[11]) << 16) + (uint32(msg[12]) << 24)
			node_euid := [6]uint8{uint8(msg[1]), uint8(msg[2]), uint8(msg[3]), uint8(msg[4]), uint8(msg[5]), uint8(msg[6])}
			fmt.Println("node is joining", node_euid, node_addr)
			send_data(data, num_bytes, src_addr)
			on_new_node(node_euid, node_addr)
		}
	}
	return true
}

type HandlerSettings struct {
	port  string `md:"bt port"`
}

type Output struct {
	euid [6]uint8 `md:"euid"`
	addr uint32 `md:"addr"`
}

var triggerMd = trigger.NewMetadata(&HandlerSettings{}, &Output{})

func init() {
	_ = trigger.Register(&Trigger{}, &Factory{})
}

type Factory struct {
}

// Metadata implements trigger.Factory.Metadata
func (*Factory) Metadata() *trigger.Metadata {
	return triggerMd
}

// New implements trigger.Factory.New
func (*Factory) New(config *trigger.Config) (trigger.Trigger, error) {
	return &Trigger{}, nil
}

type Trigger struct {
	handlers []trigger.Handler
	logger   log.Logger
}

// Init implements trigger.Init
func (t *Trigger) Initialize(ctx trigger.InitContext) error {
	t.handlers = ctx.GetHandlers()
	t.logger = ctx.Logger()

	return nil
}

// Start implements ext.Trigger.Start
func (t *Trigger) Start() error {
	g_handlers = t.handlers

	rand.Seed(time.Now().UTC().UnixNano())
	session = uint32(rand.Intn(0xfffffffe) + 1)
	cs := C.CString("/dev/ttyACM1")
    defer C.free(unsafe.Pointer(cs))
	C.WPC_initialize(cs, 115200)
	C.WPC_stop_stack()
    time.Sleep(2 * time.Second)
    C.WPC_register_for_data(50, (C.onDataReceived_cb_f)(unsafe.Pointer(C.onDataReceived_cgo)))
    C.WPC_set_node_address(0x123456)
    C.WPC_set_network_channel(2)
    C.WPC_set_role(0x10 + 1)
    C.WPC_set_network_address(0xABCDEF)
    C.WPC_start_stack()

	// handlers := t.handlers
	// for _, handler := range handlers {
	// }

	return nil
}

// Stop implements ext.Trigger.Stop
func (t *Trigger) Stop() error {

	C.WPC_close()

	return nil
}
