package cortx

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	multierror "github.com/hashicorp/go-multierror"
	"time"
)

//
// NOTE: Mostly Put Together With Bits of the AWS Provider - Truncated Where Possible
//

// EmptyBucket empties the specified S3 bucket by deleting all object versions and delete markers.
// NOTE: Crudely Bypasses All Object Lock Configurations
func EmptyBucket(ctx context.Context, client *s3.S3, bucket string) (int64, error) {

	// Delete Object Versions
	nObjects, err := forEachObjectVersionsPage(ctx, client, bucket, func(ctx context.Context, client *s3.S3, bucket string, page *s3.ListObjectVersionsOutput) (int64, error) {
		return deletePageOfObjectVersions(ctx, client, bucket, page)
	})

	if err != nil {
		return nObjects, err
	}

	// Delete Object Version Delete Markers
	n, err := forEachObjectVersionsPage(ctx, client, bucket, deletePageOfDeleteMarkers)
	nObjects += n

	if err != nil {
		return nObjects, err
	}

	return nObjects, nil
}

// forEachObjectVersionsPage calls the specified function for each page returned from the S3 ListObjectVersionsPages API.
func forEachObjectVersionsPage(ctx context.Context, client *s3.S3, bucket string, fn func(ctx context.Context, client *s3.S3, bucket string, page *s3.ListObjectVersionsOutput) (int64, error)) (int64, error) {

	var (
		nObjects int64
		lastErr  error
	)

	input := &s3.ListObjectVersionsInput{
		Bucket: aws.String(bucket),
	}

	err := client.ListObjectVersionsPagesWithContext(ctx, input, func(page *s3.ListObjectVersionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		n, err := fn(ctx, client, bucket, page)
		nObjects += n

		if err != nil {
			return false
		}

		return !lastPage
	})

	if err != nil {
		return nObjects, fmt.Errorf("listing S3 Bucket (%s) object versions: %w", bucket, err)
	}

	if lastErr != nil {
		return nObjects, lastErr
	}

	return nObjects, nil
}

// deletePageOfObjectVersions deletes a page (<= 1000) of S3 object versions.
func deletePageOfObjectVersions(ctx context.Context, client *s3.S3, bucket string, page *s3.ListObjectVersionsOutput) (int64, error) {

	var (
		nObjects   int64
		deleteErrs *multierror.Error
	)

	toDelete := make([]*s3.ObjectIdentifier, 0, len(page.Versions))

	for _, v := range page.Versions {
		toDelete = append(toDelete, &s3.ObjectIdentifier{
			Key:       v.Key,
			VersionId: v.VersionId,
		})
	}

	if nObjects = int64(len(toDelete)); nObjects == 0 {
		return nObjects, nil
	}

	// Delete Objects -
	output, err := client.DeleteObjectsWithContext(
		ctx,
		&s3.DeleteObjectsInput{
			Bucket: aws.String(bucket),
			Delete: &s3.Delete{
				Objects: toDelete,       // Array of Keys to Delete
				Quiet:   aws.Bool(true), // Only report errors.
			},
			BypassGovernanceRetention: aws.Bool(true), // Bypass Object Governance Configuration
		})

	if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		return nObjects, nil
	}

	if err != nil {
		return nObjects, fmt.Errorf("deleting S3 Bucket (%s) objects: %w", bucket, err)
	}

	nObjects -= int64(len(output.Errors))

	for _, v := range output.Errors {
		code := aws.StringValue(v.Code)

		if code == s3.ErrCodeNoSuchKey {
			continue
		}

		// Remove any objects with a hold configuration by removing object hold
		if code == ErrCodeAccessDenied {
			key := aws.StringValue(v.Key)
			versionID := aws.StringValue(v.VersionId)

			_, err := client.PutObjectLegalHoldWithContext(ctx, &s3.PutObjectLegalHoldInput{
				Bucket:    aws.String(bucket),
				Key:       aws.String(key),
				VersionId: aws.String(versionID),
				LegalHold: &s3.ObjectLockLegalHold{
					Status: aws.String(s3.ObjectLockLegalHoldStatusOff),
				},
			})

			if err != nil {
				// Add the original error and the new error.
				deleteErrs = multierror.Append(deleteErrs, newDeleteObjectVersionError(v))
				deleteErrs = multierror.Append(deleteErrs, fmt.Errorf("removing legal hold: %w", newObjectVersionError(key, versionID, err)))
			} else {
				// Attempt to delete the object once the legal hold has been removed.
				_, err := client.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
					Bucket:    aws.String(bucket),
					Key:       aws.String(key),
					VersionId: aws.String(versionID),
				})

				if err != nil {
					deleteErrs = multierror.Append(deleteErrs, fmt.Errorf("deleting: %w", newObjectVersionError(key, versionID, err)))
				} else {
					nObjects++
				}
			}
		} else {
			deleteErrs = multierror.Append(deleteErrs, newDeleteObjectVersionError(v))
		}
	}

	if err := deleteErrs.ErrorOrNil(); err != nil {
		return nObjects, fmt.Errorf("deleting S3 Bucket (%s) objects: %w", bucket, err)
	}

	return nObjects, nil
}

