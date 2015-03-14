#!/bin/bash

NAME=gojenkins
GO_BINARY=https://github.com/bndr/gojenkins.git
GO_BINARY_DEST=/opt/$NAME
GO_VERSION=1.4.2
GO_SOURCE=/opt/golang
JENKINS_HOME=/var/lib/jenkins

# Install Jenkins
echo "Adding Jenkins repo..."
wget -q -O - https://jenkins-ci.org/debian/jenkins-ci.org.key | sudo apt-key add -
sudo sh -c 'echo deb http://pkg.jenkins-ci.org/debian binary/ > /etc/apt/sources.list.d/jenkins.list'
echo "Updating packages..."
apt-get update > /dev/null 2>&1
apt-get install -y jenkins git curl > /dev/null 2>&1
echo "Jenkins Installed Successfully"

# Install GO
if [ ! -d "$GO_SOURCE" ]; then
echo "installing Go..."
git clone https://github.com/golang/go $GO_SOURCE
cd $GO_SOURCE && git checkout go$GO_VERSION > /dev/null 2>&1
cd src
./all.bash > /dev/null 2>&1
echo "OK"
fi
echo "Setting ENV parameters"
export PATH=$PATH:$GO_SOURCE/bin
export GOPATH=/tmp
echo "OK"

# Configure Jenkins
cd $JENKINS_HOME/plugins
curl -qO https://updates.jenkins-ci.org/latest/git.hpi
curl --silent -X POST http://localhost:8080/reload > /dev/null 2>&1
sudo su jenkins
ssh-keygen -t rsa -q -N "" -f ~/.ssh/id_rsa

# Compile Go binary
echo "Installing go Binary: $NAME"
if [ ! -d "$GO_BINARY_DEST" ]; then
git clone $GO_BINARY $GO_BINARY_DEST
fi
cd $GO_BINARY_DEST
go get github.com/stretchr/testify/assert
go build
echo "OK"

# Add Test Data to Jenkins
# TODO
# Run Tests
go test