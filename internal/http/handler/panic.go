package handler

import "net/http"

func Panic(_ http.ResponseWriter, _ *http.Request) {
	panic("test panic")
}
