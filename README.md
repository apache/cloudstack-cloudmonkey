# cloudmonkey-interoute (using Cloudmonkey version 5.3.2)

This is a fork of the [Cloudmonkey command line interface](https://github.com/apache/cloudstack-cloudmonkey), version 5.3.2, modified for use with the [Interoute Virtual Data Centre](https://cloudstore.interoute.com/what_is_vdc).

The modification required is to handle API access to the different VDC regions (Europe, North America (USA), Asia). This is implemented using a new configuration variable `region`.

The default value is 'europe'. Note that the master version of Cloudmonkey will access the default region without modification.

See the original repo here: [cloudstack-cloudmonkey](https://github.com/apache/cloudstack-cloudmonkey)

## Modifications

The files `cloudmonkey.py`, `requester.py` and `config.py` have been modified. To access the Asia or USA region, add a line to each required profile setting in the Cloudmonkey `config` file:

    region = asia

or 

    region = usa

(the region name is not case sensitive). You can also change the region interactively while Cloudmonkey is running:

    set region usa


## How to install this modified version

You can make a fresh install of this modified version of Cloudmonkey using the `pip` command:

    sudo pip install git+https://github.com/Interoute/cloudmonkey-interoute.git

Or to upgrade an existing installation:

    sudo pip install --upgrade git+https://github.com/Interoute/cloudmonkey-interoute.git

This version of Cloudmonkey may not work with other cloud computing providers that are also compatible with Cloudmonkey. So be careful if you use Cloudmonkey with several providers.

(The reason is that this version adds a new parameter input 'region=...' to every API call. API servers generally ignore any parameters that are not recognised, however if the server is programmed to do validity checking then Cloudmonkey may not work.)

An alternative is to use [Python's virtualenv](http://www.pythoncentral.io/how-to-install-virtualenv-python) tool so that you can operate multiple versions of Cloudmonkey at the same time.

