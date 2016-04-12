#!/bin/bash

set -e

echo "Installing Ansible..."
apt-get install -y --force-yes software-properties-common
apt-add-repository ppa:ansible/ansible
apt-get update
apt-get install -y ansible
cp /vagrant/ansible/ansible.cfg /etc/ansible/ansible.cfg
