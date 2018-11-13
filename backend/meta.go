package backend

import (
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
)

// ObjectMeta is the metadata we store about an S3 object.
type ObjectMeta struct {
	ContentType   string            `json:"contentType,omitempty"`
	ContentLength *int64            `json:"contentLength"`
	ETag          string            `json:"eTag,omitempty"`
	Expires       time.Time         `json:"expires,omitempty"`
	LastModified  time.Time         `json:"lastModified,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	VersionID     string            `json:"versionID,omitempty"`
}

func ObjectMetaFromObjectHead(head *s3.HeadObjectOutput) *ObjectMeta {
	var expires, lastModified time.Time
	if head.Expires != nil {
		var err error
		expires, err = http.ParseTime(*head.Expires)
		if err != nil {
			expires = expires.UTC()
		}
	}
	if head.LastModified != nil {
		lastModified = *head.LastModified
		if !lastModified.IsZero() {
			lastModified = lastModified.UTC()
		}
	}

	var contentType, eTag, versionID string
	if head.ContentType != nil {
		contentType = *head.ContentType
	}
	if head.ETag != nil {
		eTag = *head.ETag
	}
	if head.VersionId != nil {
		versionID = *head.VersionId
	}

	metadata := make(map[string]string, len(head.Metadata))
	for k, v := range head.Metadata {
		if v != nil {
			metadata[k] = *v
		}
	}

	return &ObjectMeta{
		ContentType:   contentType,
		ContentLength: head.ContentLength,
		ETag:          eTag,
		Expires:       expires,
		LastModified:  lastModified,
		Metadata:      metadata,
		VersionID:     versionID,
	}
}

func ObjectMetaFromObject(obj *s3.GetObjectOutput) *ObjectMeta {
	head := s3.HeadObjectOutput{
		ContentType:   obj.ContentType,
		ContentLength: obj.ContentLength,
		ETag:          obj.ETag,
		Expires:       obj.Expires,
		LastModified:  obj.LastModified,
		Metadata:      obj.Metadata,
		VersionId:     obj.VersionId,
	}

	return ObjectMetaFromObjectHead(&head)
}

func (m *ObjectMeta) WriteHeaders(h http.Header) {
	if m.ContentType != "" {
		h.Set("Content-Type", m.ContentType)
	}
	if m.ContentLength != nil {
		h.Set("Content-Length", fmt.Sprintf("%d", *m.ContentLength))
	}
	if m.ETag != "" {
		h.Set("ETag", m.ETag)
	}
	if !m.Expires.IsZero() {
		h.Set("Expires", m.Expires.Format(http.TimeFormat))
	}
	if !m.LastModified.IsZero() {
		h.Set("Last-Modified", m.LastModified.Format(http.TimeFormat))
	}
	for k, v := range m.Metadata {
		h.Set(fmt.Sprintf("x-amz-meta-%s", k), v)
	}
	if m.VersionID != "" {
		h.Set("x-amz-version-id", m.VersionID)
	}
}
