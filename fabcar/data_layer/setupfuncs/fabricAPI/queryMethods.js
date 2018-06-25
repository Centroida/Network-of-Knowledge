require('dotenv').config()

var Fabric_Client = require('fabric-client');

var path = require('path');
var util = require('util');
var os = require('os');

var fabric_client = new Fabric_Client();

var txId = null;

var store_path = path.join(__dirname, 'hfc-key-store');
console.log(`DEBUG: Successfully retrieved the store_path : ${store_path}`);
if (!process.env.CHANNEL_NAME) {
    console.log("ERROR: CHANNEL_NAME is not set. Deliberately crashing.");
}
var channel = fabric_client.newChannel(process.env.CHANNEL_NAME);
var peer = fabric_client.newPeer(process.env.PEER_PORT_CHANNEL);
channel.addPeer(peer);
var order = fabric_client.newOrderer(process.env.ORDER_PORT);
channel.addOrderer(order);

exports.setup = function (currentuser) {
    return Fabric_Client.newDefaultKeyValueStore({
        path: store_path
    }).then((state_store) => {
        // assign the store to the fabric client
        fabric_client.setStateStore(state_store);
        var crypto_suite = Fabric_Client.newCryptoSuite();
        // use the same location for the state store (where the users' certificate are kept)
        // and the crypto store (where the users' keys are kept)
        var crypto_store = Fabric_Client.newCryptoKeyStore({ path: store_path });
        crypto_suite.setCryptoKeyStore(crypto_store);
        fabric_client.setCryptoSuite(crypto_suite);
        var tlsOptions = {
            trustedRoots: [],
            verify: false
        };
        return fabric_client.getUserContext(currentuser, true)
    })
}

exports.queryChaincode = function (user, chaincodeId, chaincodeFunction, argsArray) {
    if (user == null) {
        throw new Error(`ERROR: User is null.`);
    }

    if (user && user.isEnrolled()) {
        console.log("INFO: User successfully pulled from db.");
    } else {
        throw new Error(`ERROR: Failed to get user ${user.username}`);
    }

    var request = {
        chaincodeId: chaincodeId,
        fcn: chaincodeFunction,
        args: argsArray
    };
    
    return channel.queryByChaincode(request);
}

exports.invokeChaincode = function (user, chaincodeId, chaincodeFunction, argsArray) {
    if (user && user.isEnrolled()) {
        console.log(`INFO: Successfully loaded ${user.username} from persistence`);
    }
    else {
        throw new Error(`ERROR: Failed to get ${user.username}`);
    }

    txId = fabric_client.newTransactionID();
    console.log("INFO: Assigning transaction id : " + txId._transactionId)

    const request = {
        chaincodeId: chaincodeId,
        fcn: chaincodeFunction,
        args: argsArray,
        chainId: process.env.CHANNEL_NAME,
        txId: txId
    };

    return channel.sendTransactionProposal(request);
}

exports.parseResponsePayloads = function (prms_res) {
    if (prms_res && prms_res.length == 1) {
        if (prms_res[0] instanceof Error) {
            console.error("ERROR: Error from query = ", prms_res[0].toString());
        } else {

            return prms_res[0].toString();
        }
    }
}

exports.sendTransactionProposal = function (user, arr, funcname) {

    if (user && user.isEnrolled()) {
        console.log(`INFO: Successfully loaded ${user.username} from persistence`);
    }
    else {
        throw new Error(`ERROR: Failed to get ${user.username}`);
    }

    txId = fabric_client.newTransactionID();
    console.log("INFO: Assigning transaction id : " + txId._transactionId)

    const request = {
        chaincodeId: 'fabcar',
        fcn: funcname,
        args: arr,
        chainId: 'mychannel',
        txId: txId
    };

    return channel.sendTransactionProposal(request);
}

exports.getPeerResponses = function (results) {
    var peerResponses = results[0];
    var proposal = results[1];

    var isProposalOK = false;
    if (peerResponses && peerResponses[0].response && peerResponses[0].response.status === 200) {
        isProposalOK = true;
        console.log('INFO: Transaction proposal was good');
    } else {
        console.error('INFO: Transaction proposal was bad');
    }

    if (isProposalOK) {
        console.log(
            util.format(
                'INFO: Successfully sent Proposal and received ProposalResponse: Status - %s, message - "%s"',
                peerResponses[0].response.status, peerResponses[0].response.message
            )
        );

        const request = {
            proposalResponses: peerResponses,
            proposal: proposal,
        };

        var transactionId = txId.getTransactionID();
        var promises = [];
        var transaction_status = channel.sendTransaction(request);
        promises.push(transaction_status);
        var event_hub = fabric_client.newEventHub();
        event_hub.setPeerAddr(process.env.EVENT_HUB_PEER_PORT);

        let txPromise = new Promise((resolve, reject) => {
            let handle = setTimeout(() => {
                event_hub.disconnect();
                resolve({ event_status: 'TIMEOUT' });
            }, 30000);

            event_hub.connect();
            event_hub.registerTxEvent(transactionId, (tx, code) => {
                clearTimeout(handle);
                event_hub.unregisterTxEvent(transactionId);
                event_hub.disconnect();
                var return_status = { event_status: code, txId: transactionId };

                if (code !== 'VALID') {
                    console.error('ERROR: The transaction was invalid, code = ' + code);
                    resolve(return_status); // we could use reject(new Error('Problem with the tranaction, event status ::'+code));
                } else {
                    console.log('INFO: The transaction has been committed on peer ' + event_hub._ep._endpoint.addr);
                    resolve(return_status);
                }
            }, (err) => {
                console.log('ERROR: There was a problem with the eventhub :' + err);
            });
        });

        promises.push(txPromise);
        return Promise.all(promises);
    } else {
        console.error('ERROR: Failed to send Proposal or receive valid response. Response null or status is not 200. exiting...');
        throw new Error('Failed to send Proposal or receive valid response. Response null or status is not 200. exiting...');
    }
}

exports.confirmation = function (results) {
    console.log('INFO: Send transaction promise and event listener promise have completed.');

    if (results && results[0] && results[0].status === 'SUCCESS') {
        console.log('INFO: Successfully sent transaction to the orderer.');
    } else {
        console.error('ERROR: Failed to order the transaction. Error code: ' + response.status);
    }

    if (results && results[1] && results[1].event_status === 'VALID') {

        return 'Successfully committed the change to the ledger by the peer';
        console.log('INFO: Successfully committed the change to the ledger by the peer');
    } else {
        console.log('ERROR: Transaction failed to be committed to the ledger due to ::' + results[1].event_status);
    }
}