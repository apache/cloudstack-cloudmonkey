# MODIFICATION INSTRUCTIONS FOR cloudmonkey-interoute

This is a fork of the [Cloudmonkey command line interface](https://github.com/apache/cloudstack-cloudmonkey), modified for use with the [Interoute Virtual Data Centre](http://cloudstore.interoute.com/main/WhatInterouteVDC).

The modification required is to handle API access to the different VDC regions (Europe, North America (USA), Asia). This is implemented using a new configuration variable `region`.

The files `cloudmonkey.py`, `requester.py` and `config.py` have been modified. 

## How to apply the modification to a new version of Cloudmonkey

Create a local clone of the `cloudmonkey-interoute` master branch.

Add the Cloudmonkey master repo as a 'remote upstream' repo:

    git remote add upstream https://github.com/apache/cloudstack-cloudmonkey

Create a new branch (with appropriate version number in place of "531") in which to merge the modifications with the latest 'upstream' Cloudmonkey, and checkout that branch:

    git branch interoute_mod_531_regions

    git checkout interoute_mod_531_regions

Do the merge:

    git pull upstream master

This will create some conflicts which will require manual edits, then 'git commit' these edits in the branch. Also edit `README.md` to change the Cloudmonkey version number.
 
Push the new branch to github.com:

    git push origin interoute_mod_531_regions

At this point, download the modified branch into a 'virtualenv' and test it. If testing is OK, go ahead and merge the new branch into the `cloudmonkey-interoute` master branch:

    git merge interoute_mod_531_regions

And push the changed master to github.com:

    git push

I suggest to keep the branch `interoute_mod_531_regions` as it can be used anytime later to run the modified Cloudmonkey at the point of that version (v5.3.1 in this case).
    
    