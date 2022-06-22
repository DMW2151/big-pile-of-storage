package cortx

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strings"
	"time"
)

// datasourceObject
func datasourceObject() *schema.Resource {

	return &schema.Resource{
		ReadContext: datasourceObjectRead,
		Schema: map[string]*schema.Schema{
			"body": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bucket": {
				Type:     schema.TypeString,
				Required: true,
			},
			"bucket_key_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"cache_control": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"content_disposition": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"content_encoding": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"content_language": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"content_length": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"content_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"expiration": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"expires": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"key": {
				Type:     schema.TypeString,
				Required: true,
			},
			"last_modified": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"metadata": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"object_lock_legal_hold_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"object_lock_mode": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"object_lock_retain_until_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"range": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"server_side_encryption": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"sse_kms_key_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"storage_class": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"website_redirect_location": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": TagsSchemaComputed(),
		},
	}
}

func datasourceObjectRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics
	client := meta.(*s3.S3)

	// Bucket
	bucket := d.Get("bucket").(string)
	key := d.Get("key").(string)

	input := s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	// Check For Bytes Range
	if v, ok := d.GetOk("range"); ok {
		input.Range = aws.String(v.(string))
	}

	// Check for Object Version ID
	if v, ok := d.GetOk("version_id"); ok {
		input.VersionId = aws.String(v.(string))
	}

	out, err := client.HeadObject(&input)

	// Failed on GetHead Object
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("[ERROR] Failed getting Bucket (%s) Object (%s)", bucket, key),
			Detail:   fmt.Sprintf("[ERROR] %w", err),
		})
		return diags
	}

	// Object Has Version Been Deleted In Intermediary
	if aws.BoolValue(out.DeleteMarker) {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("[ERROR] Requested Object (%s%s%s) has been deleted", bucket, key, input.VersionId),
			Detail:   fmt.Sprintf("[ERROR] %w", err),
		})
		return diags
	}

	// Set All Reader Params - Some removed to enforce compatbility with the
	// CORTX API Implementation
	d.SetId(fmt.Sprintf("%s/%s%@s", bucket, key, input.VersionId))
	d.Set("bucket_key_enabled", out.BucketKeyEnabled)
	d.Set("cache_control", out.CacheControl)
	d.Set("content_disposition", out.ContentDisposition)
	d.Set("content_encoding", out.ContentEncoding)
	d.Set("content_language", out.ContentLanguage)
	d.Set("content_length", out.ContentLength)
	d.Set("content_type", out.ContentType)
	d.Set("etag", strings.Trim(aws.StringValue(out.ETag), `"`))
	d.Set("metadata", PointersMapToStringList(out.Metadata))

	// Removed Expiration - Concept doesn't make sense in the context of CORTX
	//
	// d.Set("expiration", out.Expiration)
	// d.Set("expires", out.Expires)

	// Removed Legal Hold - Concept doesn't make sense in the context of CORTX
	//
	// d.Set("object_lock_legal_hold_status", out.ObjectLockLegalHoldStatus)
	// d.Set("object_lock_mode", out.ObjectLockMode)
	// d.Set("object_lock_retain_until_date", flattenObjectDate(out.ObjectLockRetainUntilDate))

	// Removed SSE - Concept doesn't make sense in the context of CORTX
	//
	// d.Set("server_side_encryption", out.ServerSideEncryption)
	// d.Set("sse_kms_key_id", out.SSEKMSKeyId)

	// Removed Storage Class - Concept doesn't make sense in the context of CORTX
	//
	// d.Set("storage_class", s3.StorageClassStandard)

	d.Set("version_id", out.VersionId)
	d.Set("website_redirect_location", out.WebsiteRedirectLocation)

	// Set Last Modified for Reader - if new object, then no last modified
	// timestamp -> skip
	if out.LastModified != nil {
		d.Set("last_modified", out.LastModified.Format(time.RFC1123))
	}

	// Remove Concept of `Body` -> This is kinda an Anti-Pattern in the original provider
	// why keep body for large files, => assumes user won't upload large files, need to be
	// able to make that assumption w. CORTX

	// TODO: This is a valid feature - want to implement tag mnagement
	// tags, err := ObjectListTags(client, bucket, key)

	// if err != nil {
	// 	diags = append(diags, diag.Diagnostic{
	// 		Severity: diag.Error,
	// 		Summary:  fmt.Sprintf("[ERROR] Failed listing tags for S3 Bucket (%s) Object (%s)", bucket, key),
	// 		Detail:   fmt.Sprintf("[ERROR] %w", err),
	// 	})
	// 	return diags
	// }

	// if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
	// 	diags = append(diags, diag.Diagnostic{
	// 		Severity: diag.Error,
	// 		Summary:  fmt.Sprintf("[ERROR] Failed setting tags"),
	// 		Detail:   fmt.Sprintf("[ERROR] %w", err),
	// 	})
	// 	return diags
	// }

	return diags
}
