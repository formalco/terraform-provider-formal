#!/bin/bash

rm ./dev/.terraform.lock.hcl
rm ./dev/terraform.tfstate
rm ./dev/terraform.tfstate.backup
rm -rf ./dev/.terraform

make install