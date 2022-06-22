#!/bin/bash


# Assumes C5.Large Instance Type
# See: https://github.com/Seagate/cortx/blob/main/doc/integrations/AWS_EC2/README.md
#
# This section runs once, only on first boot to create stable interfaces for CORTX

mkdir -p /etc/udev/rules.d/
touch /etc/udev/rules.d/70-persistent-net.rules

MGMT_MAC=$(ip addr show ens5 | grep -oP '(?<=ether\s)[0-9a-z]{2}(:[0-9a-z]{2}){5}')
PUBLIC_MAC=$(ip addr show ens6 | grep -oP '(?<=ether\s)[0-9a-z]{2}(:[0-9a-z]{2}){5}')
PRIVATE_MAC=$(ip addr show ens7 | grep -oP '(?<=ether\s)[0-9a-z]{2}(:[0-9a-z]{2}){5}')

echo """SUBSYSTEM==\"net\", ACTION==\"add\", DRIVERS==\"?*\", ATTR{address}==\"$MGMT_MAC\", NAME=\"ens32\"
SUBSYSTEM==\"net\", ACTION==\"add\", DRIVERS==\"?*\", ATTR{address}==\"$PUBLIC_MAC\", NAME=\"ens33\"
SUBSYSTEM==\"net\", ACTION==\"add\", DRIVERS==\"?*\", ATTR{address}==\"$PRIVATE_MAC\", NAME=\"ens34\"""" > /etc/udev/rules.d/70-persistent-net.rules


echo "**** Reassigned Network Interfaces *****"
cat /etc/udev/rules.d/70-persistent-net.rules
echo "****************************************"

sudo shutdown -r +1