#!/bin/bash
set -ex

export PATH=~/go/bin:$PATH

make --directory=smb-volume-k8s-release/log-forwarder test
