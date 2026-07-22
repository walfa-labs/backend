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
	cli, err := minio.New(endpoint, &minio.Options{
		Creds:        credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure:       useSecureTLS(endpoint),
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

// useSecureTLS returns true unless the endpoint explicitly uses http://.
func useSecureTLS(endpoint string) bool {
	return !strings.HasPrefix(strings.ToLower(endpoint), "http://")
}
