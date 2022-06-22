package cortx

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider
func Provider() *schema.Provider {

	return &schema.Provider{

		Schema: map[string]*schema.Schema{
			"cortx_endpoint_host": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("CORTX_ENDPOINT_HOST", nil),
			},
			"cortx_endpoint_port": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CORTX_ENDPOINT_PORT", "80"),
			},
			"cortx_region": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CORTX_REGION", "us-east-1"),
			},
			"cortx_access_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("CORTX_ACCESS_KEY", nil),
			},
			"cortx_secret_access_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("CORTX_SECRET_ACCESS_KEY", nil),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"cortx_bucket": resourceBucket(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"cortx_bucket": datasourceBucket(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

// providerConfigure
func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {

	var (
		diags diag.Diagnostics // Warning or errors can be collected in a slice type
	)

	// Server Endpoint
	cortx_endpoint_host := d.Get("cortx_endpoint_host").(string)
	cortx_endpoint_port := d.Get("cortx_endpoint_port").(string)

	// Server Auth
	cortx_access_key := d.Get("cortx_access_key").(string)
	cortx_secret_access_key := d.Get("cortx_secret_access_key").(string)
	cortx_region := d.Get("cortx_region").(string)

	// Authentication Method - CORTX_ACCESS_KEY and CORTX_SECRET_ACCESS_KEY
	if (cortx_access_key != "") && (cortx_secret_access_key != "") {

		// Initialize a New S3 Client Connection
		sess, err := session.NewSession(
			&aws.Config{
				Credentials: credentials.NewStaticCredentials(cortx_access_key, cortx_secret_access_key, ""),
				Endpoint: aws.String(
					fmt.Sprintf("http://%s:%s", cortx_endpoint_host, cortx_endpoint_port),
				),
				Region:           aws.String(cortx_region),
				DisableSSL:       aws.Bool(true), // Hardcode - Require SSL to be OFF for CORTX!
				S3ForcePathStyle: aws.Bool(true), // Hardcode - Require PathStyle for CORTX!
			},
		)

		// Append Error to Diagnostics && Exit
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to create CORTX S3 client - No Valid Credentials",
				Detail:   "Unable to authenticate user to CORTX S3 Server. Require cortx_access_key and cortx_secret_access_key",
			})
			return nil, diags
		}

		// Create New client
		client := s3.New(sess)

		return client, diags
	}

	// Failed to Provide CORTX_ACCESS_KEY and CORTX_SECRET_ACCESS_KEY
	// Append Error to Diagnostics && Exit
	diags = append(diags, diag.Diagnostic{
		Severity: diag.Error,
		Summary:  "Unable to create CORTX S3 client - No Valid Credentials",
		Detail:   "Unable to authenticate user to CORTX S3 Server. Require cortx_access_key and cortx_secret_access_key",
	})

	return nil, diags

}
