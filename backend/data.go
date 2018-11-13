package backend

import (
	"context"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

type EntryType string

const (
	TypeFile      EntryType = "file"
	TypeDirectory EntryType = "directory"
)

type Entry struct {
	Name         string     `json:"name"`
	Type         EntryType  `json:"type"`
	Size         int64      `json:"size,omitempty"`
	ETag         string     `json:"eTag,omitempty"`
	LastModified *time.Time `json:"lastModified,omitempty"`
}

type Storage struct {
	Bucket string
	s3     *s3.S3
}

func NewStorage(src *s3.S3, bucket string) *Storage {
	return &Storage{s3: src, Bucket: bucket}
}

func (s *Storage) List(ctx context.Context, path string) ([]*Entry, error) {
	prefixes := []*s3.CommonPrefix{}
	objects := []*s3.Object{}

	var prefix *string
	if path != "" {
		prefix = &path
	}

	err := s.s3.ListObjectsV2PagesWithContext(
		ctx,
		&s3.ListObjectsV2Input{Bucket: &s.Bucket, Prefix: prefix, Delimiter: aws.String("/")},
		func(page *s3.ListObjectsV2Output, _ bool) bool {
			prefixes = append(prefixes, page.CommonPrefixes...)
			objects = append(objects, page.Contents...)
			return true
		},
	)
	if err != nil {
		return nil, err
	}

	entries := make([]*Entry, 0, len(prefixes)+len(objects))
	for _, p := range prefixes {
		entries = append(entries, &Entry{Name: strings.TrimPrefix(*p.Prefix, path), Type: TypeDirectory})
	}
	for _, o := range objects {
		e := &Entry{
			Name: strings.TrimPrefix(*o.Key, path),
			Type: TypeFile,
		}
		if o.Size != nil {
			e.Size = *o.Size
		}
		if o.ETag != nil {
			e.ETag = *o.ETag
		}
		if o.LastModified != nil {
			e.LastModified = o.LastModified
		}
		entries = append(entries, e)
	}

	return entries, nil
}

func (s *Storage) Head(ctx context.Context, path string, cond *Conditions) (*ObjectMeta, error) {
	inp := s3.HeadObjectInput{
		Bucket: &s.Bucket,
		Key:    &path,
	}
	if cond.IfMatch != "" {
		inp.IfMatch = &cond.IfMatch
	} else if cond.IfNoneMatch != "" {
		inp.IfNoneMatch = &cond.IfNoneMatch
	}
	if !cond.IfModifiedSince.IsZero() {
		inp.IfModifiedSince = &cond.IfModifiedSince
	} else if !cond.IfUnmodifiedSince.IsZero() {
		inp.IfUnmodifiedSince = &cond.IfUnmodifiedSince
	}

	out, err := s.s3.HeadObject(&inp)
	if err != nil {
		return nil, err
	}
	return ObjectMetaFromObjectHead(out), nil
}

func (s *Storage) Get(ctx context.Context, path string, cond *Conditions) (*ObjectMeta, io.ReadCloser, error) {
	inp := s3.GetObjectInput{
		Bucket: &s.Bucket,
		Key:    &path,
	}
	if cond.IfMatch != "" {
		inp.IfMatch = &cond.IfMatch
	} else if cond.IfNoneMatch != "" {
		inp.IfNoneMatch = &cond.IfNoneMatch
	}
	if !cond.IfModifiedSince.IsZero() {
		inp.IfModifiedSince = &cond.IfModifiedSince
	} else if !cond.IfUnmodifiedSince.IsZero() {
		inp.IfUnmodifiedSince = &cond.IfUnmodifiedSince
	}

	out, err := s.s3.GetObject(&inp)
	if err != nil {
		return nil, nil, err
	}
	return ObjectMetaFromObject(out), out.Body, nil
}
