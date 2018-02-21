package async

import (
	"encoding/json"
	"net/http"
)

type handler struct {
	dispatcher *dispatcher
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()

	logger.Printf("handler: %s - %s %s\n", r.RemoteAddr, r.Method, r.URL.Path)

	if r.Method != "POST" {
		http.Error(w, `method not allowed`, http.StatusMethodNotAllowed)
		return
	}

	var in FunctionRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		logger.Printf("handler: failed decoding request payload: %v", err)
		http.Error(w, "failed decoding request payload", http.StatusBadRequest)
		return
	}

	if err := h.dispatcher.dispatch(in); err != nil {
		logger.Printf("handler: failed to dispatch %s: %v", in.Function, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
