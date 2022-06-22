# Terraform Provider - Seagate CORTX

<mark>**NOTE:** This repo is part of an entry to the 2022 Seagate CORTX [Hackathon](https://seagate-cortx-hackathon2022.devpost.com/).</mark>

This repo builds a integration between the CORTX API and Hashicorp [Terraform](https://www.terraform.io/). Infrastructure As Code tools like Terraform can allow organizations to standardize infrastructure and policies across multiple regions, environments, applications, etc. With the this provider, users can now include CORTX resources with their other infrastructure deployments rather than relying on custom scripts or one-off deployment patterns. 

This provider is meant for provisioning S3 compatible resources (buckets, objects, access configs) on an already existing CORTX server. While Terraform is a great tool, and can be used for bootstrapping a CORTX server (see `./examples/aws-demo-server-setup/`), this provider is meant for working with S3 API resources, not managing the server itself. The provider is cloud-agnostic, as long as a developer has access to a CORTX public-data endpoint and credentials for that CORTX deployment.

Because the CORTX API is S3 compatible, it is *technically* possible to cobble together a pseudo CORTX provider from resources in the undifferentiated AWS Terraform Provider (e.g. `aws_s3_bucket`, `aws_s3_object`, ...). However, the CORTX API is not at parity with the S3 API ([status](https://seagate-systems.atlassian.net/wiki/spaces/PUB/pages/759333066/CORTX+S3+API+Guide)) and as a result, some of the code paths in AWS' provider can fail on a CORTX deployment (or any 3rd party S3 implementation, i.e. [Issue w. NetAPP](https://github.com/hashicorp/terraform-provider-aws/issues/23291)). This provider is opinionated in that it simplifies AWS' provider, abstracts away the elements of the S3 API that aren't applicable for CORTX, and provides a cleaner interface between Terraform and CORTX than the AWS provider can.



## Building and Testing The Provider

Terraform providers are [Golang](https://go.dev/dl/) modules. If you've got Go >=v1.17 on your machine, you can run the following command to build the provider locally. Please see notes in the MakeFile, you may need to change the `$TARGET_ARCH` and `$TARGET_OS` arguments to be compatible with your system's operating system and architecture before building.

```bash 
make install
```

The resources in the provider are individually unit-tested with `go test`. If you'd like, you can modify the provider's source code and check that all tests are still passing with the following command(s)

```bash 
# To Run Unit Tests 
make test

# To Run Acceptance Tests - See: https://www.terraform.io/plugin/sdkv2/testing/acceptance-tests
make testacc
```

## Examples - Using the Provider

The directory `./examples/**` contains a working example of ways that you might use the CORTX provider with a Terraform module. I've tested this code with v2.0.0 on Cloudshare, and v1.0.4 on my own VM, but because the provider is cloud-agnostic, the modules should work on any CORTX deployment (K8s, private network running OVAs, AWS, GCP, etc.). 

For a quick (and free) test, I'd recommend running the CloudShare example. The AWS example is a bit more involved, involves importing an OVA and bootstrapping your own CORTX node, but it more closely simulates the experience of running your own CORTX server.


### [Optional] AWS Server Setup (v1.0.4)

The AWS example runs the same operations as the CloudShare example, however it has a bit higher barrier to entry. You'll need the following to successfully run this demo.

- An AWS account you have permission to deploy resources to, and comfort with the cost (<mark>**Est. Cost ~$0.25/hr**</mark>) of the resources deployed. Please also read note on billing at end of demo!
- A basic understanding of Terraform, SSH, AWS networking, and IAM permissions

This demo is broken into two components, the CORTX server setup (`./examples/aws-demo-server-setup/`) and the actual provider demo (`./examples/demo/`). If you'd prefer to just test the provider, I would suggest skipping this section and launching a Cloudshare instance.

Let's start with the server setup, `server-setup` provisions a CORTX server and a jump server in a new VPC on AWS. Once it's completed, you'll have a CORTX instance on a  `c5.large` instance. This module automates the actions of the [CORTX EC2 Setup Guide](https://github.com/Seagate/cortx/blob/main/doc/integrations/AWS_EC2/README.md). The contents of `./examples/aws-demo-server-setup/main.tf` assume that the user has already imported an OVA into their AWS account. If you have not yet done so, you can run the following from the terminal to kick that process off: 

```bash
# Create a VMImporter role (ETA: 3s)
sh ./examples/setup-utils/aws/create-iam-role.sh 

# Download OVA and Import to S3 (ETA: 20-30 min) 
# WARN - script can timeout on import, see notes in script for handling timeouts...
sh ./examples/setup-utils/aws/1-0-4-import-cortx-ami.sh $VM_STORAGE_BUCKET 
```

Once an OVA is available, there a few small changes to be made to the boilerplate `main.tf` file included in this example. 

- The `cortx_instance_ssh_key` argument should be updated. You'll need to port forward all requests from your machine to the CORTX server through an intermediate instance. To provision the SSH access for these instances, you'll need a keypair registered with AWS and saved in your `~/.ssh` directory (e.g. in the example I use `~/.ssh/public-jump-1.pem`). Refer to the [AWS docs](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/create-key-pairs.html) if you're unsure of how to do this.

- The configuration block, `backend` is used for specifying where terraform stores a state file with your resource configuration. It's a best practice to store these remotely (e.g. in an S3 bucket). For the demo, you can remove that entire block **OR** change the fields, `bucket`, `region`, and `profile` to your desired configuration.

Once these fields are filled, try running `terraform apply`. This module will run for about 10 minutes and spin up a CORTX VPC as described above. At the end of the `apply`, you should see an output like the following:

```bash
```

If you've deployed everything properly thus far, you should be able to open a seperate terminal and run the port forwarding commands:

- `cmd_cortx_mgmt_port_forward` - Makes the management endpoint from CORTX available at [localhost:28000](https://127.0.0.1:28100/#/preboarding/login/). You can follow the pre-boarding, onboarding, and GUI setup steps provided by Seagate ([link](https://github.com/Seagate/cortx/blob/main/doc/Preboarding_and_Onboarding.rst)) to get an understanding of the management console.

- `cmd_cortx_data_port_forward` - Allows you to connect to the AWS hosted CORTX server's data endpoint from localhost:28001. You can run the following to confirm the connection has been networked properly. In this example, you must populate `$ACCESS_KEY` and `$SECRET_KEY` with the crendentials generated during CORTX server bootstrapping. These will be in the log output of the Terraform apply. 

	```bash
	python3 /examples/setup-utils/common/cortx-server-check-connections.py \
		--hostname 127.0.0.1 --port 28001 --access-key $ACCESS_KEY --secret-key $SECRET_KEY
	```

If you've run through the steps up until this point, you can move to the `./examples/aws-demo/` and run it in the same way you walked through the CloudShare demo. Remember to update the variables in the CORTX provider block!


**WARNING:** One final note, after terminating both the server and the main demo with `terraform destroy`, please check your AWS account to verify no resources are still running. By default, the CORTX OVA **does not delete storage volumes** after you delete the instance, this is a very handy feature, but you still will be billed for this storage! In the AWS console, you'll want to click `Elastic Block Store > Volumes > Actions > Delete Volumes` to delete all hanging volumes associated with this demo. You may also want to delete the OVA from your environment with `AMIs > Images > Actions > Deregister AMI` 
