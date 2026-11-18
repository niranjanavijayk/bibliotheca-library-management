// Package handler exposes the REST API.
package handler

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"bibliotheca/internal/apperror"
)

type envelope struct {
	Data  any   `json:"data,omitempty"`
	Meta  *meta `json:"meta,omitempty"`
	Error any   `json:"error,omitempty"`
}

type meta struct {
	Total    int `json:"total"`
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("encode response: %v", err)
	}
}

func ok(w http.ResponseWriter, status int, data any) {
	if data == nil {
		data = map[string]any{}
	}
	writeJSON(w, status, envelope{Data: data})
}

func okList(w http.ResponseWriter, data any, total, page, size int) {
	writeJSON(w, http.StatusOK, envelope{Data: data, Meta: &meta{Total: total, Page: page, PageSize: size}})
}

func fail(w http.ResponseWriter, err error) {
	ae := apperror.From(err)
	if ae.Status >= 500 {
		log.Printf("internal error: %v", err)
	}
	writeJSON(w, ae.Status, envelope{Error: ae})
}

func pathID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, apperror.BadRequest("Invalid id")
	}
	return id, nil
}

func decode(r *http.Request, v any) error {
	r.Body = http.MaxBytesReader(nil, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(v); err != nil {
		if errors.Is(err, io.EOF) {
			return apperror.BadRequest("Request body is empty")
		}
		return apperror.BadRequest("Request body is not valid JSON")
	}
	return nil
}

func intParam(r *http.Request, name string, def, min, max int) int {
	v, err := strconv.Atoi(r.URL.Query().Get(name))
	if err != nil || v < min {
		return def
	}
	if v > max {
		return max
	}
	return v
}

func boolParam(r *http.Request, name string) *bool {
	switch strings.ToLower(r.URL.Query().Get(name)) {
	case "true", "1", "yes":
		t := true
		return &t
	case "false", "0", "no":
		f := false
		return &f
	}
	return nil
}
