package cortx

import (
	"context"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"time"
)

const (
	ErrCodeOperationAborted = "OperationAborted"
	ErrCodeBucketNotEmpty   = "BucketNotEmpty"
	ErrCodeAccessDenied     = "AccessDenied"
	ErrCodeNoSuchTagSet     = "NoSuchTagSet"
)

// Retryable is a function that is used to decide if a function's error is retryable or not.
type Retryable func(error) (bool, error)

// TimedOut -
func TimedOut(err error) bool {

	// NOTE: DW 6/15/2022:
	//    - leaving AWS S3 nolint:errorlint annotations, artifact from 1.16 (?)
	// 	  - This does *not* match wrapped TimeoutErrors
	timeoutErr, ok := err.(*resource.TimeoutError) // nolint:errorlint
	return ok && timeoutErr.LastError == nil
}

// Ref: https://github.com/hashicorp/terraform-provider-aws/blob/98ae6c4d53015f266cef2f89094fd6b3e3220888/internal/tfresource/retry.go

// RetryWhenAWSErrCodeEqualsContext retries the specified function when it returns one of the specified AWS error code.
func RetryWhenAWSErrCodeEqualsContext(ctx context.Context, timeout time.Duration, f func() (interface{}, error), codes ...string) (interface{}, error) { // nosemgrep:aws-in-func-name
	return RetryWhenContext(ctx, timeout, f, func(err error) (bool, error) {
		if tfawserr.ErrCodeEquals(err, codes...) {
			return true, err
		}

		return false, err
	})
}

// RetryWhenAWSErrCodeEquals retries the specified function when it returns one of the specified AWS error code.
func RetryWhenAWSErrCodeEquals(timeout time.Duration, f func() (interface{}, error), codes ...string) (interface{}, error) { // nosemgrep:aws-in-func-name
	return RetryWhenAWSErrCodeEqualsContext(context.Background(), timeout, f, codes...)
}

// RetryWhenContext retries the function `f` when the error it returns satisfies `predicate`.
// `f` is retried until `timeout` expires.
func RetryWhenContext(ctx context.Context, timeout time.Duration, f func() (interface{}, error), retryable Retryable) (interface{}, error) {
	var output interface{}

	err := resource.Retry(timeout, func() *resource.RetryError { // nosemgrep: helper-schema-resource-Retry-without-TimeoutError-check
		var err error
		var retry bool

		output, err = f()
		retry, err = retryable(err)

		if retry {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if TimedOut(err) {
		output, err = f()
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}
