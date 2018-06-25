'use strict';

require('dotenv').config()

var Fabric_Client = require('fabric-client');
var Fabric_CA_Client = require('fabric-ca-client');

var path = require('path');
var util = require('util');
var os = require('os');

class UserManager {

    constructor() {
        this.fabric_client = new Fabric_Client();
        this.fabric_ca_client = null;
        this.admin_user = null;
        this.new_admin = null;
        this.member_user = null;
        this.store_path = path.join(__dirname, 'hfc-key-store');
        this.IsInited = false;
        this.userSecret = null;
    }

    async init() {
        console.log("INFO: Initializing connection.");
        await Fabric_Client.newDefaultKeyValueStore({ path: this.store_path }).then((state_store) => {
            this.fabric_client.setStateStore(state_store);
            var crypto_suite = Fabric_Client.newCryptoSuite();
            var crypto_store = Fabric_Client.newCryptoKeyStore({ path: this.store_path });
            crypto_suite.setCryptoKeyStore(crypto_store);
            this.fabric_client.setCryptoSuite(crypto_suite);
            var tlsOptions = {
                trustedRoots: [],
                verify: false
            };
            this.fabric_ca_client = new Fabric_CA_Client(process.env.FABRIC_CA_CLIENT_PORT, tlsOptions, process.env.INTERMDIATE_CA_NOKORG, crypto_suite);
            this.IsInited = true;
        });
    }
    
    async setUpInteractors() {
        if (!this.IsInited) {
            await this.init();
        }
        var found = false;
        console.log("INFO: Attempting to get admin from persistence.");

        return this.fabric_client.getUserContext('admin', true).then((admin) => {
            if (admin && admin.isEnrolled()) {
                console.log("INFO: Found admin from store, enrolling");
                found = true;
                this.admin_user = admin;
                return this.registerUser(process.env.REGISTER_USER_USERNAME);
            }
            else {

                console.log("INFO: Didn't find the admin, attempting to enroll him/her.");
                return this.enrollAdmin().then(() => this.registerUser(process.env.REGISTER_USER_USERNAME));
            }
        })
    }

    async enrollAdmin() {
        return this.fabric_ca_client.enroll({
            enrollmentID: 'admin',
            enrollmentSecret: 'admin'
        }).then((enrollment) => {
            return this.fabric_client.createUser(
                {
                    username: 'admin', mspid: process.env.ORGANIZATION_MSP,
                    cryptoContent: { privateKeyPEM: enrollment.key.toBytes(), signedCertPEM: enrollment.certificate }
                });
        }).then((user) => {
            this.admin_user = user;
            console.log("INFO: Registered and enrolled admin:" + this.admin_user);
            return this.fabric_client.setUserContext(this.admin_user);
        })
    }

    async registerUser(enrollmentId) {
        if (!this.IsInited) {
            await this.init();
        }
        return this.fabric_client.getUserContext("admin", true)
            .then((admin) => {
                return this.fabric_ca_client.register({
                    enrollmentID: enrollmentId, affiliation: 'org1.department1', role: 'client'
                }, admin);
            })
            .then((secret) => {
                this.userSecret = secret;
                console.log("INFO: User secret captured " + this.userSecret)
            });
    }

    async Enrolluser(secret, enrollmentId) {
        console.log('INFO: Successfully registered ' + enrollmentId + ' - secret:' + secret);
        return this.fabric_ca_client.enroll(
            { enrollmentID: enrollmentId, enrollmentSecret: secret })
            .then((enrollment) => {
                console.log('INFO: Successfully enrolled member ' + enrollmentId)
                return this.fabric_client
                    .createUser(
                        {
                            username: enrollmentId,
                            mspid: process.env.ORGANIZATION_MSP,
                            cryptoContent: { privateKeyPEM: enrollment.key.toBytes(), signedCertPEM: enrollment.certificate }
                        });
            })
            .then((user) => this.fabric_client.setUserContext(user))//set the user context 
    }

    register_New_User(username) {
        return this.registerUser(username)
    }
};

module.exports = UserManager;