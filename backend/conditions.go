package backend

import (
	"net/http"
	"time"
)

type Conditions struct {
	IfMatch           string
	IfModifiedSince   time.Time
	IfNoneMatch       string
	IfUnmodifiedSince time.Time
}

func ConditionsFromRequest(req *http.Request) *Conditions {
	cond := &Conditions{}

	if c := req.Header.Get("If-Match"); c != "" {
		cond.IfMatch = c
	} else if c := req.Header.Get("If-None-Match"); c != "" {
		cond.IfNoneMatch = c
	}

	if c := req.Header.Get("If-Modified-Since"); c != "" {
		if t, err := http.ParseTime(c); err == nil {
			cond.IfModifiedSince = t
		}
	} else if c := req.Header.Get("If-Unmodified-Since"); c != "" {
		if t, err := http.ParseTime(c); err == nil {
			cond.IfUnmodifiedSince = t
		}
	}

	return cond
}
