#!/bin/bash

# This script generates Terraform provider documentation using tfplugindocs

# Create a docs directory if it doesn't exist
mkdir -p docs

# Run the tfplugindocs tool to generate documentation
go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate \
  --provider-dir="." \
  --provider-name="multicdn" \
  --rendered-provider-name="Constellix MultiCDN"
