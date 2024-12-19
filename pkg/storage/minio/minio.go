package minio

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/wuqunyong/file_storage/pkg/storage/controller"
)

const (
	// TempPath specifies the directory for temporary files in the 'openim' structure.
	tempPath = "openimtest/temp/"
)

type Config struct {
	Bucket          string
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	PublicRead      bool
}

func NewMinio(ctx context.Context, conf Config) (*Minio, error) {
	u, err := url.Parse(conf.Endpoint)
	if err != nil {
		return nil, err
	}
	opts := &minio.Options{
		Creds:  credentials.NewStaticV4(conf.AccessKeyID, conf.SecretAccessKey, conf.SessionToken),
		Secure: u.Scheme == "https",
	}
	client, err := minio.New(u.Host, opts)
	if err != nil {
		return nil, err
	}
	m := &Minio{
		conf:   conf,
		bucket: conf.Bucket,
		client: client,
		lock:   &sync.Mutex{},
		init:   false,
	}

	if err := m.initMinio(ctx); err != nil {
		return nil, err
	}
	return m, nil
}

type Minio struct {
	conf     Config
	bucket   string
	location string
	client   *minio.Client
	lock     sync.Locker
	init     bool
}

func (m *Minio) initMinio(ctx context.Context) error {
	if m.init {
		return nil
	}
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.init {
		return nil
	}
	exists, err := m.client.BucketExists(ctx, m.conf.Bucket)
	if err != nil {
		return fmt.Errorf("check bucket exists error: %w", err)
	}
	if !exists {
		if err = m.client.MakeBucket(ctx, m.conf.Bucket, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("make bucket error: %w", err)
		}
	}
	if m.conf.PublicRead {
		policy := fmt.Sprintf(
			`{
					"Version": "2012-10-17",
					"Statement": [
						{
							"Action": [
								"s3:GetObject",
								"s3:PutObject"
							],
							"Effect": "Allow",
							"Principal": {
								"AWS": [
									"*"
								]
							},
							"Resource": [
								"arn:aws:s3:::%s/*"
							],
							"Sid": ""
						}
					]
				}`, m.conf.Bucket)
		if err = m.client.SetBucketPolicy(ctx, m.conf.Bucket, policy); err != nil {
			return err
		}
	}
	m.location, err = m.client.GetBucketLocation(ctx, m.conf.Bucket)
	if err != nil {
		return err
	}

	m.init = true
	return nil
}

func (m *Minio) InitiateUpload(ctx context.Context, hash string, size int64, expire time.Duration) (string, error) {
	// Pre-signed upload
	key := path.Join(tempPath, fmt.Sprintf("%s_%d.presigned", hash, size))
	rawURL, err := m.client.PresignedPutObject(ctx, m.bucket, key, expire)
	if err != nil {
		return "", err
	}
	return rawURL.String(), nil
}

func (m *Minio) CompleteUpload(ctx context.Context, dst, src string) (*controller.CopyObjectInfo, error) {
	result, err := m.client.CopyObject(ctx, minio.CopyDestOptions{
		Bucket: m.bucket,
		Object: dst,
	}, minio.CopySrcOptions{
		Bucket: m.bucket,
		Object: src,
	})
	if err != nil {
		return nil, err
	}

	return &controller.CopyObjectInfo{
		Key:  dst,
		ETag: strings.ToLower(result.ETag),
	}, nil
}

func (m *Minio) AccessURL(ctx context.Context, name string, expire time.Duration, opt controller.AccessURLOpt) (string, error) {
	reqParams := make(url.Values)
	if opt.ContentType != "" {
		reqParams.Set("response-content-type", opt.ContentType)
	}
	if opt.Filename != "" {
		reqParams.Set("response-content-disposition", `attachment; filename=`+strconv.Quote(opt.Filename))
	}

	url, err := m.client.PresignedGetObject(ctx, m.bucket, name, expire, reqParams)
	if err != nil {
		return "", err
	}
	return url.String(), nil
}
