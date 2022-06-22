#! /bin/bash

# IAM and Resource Management...
aws iam create-role --role-name vmimport \
	--assume-role-policy-document '{
	   "Version": "2012-10-17",
	   "Statement": [{
	         "Effect": "Allow",
	         "Principal": { "Service": "vmi.amazonaws.com" },
	         "Action": "sts:AssumeRole",
	         "Condition": {
	            "StringEquals":{
	               "sts:Externalid": "vmimport"
	            }
	         }
	      }
	   ]}'

aws iam put-role-policy \
	--role-name vmimport \
	--policy-name vmimport \
	--policy-document '{
	   "Version":"2012-10-17",
	   "Statement":[
	      {
	         "Effect": "Allow",
	         "Action": [
	            "s3:GetBucketLocation",
	            "s3:GetObject",
	            "s3:ListBucket",
	            "s3:GetBucketLocation",
	            "s3:GetObject",
	            "s3:ListBucket",
	            "s3:PutObject",
	            "s3:GetBucketAcl"
	         ],
	         "Resource": [
	            "arn:aws:s3:::*",
	            "arn:aws:s3:::*/*"
	         ]
	      },
	      {
	         "Effect": "Allow",
	         "Action": [
	            "ec2:ModifySnapshotAttribute",
	            "ec2:CopySnapshot",
	            "ec2:RegisterImage",
	            "ec2:Describe*"
	         ],
	         "Resource": "*"
	      },
	      {
			  "Effect": "Allow",
			  "Action": [
			    "kms:CreateGrant",
			    "kms:Decrypt",
			    "kms:DescribeKey",
			    "kms:Encrypt",
			    "kms:GenerateDataKey*",
			    "kms:ReEncrypt*"
			  ],
			  "Resource": "*"
		},
		{
		  "Effect": "Allow",
		  "Action": [
		    "license-manager:GetLicenseConfiguration",
		    "license-manager:UpdateLicenseSpecificationsForResource",
		    "license-manager:ListLicenseSpecificationsForResource"
		  ],
		  "Resource": "*"
		}
	   ]
	}'
