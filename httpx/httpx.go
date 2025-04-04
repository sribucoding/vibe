package httpx

import "net/http"

type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

func (h HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h(w, r); err != nil {
		err = InternalError(w, err)
		if err != nil {
			panic(err)
		}
	}
}
