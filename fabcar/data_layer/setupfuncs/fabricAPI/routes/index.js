'use strict';

require('dotenv').config()
var express = require('express');
var queryMethods = require('../queryMethods.js');
var url = require('url');
const UserManager = require('../usermanager');
var url = require('url');
var bodyParser = require('body-parser');
var usermanager = new UserManager();
var router = express.Router();
var user = null;
var Fabric_Client = require('fabric-client');

usermanager.setUpInteractors()
  .then(() => { return usermanager.Enrolluser(usermanager.userSecret, process.env.REGISTER_USER_USERNAME) })
  .then(() => { return queryMethods.setup(process.env.USER_NAME) }).then((usr) => { user = usr; });

router.post('/api/user/register', (req, res) => {
  var username = req.body.username
  usermanager.register_New_User(username)
    .then(() => { res.send(usermanager.userSecret) })
    .catch((error) => {
      console.log("error : " + error);
    })
});

router.post('/api/user/enroll', (req, res) => {
  var secret = req.body.secret;
  var username = req.body.username

  return usermanager.Enrolluser(secret, username)
    .then(() => { queryMethods.setup(username) });
});

// Query chaincode using chaincode id and chaincode function.
router.post('/api/chaincode/query/:ccid/:ccfn', (req, res) => {
  var chaincodeId = req.params.ccid;
  var chaincodeFunction = req.params.ccfn;
  var queryArgsArray = req.body.Args;

  queryMethods.queryChaincode(user, chaincodeId, chaincodeFunction, queryArgsArray)
    .then((responsePayloads) => queryMethods.parseResponsePayloads(responsePayloads))
    .then((parsedResponse) => res.send(JSON.parse(parsedResponse)))
    .catch((error) => {
      console.log(error, 'error');
    })
});

// Invoke chaincode using chaincode id and chaincode function.
router.post('/api/chaincode/invoke/:ccid/:ccfn', (req, res) => {
  var chaincodeId = req.params.ccid;
  var chaincodeFunction = req.params.ccfn;
  var queryArgsArray = req.body.Args;

  queryMethods.invokeChaincode(user, chaincodeId, chaincodeFunction, queryArgsArray)
    .then((proposal) => {
      return queryMethods.getPeerResponses(proposal);
    })
    .then((invocationResult) => res.send(invocationResult))
});

module.exports = router;    