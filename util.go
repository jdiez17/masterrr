package main

import (
	"fmt"
	"net/http"
)

func key(key string) string {
	return fmt.Sprintf("masterrr.%d.%s", State.ID, key)
}

func allowCrossOrigin(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8080")
		f(w, r)
	}
}
