package cortx

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// datasourceBucket
//
// Uses the official AWS compatible definition for an S3 bucket
// Updated as of 6/15/2022 (Hash: 15d4edf98b4fb6c40a73397ee7504f6e61ec3574)
//
// See: https://github.com/hashicorp/terraform-provider-aws/blob/main/internal/service/s3/bucket_data_source.go
//
func datasourceBucket() *schema.Resource {

	return &schema.Resource{
		ReadContext: datasourceBucketRead,
		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:     schema.TypeString,
				Required: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bucket_domain_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bucket_regional_domain_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"hosted_zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"region": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"website_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"website_domain": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func datasourceBucketRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics
	client := meta.(*s3.S3)

	// Get user provided bucket name and validate existence
	bucket := d.Get("bucket").(string)

	_, err := client.HeadBucket(
		&s3.HeadBucketInput{
			Bucket: aws.String(bucket),
		},
	)

	// Failed to Get Bucket - Return Diagnostics
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("[ERROR] Failed on GetHeadBucket (%s):", bucket),
			Detail:   fmt.Sprintf("[ERROR] %w", err),
		})
		return diags
	}

	// Set DataSource Attributes - Name, ID, ARN, Location, etc.
	d.SetId(bucket)

	// Synthetic ARN using CORTX as provider Partition
	d.Set("arn", arn.ARN{
		Partition: endpoints.AwsPartition().ID(),
		Service:   "s3",
		Resource:  bucket,
	}.String())

	// FROM AWS PROVIDER SOURCE: By default, GetBucketRegion forces virtual host
	// addressing, which is not compatible with many non-AWS implementations. Instead,
	// pass the provider s3_force_path_style configuration, which defaults to false
	region, err := s3manager.GetBucketRegionWithClient(context.Background(), client, bucket, func(r *request.Request) {
		r.Config.S3ForcePathStyle = client.Config.S3ForcePathStyle
		r.Config.Credentials = client.Config.Credentials
	})

	// Region
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("[ERROR] Failed on GetBucketRegionwithClient (%s):", bucket),
			Detail:   fmt.Sprintf("[ERROR] %w", err),
		})
		return diags
	}

	d.Set("region", region)

	// Hosted Zone
	if hostedZoneID, ok := hostedZoneIDsMap[region]; !ok {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("[ERROR] Failed on GetBucketHostedZone (%s):", bucket),
			Detail:   fmt.Sprintf("[ERROR] %w", err),
		})
		return diags
	} else {
		d.Set("hosted_zone_id", hostedZoneID)
	}

	// Website Endpoint Params - Hypothetical, as of 6/15/2022 - the following requests
	// would fail on CORTX with MethodNotAllowed, see the full list @
	// https://seagate-systems.atlassian.net/wiki/spaces/PUB/pages/759333066/CORTX+S3+API+Guide
	//
	// aws s3api get-bucket-website --bucket test-bucket --profile cloudshare \
	//	--endpoint-url http://$PATH.vm.cld.sr:31949
	//
	// aws s3 website s3://test-bucket --profile cloudshare \
	//	--endpoint-url http://$PATH.vm.cld.sr:31949
	//
	//

	domain := fmt.Sprintf("s3-website.%s.%s", region, endpoints.AwsPartition().DNSSuffix())

	if err := d.Set("website_endpoint", fmt.Sprintf("%s.%s", bucket, domain)); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("[ERROR] Failed on GetWebsiteEndpoint (%s):", bucket),
			Detail:   fmt.Sprintf("[ERROR] %w", err),
		})
		return diags
	}

	if err := d.Set("website_domain", domain); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("[ERROR] Failed on GetWebsiteEndpoint (%s):", bucket),
			Detail:   fmt.Sprintf("[ERROR] %w", err),
		})
		return diags
	}

	return diags
}
