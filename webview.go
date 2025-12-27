package webview

/*
#cgo CXXFLAGS: -std=c++14
#cgo windows CXXFLAGS: -DWEBVIEW_API=extern
#cgo windows LDFLAGS: -lole32 -lcomctl32 -loleaut32 -luuid -lgdi32 -luser32 -lversion -lshlwapi

// [리눅스 설정 - ★ 추가됨]
#cgo linux pkg-config: gtk+-3.0 webkit2gtk-4.0

#include <stdlib.h>
#include <stdint.h>
#include "webview.h"

// [1] Go 함수 선언
extern void _webview_binding_cb(char *seq, char *req, void *arg);
extern void _webview_dispatch_cb(webview_t w, void *arg);

// [2] C++ 함수 선언 (구현은 webview_impl.cc에 있음)
extern void CgoWebViewShow(webview_t w);

// ★ [추가] 위치와 크기를 설정하는 C 함수 선언
extern void CgoWebViewSetBounds(webview_t w, int x, int y, int width, int height, int hint);

// [3] 접착제 코드 (간단한 C 함수들)
static void CgoWebViewBind(webview_t w, const char *name, uintptr_t arg) {
	typedef void (*callback_t)(const char *seq, const char *req, void *arg);
	webview_bind(w, name, (callback_t)_webview_binding_cb, (void *)arg);
}

static void CgoWebViewUnbind(webview_t w, const char *name) {
	webview_unbind(w, name);
}

static void CgoWebViewDispatch(webview_t w, uintptr_t arg) {
	webview_dispatch(w, _webview_dispatch_cb, (void *)arg);
}
*/
import "C"

import (
	"encoding/json"
	"errors"
	"reflect"
	"sync"
	"unsafe"
)

type WebView interface {
	Run()
	Terminate()
	Dispatch(f func())
	Destroy()
	Window() unsafe.Pointer
	SetTitle(title string)
	SetSize(x int, y int, w int, h int, hint int) // [수정] x, y 추가
	Navigate(url string)
	SetHtml(html string)
	Init(js string)
	Eval(js string)
	Bind(name string, f interface{}) error
	Unbind(name string) error
	Show() // ★ [추가] Go에서 호출할 메서드
}

type webview struct {
	w C.webview_t
}

type bindingData struct {
	w C.webview_t
	f interface{}
}

var (
	m        sync.Mutex
	bindings = map[uintptr]bindingData{}
)

const (
	HintNone  = 0
	HintMin   = 1
	HintMax   = 2
	HintFixed = 3
)

func New(debug bool) WebView {
	w := &webview{}
	d := 0
	if debug {
		d = 1
	}
	w.w = C.webview_create(C.int(d), nil)
	return w
}

func (w *webview) Destroy() {
	C.webview_destroy(w.w)
}

func (w *webview) Run() {
	Show()
	C.webview_run(w.w)
}

func (w *webview) Terminate() {
	C.webview_terminate(w.w)
}

// ★ [추가] 창 보이기 구현
func (w *webview) Show() {
	C.CgoWebViewShow(w.w)
}

func (w *webview) Window() unsafe.Pointer {
	return unsafe.Pointer(C.webview_get_window(w.w))
}

func (w *webview) Navigate(url string) {
	s := C.CString(url)
	defer C.free(unsafe.Pointer(s))
	C.webview_navigate(w.w, s)
}

func (w *webview) SetHtml(html string) {
	s := C.CString(html)
	defer C.free(unsafe.Pointer(s))
	C.webview_set_html(w.w, s)
}

func (w *webview) SetTitle(title string) {
	s := C.CString(title)
	defer C.free(unsafe.Pointer(s))
	C.webview_set_title(w.w, s)
}

// ★ [수정] SetSize 구현 변경
// 기존 C.webview_set_size 대신 새로 만든 C.CgoWebViewSetBounds를 호출
func (w *webview) SetSize(x int, y int, width int, height int, hint int) {
	C.CgoWebViewSetBounds(w.w, C.int(x), C.int(y), C.int(width), C.int(height), C.int(hint))
}

func (w *webview) Init(js string) {
	s := C.CString(js)
	defer C.free(unsafe.Pointer(s))
	C.webview_init(w.w, s)
}

func (w *webview) Eval(js string) {
	s := C.CString(js)
	defer C.free(unsafe.Pointer(s))
	C.webview_eval(w.w, s)
}

func (w *webview) Dispatch(f func()) {
	m.Lock()
	for ; ; {
		p := uintptr(unsafe.Pointer(new(int)))
		if _, ok := bindings[p]; !ok {
			bindings[p] = bindingData{f: f}
			m.Unlock()
			C.CgoWebViewDispatch(w.w, C.uintptr_t(p))
			return
		}
	}
}

func (w *webview) Bind(name string, f interface{}) error {
	v := reflect.ValueOf(f)
	if v.Kind() != reflect.Func {
		return errors.New("only functions can be bound")
	}
	if v.Type().NumOut() > 2 {
		return errors.New("function may only return a value or a value+error")
	}
	m.Lock()
	p := uintptr(unsafe.Pointer(new(int)))
	bindings[p] = bindingData{w: w.w, f: f}
	m.Unlock()
	
	s := C.CString(name)
	defer C.free(unsafe.Pointer(s))
	C.CgoWebViewBind(w.w, s, C.uintptr_t(p))
	return nil
}

func (w *webview) Unbind(name string) error {
	s := C.CString(name)
	defer C.free(unsafe.Pointer(s))
	C.CgoWebViewUnbind(w.w, s)
	m.Lock()
	m.Unlock()
	return nil
}

//export _webview_dispatch_cb
func _webview_dispatch_cb(w C.webview_t, arg unsafe.Pointer) {
	p := uintptr(arg)
	m.Lock()
	data, ok := bindings[p]
	delete(bindings, p)
	m.Unlock()
	if ok {
		if f, ok := data.f.(func()); ok {
			f()
		}
	}
}

//export _webview_binding_cb
func _webview_binding_cb(seq *C.char, req *C.char, arg unsafe.Pointer) {
	p := uintptr(arg)
	m.Lock()
	data, ok := bindings[p]
	m.Unlock()
	if !ok {
		return
	}

	w := data.w
	f := data.f

	jsString := func(v interface{}) string { b, _ := json.Marshal(v); return string(b) }
	
	var params []json.RawMessage
	if err := json.Unmarshal([]byte(C.GoString(req)), &params); err != nil {
		return
	}

	isError := func(t reflect.Type) bool {
		return t.Implements(reflect.TypeOf((*error)(nil)).Elem())
	}
	
	v := reflect.ValueOf(f)
	in := make([]reflect.Value, v.Type().NumIn())
	for i := range in {
		if i >= len(params) {
			in[i] = reflect.Zero(v.Type().In(i))
			continue
		}
		arg := reflect.New(v.Type().In(i))
		json.Unmarshal(params[i], arg.Interface())
		in[i] = arg.Elem()
	}

	out := v.Call(in)
	
	result := ""
	if len(out) > 0 {
		if isError(out[len(out)-1].Type()) {
			if out[len(out)-1].Interface() != nil {
				s := C.CString(jsString(out[len(out)-1].Interface().(error).Error()))
				defer C.free(unsafe.Pointer(s))
				C.webview_return(w, seq, 1, s)
				return
			}
			out = out[:len(out)-1]
		}
	}
	
	if len(out) > 0 {
		result = jsString(out[0].Interface())
	} else {
		result = "null"
	}
	
	s := C.CString(result)
	defer C.free(unsafe.Pointer(s))
	C.webview_return(w, seq, 0, s)
}