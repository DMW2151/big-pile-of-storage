#!/bin/bash


# Developer Notes
#
# NOTE: `bootstrap.sh` is deterministic, not idempotent, second boot still
# takes 5+ minutes
#
# NOTE: Check if Consul is Running -> If No, then boostrap the node w. 
# new symlinks to /dev/sdb and /dev/sdc
#
# WARN: Running CORTX bootstrap as root, this is introduces some security 
# concerns on a real system, but since CORTX isn't meant to be run on AWS 
# for real, tolerate it..
#
# WARN/TODO: Set the following content to always run on reboot via systemd
#

ln -s /dev/nvme1n1 /dev/sdb && ln -s /dev/nvme2n1 /dev/sdc

sh /opt/seagate/cortx/provisioner/cli/virtual_appliance/bootstrap.sh
sh /opt/seagate/cortx/s3/scripts/s3-sanity-test.sh -e 127.0.0.1

exit 0
