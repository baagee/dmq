package main

import (
	"dmq/client/handle"
	"net/http"
)

func main() {
	http.HandleFunc("/send/single", handle.Single)
	//http.HandleFunc("/send/multiple", send)
	http.ListenAndServe(":8989", nil)
}
