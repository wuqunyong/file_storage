package controller

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

type SignPart struct {
	PartNumber int         `json:"partNumber"`
	URL        string      `json:"url"`
	Query      url.Values  `json:"query"`
	Header     http.Header `json:"header"`
}

type AuthSignResult struct {
	URL    string      `json:"url"`
	Query  url.Values  `json:"query"`
	Header http.Header `json:"header"`
	Parts  []SignPart  `json:"parts"`
}

type InitiateUploadResult struct {
	// UploadID uniquely identifies the upload session for tracking and management purposes.
	UploadID string `json:"uploadID"`

	// PartSize specifies the size of each part in a multipart upload. This is relevant for breaking down large uploads into manageable pieces.
	PartSize int64 `json:"partSize"`

	// Sign contains the authentication and signature information necessary for securely uploading each part. This could include signed URLs or tokens.
	Sign *AuthSignResult `json:"sign"`
}

type UploadResult struct {
	Hash string `json:"hash"`
	Size int64  `json:"size"`
	Key  string `json:"key"`
}

type CopyObjectInfo struct {
	Key  string `json:"name"`
	ETag string `json:"etag"`
}

type AccessURLOpt struct {
	ContentType string
	Filename    string `binding:"required"`
}
type S3Database interface {
	InitiateMultipartUpload(ctx context.Context, hash string, size int64, expire time.Duration, maxParts int) (*InitiateUploadResult, error)
	CompleteMultipartUpload(ctx context.Context, uploadID string, parts []string) (*UploadResult, error)
}
