#!/bin/bash
sudo apt update
sudo apt upgrade -y
sudo apt install -y podman zsh curl git
# Fix subuid/subgid for rootless Podman inside a devcontainer.
# The container's uid_map only covers 0-65536, so we must use a subrange
# of that instead of the default 100000+ range.
sudo bash -c 'echo "vscode:10000:55537" > /etc/subuid && echo "vscode:10000:55537" > /etc/subgid'
podman system migrate
go mod download

