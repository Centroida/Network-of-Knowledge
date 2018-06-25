#!/bin/bash
#
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

set -e

source $(dirname "$0")/env.sh

function main {

   done=false

   # Wait for setup to complete and then wait another 10 seconds for the orderer and peers to start
   awaitSetup
   sleep 10

   trap finish EXIT

   mkdir -p $LOGPATH
   logr "The docker 'run' container has started"

   # Set ORDERER_PORT_ARGS to the args needed to communicate with the 1st orderer
   IFS=', ' read -r -a OORGS <<< "$ORDERER_ORGS"
   initOrdererVars ${OORGS[0]} 1
  #  export ORDERER_PORT_ARGS="-o $ORDERER_HOST:7050 --tls --cafile $CA_CHAINFILE --clientauth"
     export ORDERER_PORT_ARGS="-o $ORDERER_HOST:7050"


   # Convert PEER_ORGS to an array named PORGS
   IFS=', ' read -r -a PORGS <<< "$PEER_ORGS"

   # Create the channel
   createChannel

   # All peers join the channel
   for ORG in $PEER_ORGS; do
      local COUNT=1
      while [[ "$COUNT" -le $NUM_PEERS ]]; do
         initPeerVars $ORG $COUNT
         joinChannel
         COUNT=$((COUNT+1))
      done
   done

   # Update the anchor peers
   for ORG in $PEER_ORGS; do
      initPeerVars $ORG 1
      switchToAdminIdentity
      logr "Updating anchor peers for $PEER_HOST ..."
      peer channel update -c $CHANNEL_NAME -f $ANCHOR_TX_FILE $ORDERER_CONN_ARGS
   done

   # Install chaincode on the both peers in the org
      initPeerVars $ORG 1
      installChaincode

      initPeerVars $ORG 2
      installChaincode

   # Instantiate chaincode on the 1st peer of the 2nd org
   makePolicy
   initPeerVars $ORG 1
   switchToAdminIdentity
   logr "Instantiating chaincode on $PEER_HOST ..."
   peer chaincode instantiate -C $CHANNEL_NAME -n $CHAINCODE_NAME -v $CHAINCODE_VERSION -P "$POLICY" $ORDERER_CONN_ARGS -c $CHAINCODE_INIT

   logr "Congratulations! The tests ran successfully."

   done=true
}

# Enroll as a peer admin and create the channel
function createChannel {
   initPeerVars $ORG 1
   switchToAdminIdentity
   logr "Creating channel '$CHANNEL_NAME' on $ORDERER_HOST ..."
   peer channel create --logging-level=DEBUG -c $CHANNEL_NAME -f $CHANNEL_TX_FILE $ORDERER_CONN_ARGS
   logr "Channel created."
}

# Enroll as a fabric admin and join the channel
function joinChannel {
   switchToAdminIdentity
   set +e
   local COUNT=1
   MAX_RETRY=10
   while true; do
      logr "Peer $PEER_HOST is attempting to join channel '$CHANNEL_NAME' (attempt #${COUNT}) ..."
      peer channel join -b $CHANNEL_NAME.block
      if [ $? -eq 0 ]; then
         set -e
         logr "Peer $PEER_HOST successfully joined channel '$CHANNEL_NAME'"
         return
      fi
      if [ $COUNT -gt $MAX_RETRY ]; then
         fatalr "Peer $PEER_HOST failed to join channel '$CHANNEL_NAME' in $MAX_RETRY retries"
      fi
      COUNT=$((COUNT+1))
      sleep 1
   done
}

function makePolicy  {
   POLICY="OR("
   local COUNT=0
   for ORG in $PEER_ORGS; do
      if [ $COUNT -ne 0 ]; then
         POLICY="${POLICY},"
      fi
      initOrgVars $ORG
      POLICY="${POLICY}'${ORG_MSP_ID}.member'"
      COUNT=$((COUNT+1))
   done
   POLICY="${POLICY})"
   log "policy: $POLICY"
}

function installChaincode {
   switchToAdminIdentity
   logr "Installing chaincode on $PEER_HOST ..."
   peer chaincode install -n $CHAINCODE_NAME -v $CHAINCODE_VERSION -p $CHAINCODE_LOCATION
}

function fetchConfigBlock {
   logr "Fetching the configuration block of the channel '$CHANNEL_NAME'"
   peer channel fetch config $CONFIG_BLOCK_FILE -c $CHANNEL_NAME $ORDERER_CONN_ARGS
}

function updateConfigBlock {
   logr "Updating the configuration block of the channel '$CHANNEL_NAME'"
   peer channel update -f $CONFIG_UPDATE_ENVELOPE_FILE -c $CHANNEL_NAME $ORDERER_CONN_ARGS
}

function finish {
   if [ "$done" = true ]; then
      logr "See $RUN_LOGFILE for more details"
      touch /$RUN_SUCCESS_FILE
   else
      logr "Tests did not complete successfully; see $RUN_LOGFILE for more details"
      touch /$RUN_FAIL_FILE
   fi
}

function logr {
   log $*
   log $* >> $RUN_SUMPATH
}

function fatalr {
   logr "FATAL: $*"
   exit 1
}

main
