package cortx

//
// Straight From AWS Provider
// https://github.com/hashicorp/terraform-provider-aws/blob/main/internal/service/s3/hosted_zones.go
//

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"strings"
)

// See https://docs.aws.amazon.com/general/latest/gr/s3.html#s3_website_region_endpoints.
var hostedZoneIDsMap = map[string]string{
	endpoints.AfSouth1RegionID:     "Z83WF9RJE8B12",
	endpoints.ApEast1RegionID:      "ZNB98KWMFR0R6",
	endpoints.ApNortheast1RegionID: "Z2M4EHUR26P7ZW",
	endpoints.ApNortheast2RegionID: "Z3W03O7B5YMIYP",
	endpoints.ApNortheast3RegionID: "Z2YQB5RD63NC85",
	endpoints.ApSouth1RegionID:     "Z11RGJOFQNVJUP",
	endpoints.ApSoutheast1RegionID: "Z3O0J2DXBE1FTB",
	endpoints.ApSoutheast2RegionID: "Z1WCIGYICN2BYD",
	endpoints.ApSoutheast3RegionID: "Z01613992JD795ZI93075",
	endpoints.CaCentral1RegionID:   "Z1QDHH18159H29",
	endpoints.CnNorthwest1RegionID: "Z282HJ1KT0DH03",
	endpoints.EuCentral1RegionID:   "Z21DNDUVLTQW6Q",
	endpoints.EuNorth1RegionID:     "Z3BAZG2TWCNX0D",
	endpoints.EuSouth1RegionID:     "Z30OZKI7KPW7MI",
	endpoints.EuWest1RegionID:      "Z1BKCTXD74EZPE",
	endpoints.EuWest2RegionID:      "Z3GKZC51ZF0DB4",
	endpoints.EuWest3RegionID:      "Z3R1K369G5AVDG",
	endpoints.MeSouth1RegionID:     "Z1MPMWCPA7YB62",
	endpoints.SaEast1RegionID:      "Z7KQH4QJS55SO",
	endpoints.UsEast1RegionID:      "Z3AQBSTGFYJSTF",
	endpoints.UsEast2RegionID:      "Z2O1EMRO9K5GLX",
	endpoints.UsGovEast1RegionID:   "Z2NIFVYYW2VKV1",
	endpoints.UsGovWest1RegionID:   "Z31GFT0UA1I2HV",
	endpoints.UsWest1RegionID:      "Z2F56UZL2M1ACD",
	endpoints.UsWest2RegionID:      "Z3BJ6K6RIION7M",
}

// BucketRegionalDomainName -
func BucketRegionalDomainName(bucket string, region string) (string, error) {

	// Return a default AWS Commercial domain name if no region is provided
	// Otherwise EndpointFor() will return BUCKET.s3..amazonaws.com
	if region == "" {
		return fmt.Sprintf("%s.s3.amazonaws.com", bucket), nil //lintignore:AWSR001
	}

	endpoint, err := endpoints.DefaultResolver().EndpointFor(endpoints.S3ServiceID, region)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s.%s", bucket, strings.TrimPrefix(endpoint.URL, "https://")), nil
}
