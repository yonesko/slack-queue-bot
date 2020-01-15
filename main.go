package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {
	http.HandleFunc("/add-cmd", func(writer http.ResponseWriter, request *http.Request) {
		bytes, err := ioutil.ReadAll(request.Body)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(string(bytes))
		if err := request.Body.Close(); err != nil {
			fmt.Println(err)
		}
	})
	if err := http.ListenAndServe(":5678", nil); err != nil {
		fmt.Println(err)
	}
}
