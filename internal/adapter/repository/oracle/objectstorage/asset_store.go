// Package objectstorage implements port.AssetStore against Oracle Cloud
// Infrastructure Object Storage via the OCI Go SDK.
package objectstorage

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/objectstorage"
)

// presignTTL mirrors the previous S3 adapter's one-hour pre-signed URL expiry.
const presignTTL = time.Hour

// AssetStore implements port.AssetStore against OCI Object Storage.
type AssetStore struct {
	client    objectstorage.ObjectStorageClient
	namespace string
	bucket    string
	region    string
}

// NewAssetStore constructs an AssetStore from explicit OCI credentials. The
// PEM file at privateKeyPath must be an unencrypted private key. When
// namespace is empty it is resolved via the GetNamespace API (fail fast).
func NewAssetStore(tenancyOCID, userOCID, fingerprint, region, privateKeyPath, namespace, bucket string) (*AssetStore, error) {
	// #nosec G304 -- privateKeyPath comes from server configuration (env), not user input.
	keyBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("oci objectstorage: read private key: %w", err)
	}

	provider := common.NewRawConfigurationProvider(tenancyOCID, userOCID, region, fingerprint, string(keyBytes), nil)
	client, err := objectstorage.NewObjectStorageClientWithConfigurationProvider(provider)
	if err != nil {
		return nil, fmt.Errorf("oci objectstorage: create client: %w", err)
	}

	if namespace == "" {
		resp, err := client.GetNamespace(context.Background(), objectstorage.GetNamespaceRequest{})
		if err != nil {
			return nil, fmt.Errorf("oci objectstorage: resolve namespace: %w", err)
		}
		if resp.Value == nil {
			return nil, fmt.Errorf("oci objectstorage: resolve namespace: empty response")
		}
		namespace = *resp.Value
	}

	return &AssetStore{
		client:    client,
		namespace: namespace,
		bucket:    bucket,
		region:    region,
	}, nil
}

// Upload stores an object and returns its canonical (non-signed) URL. Public
// reads still go through Presign, as before.
func (s *AssetStore) Upload(ctx context.Context, key string, r io.Reader, contentType string, size int64) (string, error) {
	body, ok := r.(io.ReadCloser)
	if !ok {
		body = io.NopCloser(r)
	}
	_, err := s.client.PutObject(ctx, objectstorage.PutObjectRequest{
		NamespaceName: common.String(s.namespace),
		BucketName:    common.String(s.bucket),
		ObjectName:    common.String(key),
		ContentType:   common.String(contentType),
		ContentLength: common.Int64(size),
		PutObjectBody: body,
	})
	if err != nil {
		return "", fmt.Errorf("oci objectstorage: upload %q: %w", key, err)
	}
	return s.objectURL(key), nil
}

// Presign returns a pre-authenticated request (PAR) URL valid for one hour —
// the OCI equivalent of an S3 pre-signed GET. Note that PARs are named
// resources that accumulate server-side until they expire, unlike SigV4
// presignatures which are stateless.
func (s *AssetStore) Presign(ctx context.Context, key string) (string, error) {
	expires := common.SDKTime{Time: time.Now().Add(presignTTL)}
	resp, err := s.client.CreatePreauthenticatedRequest(ctx, objectstorage.CreatePreauthenticatedRequestRequest{
		NamespaceName: common.String(s.namespace),
		BucketName:    common.String(s.bucket),
		CreatePreauthenticatedRequestDetails: objectstorage.CreatePreauthenticatedRequestDetails{
			Name:        common.String("presign-" + uuid.New().String()),
			AccessType:  objectstorage.CreatePreauthenticatedRequestDetailsAccessTypeObjectread,
			ObjectName:  common.String(key),
			TimeExpires: &expires,
		},
	})
	if err != nil {
		return "", fmt.Errorf("oci objectstorage: presign %q: %w", key, err)
	}
	if resp.AccessUri == nil {
		return "", fmt.Errorf("oci objectstorage: presign %q: empty access URI", key)
	}
	return s.endpointURL() + *resp.AccessUri, nil
}

// Delete removes an object from the bucket.
func (s *AssetStore) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, objectstorage.DeleteObjectRequest{
		NamespaceName: common.String(s.namespace),
		BucketName:    common.String(s.bucket),
		ObjectName:    common.String(key),
	})
	if err != nil {
		return fmt.Errorf("oci objectstorage: delete %q: %w", key, err)
	}
	return nil
}

// endpointURL builds the regional Object Storage endpoint base.
func (s *AssetStore) endpointURL() string {
	return fmt.Sprintf("https://objectstorage.%s.oraclecloud.com", s.region)
}

// objectURL builds the canonical URL for an object key. Keys keep their '/'
// path separators; PAR access URIs returned by OCI are already escaped.
func (s *AssetStore) objectURL(key string) string {
	return fmt.Sprintf("%s/n/%s/b/%s/o/%s", s.endpointURL(), s.namespace, s.bucket, key)
}
