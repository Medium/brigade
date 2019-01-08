package frontend

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Medium/brigade/backend"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/golang/gddo/httputil/header"
	"go.uber.org/zap"
)

type Server struct {
	Backend *backend.Storage
	Logger  *zap.Logger
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "HEAD" {
		http.Error(w, fmt.Sprintf("%s requests are not supported", r.Method), http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	var path string
	if r.URL.Path != "/" {
		path = strings.TrimPrefix(r.URL.Path, "/")
	}

	if path == "" || strings.HasSuffix(path, "/") {
		s.ServeListing(ctx, path, w, r)
		return
	}

	cond := backend.ConditionsFromRequest(r)

	if r.Method == "HEAD" {
		s.ServeHead(ctx, path, cond, w)
		return
	}

	s.ServeObject(ctx, path, cond, w)
}

func (s *Server) ServeListing(ctx context.Context, path string, w http.ResponseWriter, r *http.Request) {
	s.Logger.Info("Listing objects", zap.String("path", path))

	entries, err := s.Backend.List(ctx, path)
	if err != nil {
		s.serveError(w, "list S3 objects", err)
		return
	}

	if requestsHTML(r) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		listing := Listing{Host: r.URL.Host, Path: path, Entries: entries}
		if err := listingTemplate.Execute(w, &listing); err != nil {
			s.Logger.Error("Error rendering listing template", zap.Error(err))
		}
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		enc := json.NewEncoder(w)
		enc.Encode(entries)
	}
}

func (s *Server) ServeHead(ctx context.Context, path string, cond *backend.Conditions, w http.ResponseWriter) {
	s.Logger.Info("Serving object head", zap.String("path", path))

	meta, err := s.Backend.Head(ctx, path, cond)
	if err != nil {
		s.serveError(w, "HEAD an S3 object", err)
		return
	}

	meta.WriteHeaders(w.Header())
	w.WriteHeader(http.StatusOK)
}

func (s *Server) ServeObject(ctx context.Context, path string, cond *backend.Conditions, w http.ResponseWriter) {
	s.Logger.Info("Serving object", zap.String("path", path))

	meta, body, err := s.Backend.Get(ctx, path, cond)
	if err != nil {
		s.serveError(w, "GET an S3 object", err)
		return
	}

	fmt.Printf("If-Modified-Since: %v\n", cond.IfModifiedSince)

	defer body.Close()

	meta.WriteHeaders(w.Header())
	w.WriteHeader(http.StatusOK)

	io.Copy(w, body)
}

func (s *Server) serveError(w http.ResponseWriter, op string, err error) {
	status := http.StatusInternalServerError

	if reqerr, ok := err.(awserr.RequestFailure); ok {
		status = reqerr.StatusCode()
		if status == http.StatusNotModified || status == http.StatusPreconditionFailed {
			w.Header().Set("Content-Length", "0")
			w.WriteHeader(status)
			return
		}
		if status == http.StatusNotFound {
			http.Error(w, "Not found", status)
			return
		}
	}

	s.Logger.Error("Operation failed",
		zap.String("operation", op),
		zap.Int("status", status),
		zap.Error(err))

	http.Error(w, fmt.Sprintf("Failed to %s: %v", op, err), status)
}

func requestsHTML(r *http.Request) bool {
	accepts := header.ParseAccept(r.Header, "Accept")

	var html, json float64
	for _, a := range accepts {
		if a.Value == "text/html" {
			html = a.Q
		} else if a.Value == "application/json" {
			json = a.Q
		}
	}

	return html > json
}
