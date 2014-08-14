# cloudmonkey-interoute

This is a fork of the Cloudmonkey command line interface, modified for use with the [Interoute Virtual Data Centre](http://cloudstore.interoute.com/main/WhatInterouteVDC).

The modification required is to handle API access to the different VDC regions (Europe, USA, Asia). This is implemented using a new configuration variable 'region'.

The default value is 'europe'. Note that the master version of Cloudmonkey can access the default region without modification.

See the original repo here: [cloudstack-cloudmonkey](https://github.com/apache/cloudstack-cloudmonkey)

## Modifications

The files cloudmonkey.py, requester.py and config.py have been modified. To access the Asia or USA region, add a line to the Cloudmonkey config file in the [server] section:

region = asia

or 

region = usa

(the region name is not case sensitive).

## How to install this modified version

(to be added)
