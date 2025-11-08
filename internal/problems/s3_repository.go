package problems

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/gate149/core/pkg"
	"github.com/google/uuid"
)

type S3Repository struct {
	s3Client *s3.Client
	bucket   string
}

func NewS3Repository(s3Client *s3.Client, bucket string) *S3Repository {
	return &S3Repository{
		s3Client: s3Client,
		bucket:   bucket,
	}
}

func (r *S3Repository) UploadTestsFile(ctx context.Context, problemID uuid.UUID, reader io.Reader) (string, error) {
	const op = "S3Repository.UploadTestsFile"

	// Generate S3 key for the archive
	key := fmt.Sprintf("problems/%s/tests.zip", problemID)

	// Create multipart upload
	mpu, err := r.s3Client.CreateMultipartUpload(ctx, &s3.CreateMultipartUploadInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return "", pkg.Wrap(pkg.ErrInternal, err, op, "failed to create multipart upload")
	}

	const chunkSize = 5 * 1024 * 1024 // 5MB chunks
	var completedParts []types.CompletedPart
	partNumber := int32(1)

	// Upload parts
	for {
		partBuf := make([]byte, chunkSize)
		n, err := reader.Read(partBuf)
		if err != nil && err != io.EOF {
			// Abort the multipart upload on error
			_, _ = r.s3Client.AbortMultipartUpload(ctx, &s3.AbortMultipartUploadInput{
				Bucket:   aws.String(r.bucket),
				Key:      aws.String(key),
				UploadId: mpu.UploadId,
			})
			return "", pkg.Wrap(pkg.ErrInternal, err, op, "failed to read part")
		}
		if n == 0 {
			break // No more data to upload
		}

		// Upload part
		uploadPart, err := r.s3Client.UploadPart(ctx, &s3.UploadPartInput{
			Bucket:     aws.String(r.bucket),
			Key:        aws.String(key),
			PartNumber: aws.Int32(partNumber),
			UploadId:   mpu.UploadId,
			Body:       bytes.NewReader(partBuf[:n]),
		})
		if err != nil {
			// Abort the multipart upload on error
			_, _ = r.s3Client.AbortMultipartUpload(ctx, &s3.AbortMultipartUploadInput{
				Bucket:   aws.String(r.bucket),
				Key:      aws.String(key),
				UploadId: mpu.UploadId,
			})
			return "", pkg.Wrap(pkg.ErrInternal, err, op, "failed to upload part")
		}

		completedParts = append(completedParts, types.CompletedPart{
			ETag:       uploadPart.ETag,
			PartNumber: aws.Int32(partNumber),
		})
		partNumber++
	}

	// Complete multipart upload
	_, err = r.s3Client.CompleteMultipartUpload(ctx, &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(r.bucket),
		Key:      aws.String(key),
		UploadId: mpu.UploadId,
		MultipartUpload: &types.CompletedMultipartUpload{
			Parts: completedParts,
		},
	})
	if err != nil {
		// Abort the multipart upload on error
		_, _ = r.s3Client.AbortMultipartUpload(ctx, &s3.AbortMultipartUploadInput{
			Bucket:   aws.String(r.bucket),
			Key:      aws.String(key),
			UploadId: mpu.UploadId,
		})
		return "", pkg.Wrap(pkg.ErrInternal, err, op, "failed to complete multipart upload")
	}

	return key, nil
}

func (r *S3Repository) DownloadTestsFile(ctx context.Context, problemId uuid.UUID) (io.ReadCloser, error) {
	key := fmt.Sprintf("problems/%s/tests.zip", problemId)

	resp, err := r.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		return nil, pkg.Wrap(pkg.ErrInternal, err, "S3Repository.DownloadTestsFile", "failed to get object")
	}

	return resp.Body, nil
}
