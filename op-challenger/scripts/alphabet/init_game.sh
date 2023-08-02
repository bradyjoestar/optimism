#!/bin/bash
set -euo pipefail

DIR=$(cd $(dirname "${BASH_SOURCE[0]}") && pwd)
cd "$DIR"

cd ../../

make

cd ..

echo "----------------------------------------------------------------"
echo " - Cleaning up the devnet"
echo "----------------------------------------------------------------"

make devnet-clean

echo "----------------------------------------------------------------"
echo " - Starting a new devnet"
echo "----------------------------------------------------------------"

make cannon-prestate
make devnet-up

DEVNET_SPONSOR="ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
DISPUTE_GAME_PROXY=$(jq .DisputeGameFactoryProxy .devnet/addresses.json)
DISPUTE_GAME_PROXY=$(echo $DISPUTE_GAME_PROXY | tr -d '"')

echo "----------------------------------------------------------------"
echo " Dispute Game Factory at $DISPUTE_GAME_PROXY"
echo "----------------------------------------------------------------"

L2_OUTPUT_ORACLE_PROXY=$(jq .L2OutputOracleProxy .devnet/addresses.json)
L2_OUTPUT_ORACLE_PROXY=$(echo $L2_OUTPUT_ORACLE_PROXY | tr -d '"')

echo "----------------------------------------------------------------"
echo " L2 Output Oracle Proxy at $L2_OUTPUT_ORACLE_PROXY"
echo "----------------------------------------------------------------"

BLOCK_ORACLE_PROXY=$(jq .BlockOracle .devnet/addresses.json)
BLOCK_ORACLE_PROXY=$(echo $BLOCK_ORACLE_PROXY | tr -d '"')

echo "----------------------------------------------------------------"
echo " Block Oracle Proxy at $BLOCK_ORACLE_PROXY"
echo "----------------------------------------------------------------"

CHARLIE_ADDRESS="0xF45B7537828CB2fffBC69996B054c2Aaf36DC778"
CHARLIE_KEY="74feb147d72bfae943e6b4e483410933d9e447d5dc47d52432dcc2c1454dabb7"

MALLORY_ADDRESS="0x4641c704a6c743f73ee1f36C7568Fbf4b80681e4"
MALLORY_KEY="28d7045146193f5f4eeb151c4843544b1b0d30a7ac1680c845a416fac65a7715"

echo "----------------------------------------------------------------"
echo " - Fetching balance of the sponsor"
echo " - Balance: $(cast balance 0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266)"
echo "----------------------------------------------------------------"

echo "Funding Charlie"
cast send $CHARLIE_ADDRESS --value 5ether --private-key $DEVNET_SPONSOR

echo "Funding Mallory"
cast send $MALLORY_ADDRESS --value 5ether --private-key $DEVNET_SPONSOR

# Loop and wait until there are at least 2 outputs in the l2 output oracle
while [ $(cast block-number) -lt 18 ]
do
  echo "[BLOCK: $(cast block-number)] Waiting for output proposals..."
  sleep 2
done

BLOCK_NUMBER=12
echo "At block number: $BLOCK_NUMBER"
echo "Verifying that there are at least 2 outputs in the L2 Output Oracle..."
INDEX=$(cast call $L2_OUTPUT_ORACLE_PROXY "getL2OutputIndexAfter(uint256)" $BLOCK_NUMBER)
INDEX=$(cast to-dec $INDEX)
echo "Index: $INDEX"

# We will use the l2 block number of 1 for the dispute game.
# We need to check that the block oracle contains the corresponding l1 block number.
echo "Checkpointing the block oracle..."
L1_BLOCK_NUMBER=$(cast send --private-key $DEVNET_SPONSOR $BLOCK_ORACLE_PROXY "checkpoint()" --json | jq .blockNumber)
L1_BLOCK_NUMBER=$(echo $L1_BLOCK_NUMBER | tr -d '"')
echo "L1 Block Number: $L1_BLOCK_NUMBER"
L1_BLOCK_NUMBER=$(cast to-dec $L1_BLOCK_NUMBER)
echo "L1 Block Number (dec): $L1_BLOCK_NUMBER"
((EXTRA_DATA_ARG=L1_BLOCK_NUMBER-1))

echo "Getting the l2 output at index 0"
((PRIOR_INDEX=INDEX-1))
cast call $L2_OUTPUT_ORACLE_PROXY "getL2Output(uint256)" $PRIOR_INDEX

echo "Getting the l2 output at index $INDEX"
cast call $L2_OUTPUT_ORACLE_PROXY "getL2Output(uint256)" $INDEX


# (Alphabet) Fault game type = 0
GAME_TYPE=0

# Root claim commits to the entire trace.
# Alphabet game claim construction: keccak256(abi.encode(trace_index, trace[trace_index]))
ROOT_CLAIM=$(cast keccak $(cast abi-encode "f(uint256,uint256)" 15 122))

# fault dispute game extra data:
# abi.encode(uint256(l2_block_number), uint256(l1_block_number))
EXTRA_DATA=$(cast abi-encode "f(uint256,uint256)" $BLOCK_NUMBER $EXTRA_DATA_ARG)

echo "Initializing the game"
FAULT_GAME_ADDRESS=$(cast call --private-key $MALLORY_KEY $DISPUTE_GAME_PROXY "create(uint8,bytes32,bytes)" $GAME_TYPE $ROOT_CLAIM $EXTRA_DATA)
echo "Creating game at address $FAULT_GAME_ADDRESS"
cast send --private-key $MALLORY_KEY $DISPUTE_GAME_PROXY "create(uint8,bytes32,bytes)" $GAME_TYPE $ROOT_CLAIM $EXTRA_DATA

FORMATTED_ADDRESS=$(cast parse-bytes32-address $FAULT_GAME_ADDRESS)
echo "Formatted Address: $FORMATTED_ADDRESS"

cd op-challenger
echo $FORMATTED_ADDRESS > .fault-game-address
