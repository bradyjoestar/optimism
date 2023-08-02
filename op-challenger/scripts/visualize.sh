#!/bin/bash

set -euo pipefail

if [ $# -eq 0 ]
  then
    echo "Missing Fault Dispute Game address argument"
fi

DISPUTE_GAME_PROXY=$(jq .DisputeGameFactoryProxy .devnet/addresses.json)
DISPUTE_GAME_PROXY=$(echo $DISPUTE_GAME_PROXY | tr -d '"')

echo "----------------------------------------------------------------"
echo " Dispute Game Factory at $DISPUTE_GAME_PROXY"
echo "----------------------------------------------------------------"

FAULT_GAME_ADDRESS=$1

dir=$(cd $(dirname "${BASH_SOURCE[0]}") && pwd)
cd "$dir"
cd ../packages/contracts-bedrock

forge script scripts/FaultDisputeGameViz.s.sol --sig "remote(address)" $FAULT_GAME_ADDRESS --fork-url http://localhost:8545
mv dispute_game.svg "$dir"
