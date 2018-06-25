#!/bin/sh

wget https://dist.ipfs.io/ipfs-update/v1.5.2/ipfs-update_v1.5.2_linux-amd64.tar.gz

tar -xvzf ./ipfs-update_v1.5.2_linux-amd64.tar.gz

./ipfs-update/ipfs-update install v0.4.15-rc1

./ipfs-update/install.sh

ipfs init

cp ./ipfs.service /lib/systemd/system/ipfs.service

systemctl daemon-reload

systemctl enable ipfs

rm ./ipfs-update_v1.5.2_linux-amd64.tar.gz

# rm ./ipfs.service
