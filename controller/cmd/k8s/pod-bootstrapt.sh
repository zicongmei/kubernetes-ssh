#!/bin/bash
set -ex

mkdir -p /root/.ssh
chmod 700 /root/.ssh
cat /tmp/ssh/id_rsa > /root/.ssh/id_rsa
cat /tmp/ssh/id_rsa.pub > /root/.ssh/id_rsa.pub
cat /tmp/ssh/authorized_keys > /root/.ssh/authorized_keys
chmod 600 /root/.ssh/id_rsa
chmod 640 /root/.ssh/id_rsa.pub
chmod 600 /root/.ssh/authorized_keys
