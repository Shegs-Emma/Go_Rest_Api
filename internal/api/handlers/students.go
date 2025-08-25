package handlers

import "net/http"

func StudentsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method{
	case http.MethodGet:
		w.Write([]byte("Hello Get Method on Students Route"))
	case http.MethodPost:
		w.Write([]byte("Hello Post Method on Students Route"))
	case http.MethodPut:
		w.Write([]byte("Hello Put Method on Students Route"))
	case http.MethodPatch:
		w.Write([]byte("Hello Patch Method on Students Route"))
	case http.MethodDelete:
		w.Write([]byte("Hello Delete Method on Students Route"))
	}
}