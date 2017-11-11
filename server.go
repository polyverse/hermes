package hermes

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	QUERY_PARAM_TYPE_KEY        = "type"
	QUERY_PARAM_TYPE_VALUE_TEXT = "text"
	QUERY_PARAM_TYPE_VALUE_JSON = "json"

	CONTENT_TYPE_HEADER = "Content-Type"
	CONTENT_TYPE_JSON   = "application/json"
)

type handler struct{}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleGet(w, r)
	case http.MethodPost:
		handlePost(w, r)
	default:
		malformedRequest(w, "Unsupported HTTP Method for this endpoint.")
	}
}

func handleGet(w http.ResponseWriter, r *http.Request) {
	model := generateModel("", false)

	respType := resolveResponseType(r)
	switch respType {
	case QUERY_PARAM_TYPE_VALUE_TEXT:
		textResponse(w, model)
	case QUERY_PARAM_TYPE_VALUE_JSON:
		jsonResponse(w, model)
	default:
		malformedRequest(w, fmt.Sprintf("Unknown Response Data Type: %s", respType))
	}
}

func handlePost(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get(CONTENT_TYPE_HEADER)
	if contentType != CONTENT_TYPE_JSON {
		malformedRequest(w, fmt.Sprintf("We only accept %s Content-Type for POST requests. "+
			"You specified an unknown Content-Type of %s.", CONTENT_TYPE_JSON, contentType))
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		malformedRequest(w, fmt.Sprintf("Error when reading POST body: %s", err.Error()))
		return
	}

	var m Model = newEmptyModel()
	err = json.Unmarshal(body, &m)
	if err != nil {
		malformedRequest(w, fmt.Sprintf("Error when parsing POST'ed JSON: %s", err.Error()))
		return
	}

	insertModel("", m)
}

func GetHandler() http.Handler {
	return &handler{}
}

func resolveResponseType(r *http.Request) string {
	queryParams := r.URL.Query()
	responseType := queryParams.Get(QUERY_PARAM_TYPE_KEY)
	if responseType == "" {
		responseType = QUERY_PARAM_TYPE_VALUE_TEXT
	}

	return responseType
}

func malformedRequest(w http.ResponseWriter, message string) {
	logger.Errorf("Hermes: Recieved bad request: %s", message)
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(message))
}

func textResponse(w http.ResponseWriter, rm Model) {
	//simplify model for text case
	w.WriteHeader(http.StatusOK)
	for _, key := range keys {
		value := rm[key]
		w.Write([]byte(fmt.Sprintf("%s: %s\n", key, value.Value)))
	}
}

func jsonResponse(w http.ResponseWriter, rm Model) {
	jstr, err := json.MarshalIndent(rm, "", "    ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jstr)
}
