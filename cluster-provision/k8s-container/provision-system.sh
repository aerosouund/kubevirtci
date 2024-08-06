#!/bin/bash
nmcli connection add type ethernet ifname enp0s2 con-name enp0s2 ipv4.method auto ipv6.method link-local
nmcli connection modify enp0s2 connection.autoconnect yes
nmcli connection up enp0s2
sudo ostree admin unlock --hotfix