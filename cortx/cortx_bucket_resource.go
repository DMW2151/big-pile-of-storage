package cortx

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
	"net/http"
)

//
func resourceBucket() *schema.Resource {

	return &schema.Resource{
		CreateContext: resourceBucketCreate,
		ReadContext:   resourceBucketRead,
		UpdateContext: resourceBucketUpdate,
		DeleteContext: resourceBucketDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:         schema.TypeString,
				Optional:     true, // False (?)
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 63),
			},
			// Remove for Clarity (`bucket_prefix`) - Make Name Required
			"bucket_domain_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bucket_regional_domain_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"force_destroy": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"object_lock_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true, // Can be removed when object_lock_configuration.0.object_lock_enabled is removed
				ForceNew: true,

				// NOTE: Remove ConflictsWith `object_lock_configuration` - Should NOT set
				// `object_lock_configuration` on object init
			},
			"tags":     TagsSchema(),
			"tags_all": TagsSchemaComputed(),
		},
	}
}

func resourceBucketCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var (
		diags  diag.Diagnostics
		bucket string
	)

	client := meta.(*s3.S3)

	// Get The ID of the Bucket - One of a Few Possible Methods
	if v, ok := d.GetOk("bucket"); ok {
		bucket = v.(string)
	}

	// Init Create Request
	createRequest := &s3.CreateBucketInput{
		Bucket:                     aws.String(bucket),
		ObjectLockEnabledForBucket: aws.Bool(d.Get("object_lock_enabled").(bool)),
	}

	//
	// Try to Create a Bucket w. Retry
	//
	err := resource.Retry(bucketCreatedTimeout, func() *resource.RetryError {
		_, err := client.CreateBucket(createRequest)
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == ErrCodeOperationAborted {
				return resource.RetryableError(
					fmt.Errorf("error creating S3 Bucket (%s), retrying: %w", bucket, err),
				)
			}
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	// Try once more after the TimeOut
	if TimedOut(err) {
		_, err = client.CreateBucket(createRequest)

		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  fmt.Sprintf("[WARN] Waiting on CreateBucket (%s):", bucket),
			Detail:   fmt.Sprintf("[WARN] %v", err),
		})
	}

	// Assign the bucket name as the resource ID
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("[ERROR] Failed on CreateBucket (%s):", bucket),
			Detail:   fmt.Sprintf("[ERROR] %v", err),
		})
		return diags
	}

	d.SetId(bucket)
	return resourceBucketUpdate(ctx, d, meta)
}

// resourceBucketRead -
func resourceBucketRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	client := meta.(*s3.S3)

	headBucketInp := &s3.HeadBucketInput{
		Bucket: aws.String(d.Id()),
	}

	err := resource.Retry(bucketCreatedTimeout, func() *resource.RetryError {

		_, err := client.HeadBucket(headBucketInp)

		if d.IsNewResource() && tfawserr.ErrStatusCodeEquals(err, http.StatusNotFound) {
			return resource.RetryableError(err)
		}

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if TimedOut(err) {
		_, err = client.HeadBucket(headBucketInp)
	}

	// Failed to Get Bucket - Return Diagnostics
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("[ERROR] Failed on GetHeadBucket (%s):", d.Id()),
			Detail:   fmt.Sprintf("[ERROR] %v", err),
		})
		return diags
	}

	// TODO: Set Parameters Back to the Resource Data //

	return diags
}

// resourceBucketUpdate -
func resourceBucketUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	// NOTE: From AWS Provider - Order of argument updates below is important
	// Hashicorp/AWS set this in the main S3 resource definition, allow it to
	// stay unchanged
	//
	// All references to a generic block follow the pattern below, given in the original
	// provider
	//
	// if d.HasChange("generic_attribute") {
	// 		if err := resourceBucketSomeGenericAttributeUpdate(conn, d); err != nil {
	// 			return fmt.Errorf("(%d):(%v)", d.Id(), err)
	// 		}
	// }
	//
	// Removed Generic Blocks for `policy`, `cors_rule`, `website`, `acl`, `grant`,
	// `logging`, `lifecycle_rule`,  `acceleration_configuration`, `request_payer`,
	// `replication_configuration`, `server_side_encryption_configuration`, and
	// `object_lock_configuration`, leaving only `versioning` (between `website` and `acl`)

	var diags diag.Diagnostics
	client := meta.(*s3.S3)

	if d.HasChange("versioning") {
		v := d.Get("versioning").([]interface{})

		if d.IsNewResource() {
			if versioning := expandVersioningWhenIsNewResource(v); versioning != nil {
				err := resourceBucketInternalVersioningUpdate(client, d.Id(), versioning)
				if err != nil {
					// Update Diags
					diags = append(diags, diag.Diagnostic{
						Severity: diag.Error,
						Summary:  fmt.Sprintf("[ERROR] Failed on Update (%s):", d.Id()),
						Detail:   fmt.Sprintf("[ERROR] %v", err),
					})
					return diags
				}
			}
		} else {
			if err := resourceBucketInternalVersioningUpdate(client, d.Id(), expandVersioning(v)); err != nil {
				// Update Diags
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("[ERROR] Failed on Update (%s):", d.Id()),
					Detail:   fmt.Sprintf("[ERROR] %v", err),
				})
				return diags
			}
		}
	}

	return resourceBucketRead(ctx, d, meta)
}

// resourceBucketDelete -
func resourceBucketDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	client := meta.(*s3.S3)

	_, err := client.DeleteBucketWithContext(ctx, &s3.DeleteBucketInput{
		Bucket: aws.String(d.Id()),
	})

	// Delete Bucket Not Possible - Bucket DNE
	// This is OK - Successful "Delete" - Changed Outside of Plan
	if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		return diags
	}

	// In Force Delete Mode we delete *EVERYTHING* including locked objects.
	// CORTX object lock status ss
	if tfawserr.ErrCodeEquals(err, ErrCodeBucketNotEmpty) {

		if d.Get("force_destroy").(bool) {
			log.Printf("[DEBUG] S3 Bucket attempting to forceDestroy %s", err)
			if n, err := EmptyBucket(ctx, client, d.Id()); err != nil {
				return diag.Errorf("emptying S3 Bucket (%s): %s", d.Id(), err)
			} else {
				log.Printf("[DEBUG] Deleted %d S3 objects", n)
			}

			// Recurses until all objects are deleted or an error is returned
			return resourceBucketDelete(ctx, d, meta)
		}
	}

	// Delete Failed - Report Error w. Diagnostics
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("[ERROR] Failed on DeleteBucket (%s):", d.Id()),
			Detail:   fmt.Sprintf("[ERROR] %v", err),
		})
		return diags
	}

	return diags
}
