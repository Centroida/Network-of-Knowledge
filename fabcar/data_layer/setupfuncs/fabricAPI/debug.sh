#!/bin/bash
rm -rf hfc-key-store
npm install
nodemon --inspect=0.0.0.0:9229 ./bin/www