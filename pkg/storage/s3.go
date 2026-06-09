package storage

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Store stores objects through an S3-compatible API.
type S3Store struct {
	client        *s3.Client
	presign       *s3.PresignClient
	bucket        string
	publicBaseURL string
}

// NewS3Store creates an S3-compatible store.
func NewS3Store(ctx context.Context, cfg Config) (S3Store, error) {
	if strings.TrimSpace(cfg.Bucket) == "" {
		return S3Store{}, fmt.Errorf("storage bucket is required")
	}
	if strings.TrimSpace(cfg.AccessKeyID) == "" {
		return S3Store{}, fmt.Errorf("storage access key id is required")
	}
	if strings.TrimSpace(cfg.SecretAccessKey) == "" {
		return S3Store{}, fmt.Errorf("storage secret access key is required")
	}
	region := strings.TrimSpace(cfg.Region)
	if region == "" {
		region = "auto"
	}
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, "")),
	)
	if err != nil {
		return S3Store{}, fmt.Errorf("load s3 configuration: %w", err)
	}
	client := s3.NewFromConfig(awsCfg, func(options *s3.Options) {
		if strings.TrimSpace(cfg.Endpoint) != "" {
			options.BaseEndpoint = &cfg.Endpoint
			options.UsePathStyle = true
		}
	})
	return S3Store{
		client:        client,
		presign:       s3.NewPresignClient(client),
		bucket:        cfg.Bucket,
		publicBaseURL: strings.TrimRight(cfg.PublicBaseURL, "/"),
	}, nil
}

// Health verifies the S3 bucket is reachable.
func (store S3Store) Health(ctx context.Context) error {
	if _, err := store.client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: &store.bucket}); err != nil {
		return fmt.Errorf("head s3 bucket %s: %w", store.bucket, err)
	}
	return nil
}

// Put stores object bytes.
func (store S3Store) Put(ctx context.Context, object Object, body io.Reader) (StoredObject, error) {
	output, err := store.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &store.bucket,
		Key:         &object.Key,
		Body:        body,
		ContentType: &object.ContentType,
		Metadata:    object.Metadata,
	})
	if err != nil {
		return StoredObject{}, fmt.Errorf("put s3 object %s: %w", object.Key, err)
	}
	return StoredObject{Key: object.Key, ETag: trimETag(output.ETag)}, nil
}

// Delete deletes an object by key.
func (store S3Store) Delete(ctx context.Context, key string) error {
	if _, err := store.client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: &store.bucket, Key: &key}); err != nil {
		return fmt.Errorf("delete s3 object %s: %w", key, err)
	}
	return nil
}

// PresignPut creates a presigned upload request.
func (store S3Store) PresignPut(ctx context.Context, request PresignPutRequest) (PresignedRequest, error) {
	expires := request.ExpiresIn
	if expires <= 0 {
		expires = 15 * time.Minute
	}
	output, err := store.presign.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      &store.bucket,
		Key:         &request.Key,
		ContentType: &request.ContentType,
	}, func(options *s3.PresignOptions) {
		options.Expires = expires
	})
	if err != nil {
		return PresignedRequest{}, fmt.Errorf("presign s3 put %s: %w", request.Key, err)
	}
	return PresignedRequest{
		Method:    output.Method,
		URL:       output.URL,
		Headers:   putHeaders(request.ContentType),
		ExpiresAt: time.Now().UTC().Add(expires),
	}, nil
}

// PresignGet creates a presigned download URL.
func (store S3Store) PresignGet(ctx context.Context, key string, ttl time.Duration) (string, error) {
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}
	output, err := store.presign.PresignGetObject(ctx, &s3.GetObjectInput{Bucket: &store.bucket, Key: &key}, func(options *s3.PresignOptions) {
		options.Expires = ttl
	})
	if err != nil {
		return "", fmt.Errorf("presign s3 get %s: %w", key, err)
	}
	return output.URL, nil
}

// Head returns object metadata.
func (store S3Store) Head(ctx context.Context, key string) (ObjectInfo, error) {
	output, err := store.client.HeadObject(ctx, &s3.HeadObjectInput{Bucket: &store.bucket, Key: &key})
	if err != nil {
		return ObjectInfo{}, fmt.Errorf("head s3 object %s: %w", key, err)
	}
	contentType := ""
	if output.ContentType != nil {
		contentType = *output.ContentType
	}
	sizeBytes := int64(0)
	if output.ContentLength != nil {
		sizeBytes = *output.ContentLength
	}
	return ObjectInfo{
		Key:         key,
		ETag:        trimETag(output.ETag),
		ContentType: contentType,
		SizeBytes:   sizeBytes,
		Metadata:    output.Metadata,
	}, nil
}

// PublicURL returns a public URL for key when configured.
func (store S3Store) PublicURL(key string) string {
	if store.publicBaseURL == "" {
		return ""
	}
	return store.publicBaseURL + "/" + strings.TrimLeft(key, "/")
}

// trimETag normalizes quoted S3 ETags.
func trimETag(value *string) string {
	if value == nil {
		return ""
	}
	return strings.Trim(*value, `"`)
}