// deletePageOfDeleteMarkers deletes a page (<= 1000) of S3 object delete markers.
func deletePageOfDeleteMarkers(ctx context.Context, client *s3.S3, bucket string, page *s3.ListObjectVersionsOutput) (int64, error) {

	var (
		nObjects   int64
		deleteErrs *multierror.Error
	)

	toDelete := make([]*s3.ObjectIdentifier, 0, len(page.Versions))
	for _, v := range page.DeleteMarkers {
		toDelete = append(toDelete, &s3.ObjectIdentifier{
			Key:       v.Key,
			VersionId: v.VersionId,
		})
	}

	if nObjects = int64(len(toDelete)); nObjects == 0 {
		return nObjects, nil
	}

	input := &s3.DeleteObjectsInput{
		Bucket: aws.String(bucket),
		Delete: &s3.Delete{
			Objects: toDelete,
			Quiet:   aws.Bool(true), // Only report errors.
		},
	}

	output, err := client.DeleteObjectsWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		return nObjects, nil
	}

	if err != nil {
		return nObjects, fmt.Errorf("deleting S3 Bucket (%s) delete markers: %w", bucket, err)
	}

	nObjects -= int64(len(output.Errors))

	for _, v := range output.Errors {
		deleteErrs = multierror.Append(deleteErrs, newDeleteObjectVersionError(v))
	}

	if err := deleteErrs.ErrorOrNil(); err != nil {
		return nObjects, fmt.Errorf("deleting S3 Bucket (%s) delete markers: %w", bucket, err)
	}

	return nObjects, nil
}

// newObjectVersionError
func newObjectVersionError(key, versionID string, err error) error {
	if err != nil {
		return fmt.Errorf("S3 object (%s) version (%s): %w", key, versionID, err)
	}
	return nil
}

// newDeleteObjectVersionError
func newDeleteObjectVersionError(err *s3.Error) error {
	if err != nil {
		awsErr := awserr.New(aws.StringValue(err.Code), aws.StringValue(err.Message), nil)
		return fmt.Errorf("deleting: %w", newObjectVersionError(aws.StringValue(err.Key), aws.StringValue(err.VersionId), awsErr))
	}
	return nil
}

func expandVersioning(l []interface{}) *s3.VersioningConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	output := &s3.VersioningConfiguration{}

	if v, ok := tfMap["enabled"].(bool); ok {
		if v {
			output.Status = aws.String(s3.BucketVersioningStatusEnabled)
		} else {
			output.Status = aws.String(s3.BucketVersioningStatusSuspended)
		}
	}

	if v, ok := tfMap["mfa_delete"].(bool); ok {
		if v {
			output.MFADelete = aws.String(s3.MFADeleteEnabled)
		} else {
			output.MFADelete = aws.String(s3.MFADeleteDisabled)
		}
	}

	return output
}

//
func expandVersioningWhenIsNewResource(l []interface{}) *s3.VersioningConfiguration {

	// NOTE - From AWS Provider:
	//
	// Only set and return a non-nil VersioningConfiguration with at least one of
	// MFADelete or Status enabled as the PutBucketVersioning API request
	// does not need to be made for new buckets that don't require versioning.
	// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/4494

	output := &s3.VersioningConfiguration{}

	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	if v, ok := tfMap["enabled"].(bool); ok && v {
		output.Status = aws.String(s3.BucketVersioningStatusEnabled)
	}

	return output
}

// resourceBucketInternalVersioningUpdate
func resourceBucketInternalVersioningUpdate(client *s3.S3, bucket string, versioningConfig *s3.VersioningConfiguration) error {
	_, err := RetryWhenAWSErrCodeEquals(
		2*time.Minute,
		func() (interface{}, error) {
			return client.PutBucketVersioning(&s3.PutBucketVersioningInput{
				Bucket:                  aws.String(bucket),
				VersioningConfiguration: versioningConfig,
			})
		},
		s3.ErrCodeNoSuchBucket,
	)
	return err
}
