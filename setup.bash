#!/bin/bash

## Setting up the instance and pulling data needed to implement the paper.
## Tested on EC2 AMI-linux instance r3.large + EBS SSD 300 GB

echo "Setting up..."

cd $HOME
LOCAL_BIN=$HOME/bin
mkdir $LOCAL_BIN

echo
echo "Mounting the second EBS volume..."
# Based on: http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ebs-using-volumes.html#using-ebs-volume-linux
lsblk
read -p "What's the name of your 300GB+ EBS volume? (e.g /dev/xvdb)" dn
DEVICE_NAME=$dn
DATA_DIR=/data
sudo mkfs -t ext3 $DEVICE_NAME
sudo mkdir $DATA_DIR
sudo mount $DEVICE_NAME $DATA_DIR
sudo chown ec2-user $DATA_DIR

echo
echo "Installing s3cmd..."
curl -o s3cmd.tar.gz -L https://github.com/s3tools/s3cmd/archive/v1.5.0-rc1.tar.gz
tar -zxvf s3cmd.tar.gz
rm s3cmd.tar.gz
cd s3cmd-1.5.0-rc1/
sudo python setup.py install
cd $HOME
s3cmd --configure

echo
echo "Installing pip..."
sudo easy_install pip

echo
echo "Installing nltk..."
sudo pip install -U nltk
sudo python -m nltk.downloader -d /usr/share/nltk_data all

echo
echo "Installing git..."
sudo yum -y install git
git config --global user.name "EC2 User"
git config --global user.email ec2@user.com

echo
echo "Installing Mercurial..."
sudo yum -y install mercurial

echo
echo "Installing Go..."
curl -o go.tar.gz -L https://storage.googleapis.com/golang/go1.3.1.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go.tar.gz
rm go.tar.gz
export PATH=$PATH:/usr/local/go/bin
mkdir $HOME/go
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin

#echo
#echo "Installing Bolt database..."
#go get github.com/boltdb/bolt/...

echo
echo "Installing leveldb..."
curl -L -o leveldb.rpm http://dl.fedoraproject.org/pub/epel/6/x86_64/leveldb-1.7.0-2.el6.x86_64.rpm
yum -y install leveldb.rpm

echo
echo "Installing and configuring Cayley database..."
curl -o $LOCAL_BIN/cayley.tar.gz -L https://github.com/google/cayley/releases/download/v0.4.0/cayley_0.4.0_linux_amd64.tar.gz
cd $LOCAL_BIN
tar -zxvf cayley.tar.gz
rm cayley.tar.gz
export PATH=$PATH:$LOCAL_BIN/cayley_0.4.0_linux_amd64
cd $HOME
mkdir config
curl -L -o cayley.cfg https://github.com/srom/ensu/blob/master/cayley.cfg
echo "Testing Cayley..."
cayley init --config=$HOME/config/cayley.cfg # test if everything's fine
go get github.com/google/cayley
cd $GOPATH/src/github.com/google/cayley
git checkout v0.4.0
cd $HOME

go get github.com/srom/xmlstream
go get github.com/kennygrant/sanitize

echo "Installing Pyley, Python client for Cayley"
cd $HOME
git clone https://github.com/ziyasal/pyley

echo
echo "Download Freebase data (might take a while)..."
mkdir /data/freebase
s3cmd get --recursive s3://basekb-now/2014-08-24-00-00/sieved/ /data/freebase --add-header=x-amz-request-payer:requester
rm /data/freebase/*$folder$

echo
echo "Getting parlparse data (might take a while)..."
mkdir /data/parlparse
cd /data/parlparse
git clone https://github.com/mysociety/parlparse
mv parlparse/members ./people
rm -rf parlparse/
rsync -az --progress --exclude '.svn' --exclude 'tmp/' --relative data.theyworkforyou.com::parldata/scrapedxml/* /data/parlparse/
mv scrapedxml/ documents/
cd $HOME

echo
echo "Done"
