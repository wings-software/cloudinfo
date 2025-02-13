#!/bin/bash
# Copyright 2025 Harness Inc. All rights reserved.
# Use of this source code is governed by the PolyForm Free Trial 1.0.0 license
# that can be found in the licenses directory at the root of this repository, also available at
# https://polyformproject.org/wp-content/uploads/2020/05/PolyForm-Free-Trial-1.0.0.txt.

###################################################
# Purpose
# The purpose of this script is to facilitate auto-tagging of
# Jira tickets with fix-versions. There are two tricky parts to auto-tagging:
#
# 1. Given a change set (PR diff), what constitutes a change to a service?
# 2. Given you've determined a set of files has changed a service, which service was changed?
#
# As this repository is a single-service repository, there is no concern with number 2 above.
#
# This script endeavors to answer question number one - which files constitute a change to a service
# Given an input file which is the git diff from a PR
# this script should determine what file changes in the diff
# constitute a material change to a service. At the time
# of this writing, this currently only identifies java and go
# files, and if those files are in the change list, then those
# file names are returned in the output file.
#
# Inputs
# $1 - the git diff from the pull request
# $2 - the output file which should ultimately contain file names from the diff that affect a service
#
# See BT-10437 for more information
# See BT-11220 for more information
# See https://harness.atlassian.net/wiki/spaces/ENGOPS/pages/22218015100/Multi-service+repositories+-+Mapping+file+changes+to+affected+services+for+CX+Tagging+Communication
#
# Owner: Engops
# Author: Marc Batchelor
###################################################
uniqueFileNamesFile=$1
sourceDiffNames=$2
if [ -z "$uniqueFileNamesFile" ]; then
  echo "Missing input PR Difference file."
  exit 1
fi
if [ ! -f "$uniqueFileNamesFile" ]; then
  echo "Input file $uniqueFileNamesFile does not exist and is required."
  exit 2
fi
if [ -z "$sourceDiffNames" ]; then
  echo "Missing output file."
  exit 3
fi
if [ ! -f "$sourceDiffNames" ]; then
  echo "File $sourceDiffNames does not exist and is required."
  exit 4
fi

##### Detect git diff file, or processed filenames only
isDiffFile=$(grep -E "^diff --git a\/" "$uniqueFileNamesFile" | wc -l)
if [ $isDiffFile -gt 0 ]; then
  echo "Received a diff file... fix it to be a filenames only file"
  fileNamesOnlyVar=$(cat "$uniqueFileNamesFile"|grep -E "^diff --git" | sed 's/diff --git a\///' | sed 's/ b\/.*$//' | sort -u)
  echo -e "$fileNamesOnlyVar">"$uniqueFileNamesFile"
fi

# Java files (and other files) which end up in jars - these are kept in .../src/main/x/x/x/* 
cat "$uniqueFileNamesFile" | grep -E ".*java$" | grep -v "/test/" > $sourceDiffNames
# go files (without tests)
cat "$uniqueFileNamesFile" | grep -E ".*.go$|.*.mod$" | grep -v "test_" >> $sourceDiffNames
# Other potential source files
cat "$uniqueFileNamesFile" | grep -E ".*.(css|gradle|graphql|gv|html|iml|js|json|mustache|pl|properties|proto|ps1|py|pyc|qbg|repo|rs|sh|sha256|sql|sum|tf|tgz|tpl|xml|yaml|yml)$" >> $sourceDiffNames
