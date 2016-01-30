package main

import (
	"encoding/json"
	"log"
	"net/http"
	"runtime/debug"
)

func jsonM(f func(*http.Request) (interface{}, HTTPError)) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, req *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				log.Println("PANIC:", r)
				debug.PrintStack()
				NewHTTPError("Recovered from panic.", 500).Write(writer)
			}
		}()

		writer.Header().Set("Content-Type", "application/json")
		res, err := f(req)
		if err != nil {
			err.Write(writer)
		}

		out, err2 := json.Marshal(res)
		if err2 != nil {
			http.Error(writer, "Unable to encode response.", 500)
			return
		}
		writer.Write(out)
	}
}
