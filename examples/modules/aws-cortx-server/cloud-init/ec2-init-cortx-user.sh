#!/bin/bash


# Cleanup - Not Used In Boostrap Process - From S3 Sanity Check Script
cleanup() {

    output=$(
      s3iamcli resetaccountaccesskey -n TerraformTestAccount \
        --ldapuser sgiamadmin --ldappasswd "$ldappasswd"
    )
  
    access_key=$(echo -e "$output" | tr ',' '\n' | grep "AccessKeyId" | awk '{print $3}')
    secret_key=$(echo -e "$output" | tr ',' '\n' | grep "SecretKey" | awk '{print $3}')
    
    # Delete User && Remove Account...
    s3iamcli deleteuser -n CortxTerraformTestUser \
      --access_key $access_key --secret_key $secret_key

    s3iamcli deleteaccount -n TerraformTestAccount \
      --access_key $access_key --secret_key $secret_key
}


# Update `/root/.sgs3iamcli/config.yaml` and `/root/.s3cfg`, the S3iamcli configuration file && 
# S3cmd configuration file
update_config() {

    # Set Endpoint
    cortx_s3_endpoint="127.0.0.1"

    # Smoke Test for Cert
    s3iamcli listaccounts --ldapuser sgiamadmin --ldappasswd "$ldappasswd"

    # Update `/root/.sgs3iamcli/config.yaml` and `/root/.s3cfg`, the S3iamcli configuration file && 
    # S3cmd configuration file
    sed -i "s/IAM:.*/IAM: http:\/\/$cortx_s3_endpoint:9080/g" /root/.sgs3iamcli/config.yaml
    sed -i "s/IAM_HTTPS:.*/IAM_HTTPS: https:\/\/$cortx_s3_endpoint:9443/g" /root/.sgs3iamcli/config.yaml
    sed -i "s/VERIFY_SSL_CERT:.*/VERIFY_SSL_CERT: false/g" /root/.sgs3iamcli/config.yaml
    sed -i "s/host_base =.*/host_base = $cortx_s3_endpoint/g" /root/.s3cfg
}

# Create a Non LDAPAdmin User for Testing
create_user_if_dne() {

  output=$(
    s3iamcli createaccount -n TerraformTestAccount \
        -e TerraformTestAccount@dmw2151.com \
        --ldapuser sgiamadmin --ldappasswd "$ldappasswd"
  )

  output=$(
    s3iamcli resetaccountaccesskey -n TerraformTestAccount \
      --ldapuser sgiamadmin --ldappasswd "$ldappasswd"
  )

  access_key=$(echo -e "$output" | tr ',' '\n' | grep "AccessKeyId" | awk '{print $3}')
  secret_key=$(echo -e "$output" | tr ',' '\n' | grep "SecretKey" | awk '{print $3}')

  s3iamcli CreateUser -n CortxTerraformTestUser \
    --access_key $access_key --secret_key $secret_key

  # Developer Note 
  #
  # Again - Awful Security and Dev Experience - but this is JUST a demo,
  #
  # TODO: As of writing, cannot figure out how to pipe the result of a remote 
  # exec back to a terraform variable

  echo "****************************************"
  echo "****************************************"
  echo "AWS_ACCESS_KEY:" $access_key
  echo "AWS_SECRET_ACCESS_KEY:" $secret_key
  echo "****************************************"
  echo "****************************************"

}


# Get LDAP Credentials + Cert && Creae Test User
if rpm -q "salt"  > /dev/null; then 
  ldappasswd=$(salt-call pillar.get openldap:iam_admin:secret --output=newline_values_only)
  ldappasswd=$(salt-call lyveutil.decrypt openldap ${ldappasswd} --output=newline_values_only)
fi

update_config && create_user_if_dne
