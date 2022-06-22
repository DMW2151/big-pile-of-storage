import boto3
import logging
import argparse

# NOTE: Host variable, but port should be hardcoded on all CloudShare Instances,
# check w. `kubectl svc | grep -E load`
#
# Refer to: https://www.youtube.com/watch?v=iAlNYPU7XOM&t=1320

parser = argparse.ArgumentParser()

parser.add_argument(
    "--hostname",
    dest="hostname",
    type=str,
    default="127.0.0.1",
    help="CLOUDSHARE instance hostname, else localhost, See repo README.MD",
)

parser.add_argument(
    "--port",
    dest="port",
    type=str,
    default="31949",
    help="31949 on CloudShare, else define as appropriate, See repo README.MD",
)

parser.add_argument(
    "--access-key",
    dest="aws_access_key_id",
    type=str,
    default="sgiamadmin",
    help="Hardcoded as 'sgiamadmin' on CloudShare, else get from Terraform Bootstrap Output",
)

parser.add_argument(
    "--secret-key",
    dest="aws_secret_access_key",
    type=str,
    default="ldapadmin",
    help="Hardcoded as 'ldapadmin' on CloudShare, else get from Terraform Bootstrap Output",
)

args = parser.parse_args()


ENDPOINT_URL = args.hostname
ENDPOINT_PORT = args.port
AWS_ACCESS_KEY_ID = args.aws_access_key_id
AWS_SECRET_ACCESS_KEY = args.aws_secret_access_key


# Configure Process and Boto Logging
logging.basicConfig(
    level=logging.WARNING, format=f"%(asctime)s %(levelname)s %(message)s"
)
logger = logging.getLogger()
boto3.set_stream_logger("boto3.resources", logging.WARNING)

# Cnfigure S3 Client Connection
s3_client = boto3.client(
    "s3",
    endpoint_url=f"http://{ENDPOINT_URL}:{ENDPOINT_PORT}",
    aws_access_key_id=AWS_ACCESS_KEY_ID,
    aws_secret_access_key=AWS_SECRET_ACCESS_KEY,
    verify=False,
)

# Check List Buckets
lb_result = s3_client.list_buckets()
print(lb_result)
