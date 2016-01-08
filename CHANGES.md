Apache CloudStack CloudMonkey Changelog
---------------------------------------

Version 5.3.3
=============
This release includes
- Support for shell history

Version 5.3.2
=============
This release includes
- Spinner printing improvements
- Monkey patching SSL requests
- Encode publickey, privatekey and certificates in APIs
- Configurable signature version
- Fix tabular output mode
- Better response checking in the API response
- A new XML display output mode
- A new CSV display output mode

Version 5.3.1
=============
This release includes
- Users can specify domain when using username/password auth per server profile
- Autocompletion of args works when cursor is not at the end of the line
- CLOUDSTACK-7935: keep colons in the request to ACS
- Account parameters are sometimes UUIDs and sometimes string, CloudMonkey
  now automatically autocompletes for both UUID and string account args
- Pass verifysslcert option while user logs in using username/password
- Importing readline no longer outputs escape characters
- CloudMonkey will not output extra empty lines in stdout output
- Filtered result output is uniform across output display formats
- Async blocked API now show a spinning cursor instead of print dots
- When finding missing API args, it does case insensitive search
- New command line arg: -p or --profile (load server profile)
- New command line arg: -b, --block-async (block and poll result on async API calls)
- New command line arg: -n, --noblock-async (do not block on async API calls)

Version 5.3.0
=============
This release includes
- CloudMonkey becomes unicode friendly
- Autocompletion for filters, precache changes
- Autocompletion for config 'set' options and boolean api args
- Current server profile displayed on prompt
- Server profile related bugfixes, blank profile names are not allowed
- Filtering support for default display output
- Filtering by single key only outputs results without key names
- Non-interactive commands from command line are outputted without colors
- Parameter completion uses list api heuristics and related APIs as fallback
- Parameter completion options are cached to speed up rendering of options
- CloudMonkey returns non-zero exit code when run on shell and a error is return
  from managment server, the error message is written to stderr
- Adds new config parameter 'verifysslcert' to enable/disable SSL cert checking
- New command line arg: -d for display (json, table or default)

Make sure you backup your config before you upgrade cloudmonkey from previous releases.
With this release `cloudmonkey` will automatically fix your config file, add missing
configuration parameters and save it as the upgraded versions starts for the first time.

Version 5.2.0
=============
This release includes
 - In the config [server] section is deprecated now
 - For missing keys, cloudmonkey will set default values
 - Network requests, json decoding and shell related bugfixes
 - Based on platform, it will install either pyreadline (Windows) or readline (OSX and Linux)
 - Config options `protocol`, `host`, `port`, `path` are deprecated now
 - Backward compatibilty exists for above options but we use `url` for the mgmt server URL
 - Introduces server profiles so users can use cloudmonkey with different hosts and management server configs
 - A default profile under the section [local] is added with default values
 - Everytime `set` is called, cloudmonkey will write the config and reload config file

Make sure you backup your config before you upgrade cloudmonkey from previous releases.
With this release `cloudmonkey` will automatically fix your config file, add missing
configuration parameters and save it as the upgraded versions starts for the first time.

Version 5.1.0
=============
This release includes
 - support for using username and password instead of / in addition to api key and secret key
 - Usage of signature version 3 for the api signing process. This reduces the chance of API replay attacks
 - cleanup based on reporting from PEP8 and Flake8

If you upgrade from 5.0, then cloudmonkey will ask you to update your config file (~/.cloudmonkey/config)
Under the [user], you can add
username =
password =
Under the [server], you can add
expires = 600

Version 5.0.0
=============
This is the first release of CloudMonkey independent from the Apache CloudStack core orchestration engine. The release
includes a precache of Apache CloudStack 4.2.0 API calls, and should be backward compatible with prior 3.x and 4.x
CloudStack installations (with the obvious caveat that previous versions will have a subset of the latest API commands /
parameters).
