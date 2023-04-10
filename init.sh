
# kavachain deployment script

```
sudo apt update
sudo apt upgrade -y
sudo apt install build-essential jq -y
```

## Install Golang:

## Install latest go version https://golang.org/doc/install
```
wget -q -O - https://raw.githubusercontent.com/canha/golang-tools-install-script/master/goinstall.sh | bash -s -- --version 1.18
source ~/.profile
```

## to verify that Golang installed
```
go version
```

## Install the executables

make install

sudo rm -rf ~/.kava

kava init --chain-id=kava_1-1 validator1

kava keys add validator1 --keyring-backend os
echo "claim either tribe mercy genre drastic stamp spring attend ready believe material hedgehog space remind valley give slight cram arm release universe hybrid abuse" | kava keys add validator1 --keyring-backend test --recover

kava add-genesis-account $(kava keys show validator1 -a --keyring-backend os) 160000000ukava 

kava keys add account2
echo "require resist steak energy armed prison embody abuse huge submit host subway merit kiwi inherit distance cliff suffer general program connect link employ crew" | kava keys add account1 --keyring-backend test --recover

kava add-genesis-account $(kava keys show account1 -a --keyring-backend os) 2000000000ukava

kava gentx validator1 50000000ukava --keyring-backend os --chain-id kava_1-1

kava collect-gentxs

sed -i 's/stake/ukava/g' ~/.kava/config/genesis.json

# cd ~/.kava/config
# jq '.app_state.slashing.params.min_signed_per_window = "0.050000000000000000"' genesis.json > temp.json && mv temp.json genesis.json
# jq '.app_state.slashing.params.slash_fraction_double_sign = "0.080000000000000000"' genesis.json > temp.json && mv temp.json genesis.json
                                                                                                                                                                
kava start