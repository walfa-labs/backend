package s3

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// AssetStore implements port.AssetStore against an S3-compatible object store
// via the MinIO client.
type AssetStore struct {
	client       *minio.Client
	bucket       string
	usePathStyle bool
}

// NewAssetStore constructs an AssetStore. When usePathStyle is true, generated
// URLs are path-style ("/{bucket}/{key}") and the client is configured with
// BucketLookupPath; otherwise virtual-host-style URLs are used
// ("https://{endpoint}/{bucket}/{key}").
func NewAssetStore(endpoint, accessKey, secretKey, bucket string, usePathStyle bool) (*AssetStore, error) {
	lookup := minio.BucketLookupAuto
	if usePathStyle {
		lookup = minio.BucketLookupPath
	}
	host, secure := splitEndpointScheme(endpoint)
	cli, err := minio.New(host, &minio.Options{
		Creds:        credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure:       secure,
		BucketLookup: lookup,
	})
	if err != nil {
		return nil, fmt.Errorf("s3: create minio client: %w", err)
	}
	cli.SetAppInfo("walfa-labs-backend", "1.0.0")
	return &AssetStore{
		client:       cli,
		bucket:       bucket,
		usePathStyle: usePathStyle,
	}, nil
}

// Upload stores an object and returns its public URL.
func (s *AssetStore) Upload(ctx context.Context, key string, r io.Reader, contentType string, size int64) (string, error) {
	_, err := s.client.PutObject(ctx, s.bucket, key, r, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("s3: upload %q: %w", key, err)
	}
	return s.objectURL(key), nil
}

// Presign returns a pre-signed GET URL valid for one hour.
func (s *AssetStore) Presign(ctx context.Context, key string) (string, error) {
	u, err := s.client.PresignedGetObject(ctx, s.bucket, key, time.Hour, nil)
	if err != nil {
		return "", fmt.Errorf("s3: presign %q: %w", key, err)
	}
	return u.String(), nil
}

// Delete removes an object from the bucket.
func (s *AssetStore) Delete(ctx context.Context, key string) error {
	if err := s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("s3: delete %q: %w", key, err)
	}
	return nil
}

// objectURL builds the canonical URL for an object key.
func (s *AssetStore) objectURL(key string) string {
	if s.usePathStyle {
		return "/" + s.bucket + "/" + key
	}
	return fmt.Sprintf("https://%s/%s/%s", s.client.EndpointURL().Host, s.bucket, key)
}

// splitEndpointScheme strips an optional scheme prefix from the endpoint and
// reports whether TLS should be used. minio-go prepends the scheme itself, so
// the endpoint it receives must be a bare host[:port]. Bare endpoints default
// to TLS; an explicit http:// prefix forces plaintext (local MinIO).
func splitEndpointScheme(endpoint string) (host string, secure bool) {
	lower := strings.ToLower(endpoint)
	switch {
	case strings.HasPrefix(lower, "http://"):
		return endpoint[len("http://"):], false
	case strings.HasPrefix(lower, "https://"):
		return endpoint[len("https://"):], true
	default:
		return endpoint, true
	}
}
