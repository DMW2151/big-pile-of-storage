#! /bin/bash

#
# Create a VM importer role to download a CORTX OVA from GitHub Releases Page 
# into S3 and then creates an AMI in EC2.
#
# bash ./1-0-4-import-cortx-ami.sh dmw2151-vm-storage
#

# Set a few parameters...
VM_STORAGE_BUCKET=$1
VM_DESCRIPTION="Seagate CORTX OVA 1.0.4"
EC2_VM_AMI_NAME="seagate_cortx_ova_1_0_4"
CLI_DEBUG_FLAGS=""

# Create a Target Bucket for VM Storage if DNE
aws s3api create-bucket --bucket $VM_STORAGE_BUCKET


# Check if out target OVA exists at this location ??
FILE_EXISTS=`(aws s3 ls s3://$VM_STORAGE_BUCKET/ova/cortx-ova-1.0.4.ova)`

if [ -z "$FILE_EXISTS" ]; # If empty, then download and fire import
then 
	echo "Downloading CORTX OVA v1.0.4..."
	# Download the CORTX OVA from Github Releases Page (`https://github.com/Seagate/cortx/releases`) 
	# Push result directly to VM storage bucket
	wget -qO- https://github.com/Seagate/cortx/releases/download/cortx-ova-1.0.4-632/cortx-ova-1.0.4.ova |\
	aws s3 cp - s3://$VM_STORAGE_BUCKET/ova/cortx-ova-1.0.4.ova

	# Import the OVA from S3 to EC2 - Using ImageBuilder, create an AMI. This is a *very* 
	# long-running call. Will return a status immediatley, but may take ~15-30 minutes
	echo "Start AMI Import..."
	aws ec2 import-image $CLI_DEBUG_FLAGS --description $VM_DESCRIPTION \
		--disk-containers Format=OVA,Description=$VM_DESCRIPTION,UserBucket="{S3Bucket=$VM_STORAGE_BUCKET,S3Key=ova/cortx-ova-1.0.4.ova}"
fi

# NOTE: `aws ec2 wait` maxes out at 40 attempts w. 15 seconds between polling. May take 
# longer than 10 minutes for the import to complete, just run the script again. On a 
# subsequentattempt this will pass quickly, just waiting for an image w. these params 
# to, not a specific import task ID
echo "Waiting on AMI Import..."
aws ec2 wait image-exists $CLI_DEBUG_FLAGS --no-include-deprecated --owners self\
	--filter "Name=description,Values=$VM_DESCRIPTION"

# NOTE: Replace the imported image with a named image - The import job doesn't preserve the
# name, copy the AMI into a named AMI and deregisted the old, unnamed AMI
SRC_IMG_ID=$(
	aws ec2 describe-images $CLI_DEBUG_FLAGS --owners self \
	--filter "Name=description,Values=$VM_DESCRIPTION" | jq '.Images[0].ImageId'
)

echo "Renaming AMI to $EC2_VM_AMI_NAME"
# If there is a source image, then copy, rename, && deregister the "older" version
if [ "$SRC_IMG_ID" ];
then
	aws ec2 copy-image --source-image-id $SRC_IMG_ID \
		--name $EC2_VM_AMI_NAME \
		--source-region "us-east-1"\
		--description $VM_DESCRIPTION

	aws ec2 deregister-image --image-id $SRC_IMG_ID
fi