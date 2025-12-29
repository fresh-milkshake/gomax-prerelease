package main

/*
#include <stdlib.h>
#include <stdint.h>

typedef void (*gomax_on_message_cb)(const char* message_json, void* user_data);

static void gomax_call_on_message(gomax_on_message_cb cb, const char* message_json, void* user_data) {
	if (cb != NULL) {
		cb(message_json, user_data);
	}
}
*/
import "C"

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime/cgo"
	"time"
	"unsafe"

	gm "github.com/fresh-milkshake/gomax"
	"github.com/fresh-milkshake/gomax/types"
)

type clientHandle struct {
	client    *gm.MaxClient
	onMessage C.gomax_on_message_cb
	userData  unsafe.Pointer
}

func setError(errOut **C.char, err error) {
	if errOut == nil {
		return
	}
	if err != nil {
		*errOut = C.CString(err.Error())
	} else {
		*errOut = nil
	}
}

func getHandle(h C.uintptr_t) (ch *clientHandle, ok bool) {
	if h == 0 {
		return nil, false
	}
	defer func() {
		if recover() != nil {
			// Handle was already deleted or invalid.
			ch, ok = nil, false
		}
	}()
	v := cgo.Handle(h).Value()
	ch, ok = v.(*clientHandle)
	return
}

//export gomax_new_client
func gomax_new_client(configJSON *C.char, errOut **C.char) C.uintptr_t {
	setError(errOut, nil)
	if configJSON == nil {
		setError(errOut, fmt.Errorf("config_json is nil"))
		return 0
	}

	var cfg gm.ClientConfig
	if err := json.Unmarshal([]byte(C.GoString(configJSON)), &cfg); err != nil {
		setError(errOut, err)
		return 0
	}

	client, err := gm.NewMaxClient(cfg)
	if err != nil {
		setError(errOut, err)
		return 0
	}

	handle := cgo.NewHandle(&clientHandle{client: client})
	return C.uintptr_t(handle)
}

//export gomax_start
func gomax_start(h C.uintptr_t, timeoutMs C.int, errOut **C.char) C.int {
	setError(errOut, nil)
	ch, ok := getHandle(h)
	if !ok {
		setError(errOut, fmt.Errorf("invalid handle"))
		return 0
	}

	ctx := context.Background()
	if timeoutMs > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
		defer cancel()
	}

	if err := ch.client.Start(ctx); err != nil {
		setError(errOut, err)
		return 0
	}
	return 1
}

//export gomax_close
func gomax_close(h C.uintptr_t) {
	ch, ok := getHandle(h)
	if !ok {
		return
	}
	_ = ch.client.Close()

	cgo.Handle(h).Delete()
}

//export gomax_send_message
func gomax_send_message(h C.uintptr_t, chatID C.longlong, text *C.char, notify C.int, errOut **C.char) *C.char {
	setError(errOut, nil)
	ch, ok := getHandle(h)
	if !ok {
		setError(errOut, fmt.Errorf("invalid handle"))
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	msg, err := ch.client.SendMessage(ctx, C.GoString(text), int64(chatID), notify != 0, nil, nil, nil)
	if err != nil {
		setError(errOut, err)
		return nil
	}

	payload, _ := json.Marshal(msg)
	return C.CString(string(payload))
}

//export gomax_get_profile
func gomax_get_profile(h C.uintptr_t, errOut **C.char) *C.char {
	setError(errOut, nil)
	ch, ok := getHandle(h)
	if !ok {
		setError(errOut, fmt.Errorf("invalid handle"))
		return nil
	}

	me := ch.client.Profile()
	if me == nil {
		return nil
	}

	payload, _ := json.Marshal(me)
	return C.CString(string(payload))
}

//export gomax_get_chats
func gomax_get_chats(h C.uintptr_t, errOut **C.char) *C.char {
	setError(errOut, nil)
	ch, ok := getHandle(h)
	if !ok {
		setError(errOut, fmt.Errorf("invalid handle"))
		return nil
	}

	chats := ch.client.ChatList()
	payload, _ := json.Marshal(chats)
	return C.CString(string(payload))
}

//export gomax_set_on_message
func gomax_set_on_message(h C.uintptr_t, cb C.gomax_on_message_cb, userData unsafe.Pointer) {
	ch, ok := getHandle(h)
	if !ok {
		return
	}

	ch.onMessage = cb
	ch.userData = userData

	ch.client.OnMessage(func(ctx context.Context, msg *types.Message) {
		if ch.onMessage == nil {
			return
		}
		data, _ := json.Marshal(msg)
		cstr := C.CString(string(data))
		C.gomax_call_on_message(ch.onMessage, cstr, ch.userData)
		C.free(unsafe.Pointer(cstr))
	}, nil)
}

//export gomax_free_string
func gomax_free_string(str *C.char) {
	if str != nil {
		C.free(unsafe.Pointer(str))
	}
}

// main заглушка нужна для сборки в режиме c-shared.
func main() {}
