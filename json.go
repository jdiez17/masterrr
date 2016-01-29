package main

import (
	"net/http"
	"encoding/json"
)

func jsonM(f func(*http.Request) (interface{}, HTTPError)) func(http.ResponseWriter, *http.Request)  {
	return func(writer http.ResponseWriter, req *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		res, err := f(req)
		if err != nil {
			out, _ := json.Marshal(StatusMessageResponse{
				Message: err.Error(),
				Success: false,
			})

			writer.WriteHeader(err.Code())
			writer.Write(out)
			return
		}

		out, err2 := json.Marshal(res)
		if err2 != nil {
			http.Error(writer, "Unable to encode response.", 500)
			return
		}
		writer.Write(out)
	}
}
