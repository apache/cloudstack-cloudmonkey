## CloudMonkey

`cloudmonkey` :cloud::monkey_face: is a command line interface for
[Apache CloudStack](http://cloudstack.apache.org).
CloudMonkey can be use both as an interactive shell and as a command line tool
which simplifies Apache CloudStack configuration and management. It can be used
with Apache CloudStack 4.0-incubating and above.

![version badge](https://badge.fury.io/py/cloudmonkey.png) ![download badge](http://img.shields.io/pypi/dm/cloudmonkey.png)


### For users

Install:

    $ pip install cloudmonkey

Upgrade:

    $ pip install --upgrade cloudmonkey

Install/Upgrade latest using git repository:

    $ pip install --upgrade git+https://git-wip-us.apache.org/repos/asf/cloudstack-cloudmonkey.git

Install/upgrade using the Github git mirror:

    $ pip install --upgrade git+https://github.com/apache/cloudstack-cloudmonkey.git

Please see the [CloudMonkey Wiki](https://cwiki.apache.org/confluence/display/CLOUDSTACK/CloudStack+cloudmonkey+CLI) for usage.


### Using Docker image

The default configuration provided connect to CloudStack managemenent server as container:

Enter the CLI:

    $ docker run -ti --rm --link cloudstack:8080 cloudstack/cloudmonkey

To execute single api command:

    $ docker run -ti --rm --link cloudstack:8080 cloudstack/cloudmonkey list accounts

Use your own CloudMonkey configuration file:

    $ docker run -ti --rm -v `pwd`/.cloudmonkey:/cloudmonkey cloudstack/cloudmonkey


### Build

All:

    Cleans and then builds with precache
    $ make all

Build:

    $ make build

Build Precache:

    $ make buildcache

Build with Precache:

    $ make buildwithcache

Check changes, code styles:

    $ make check

Clean:

    $ make clean

Install:

    $ make install


### Mailing lists

[Development Mailing List](mailto:dev-subscribe@cloudstack.apache.org)

[Users Mailing List](mailto:users-subscribe@cloudstack.apache.org)

[Commits Mailing List](mailto:commits-subscribe@cloudstack.apache.org)

[Issues Mailing List](mailto:issues-subscribe@cloudstack.apache.org)

[Marketing Mailing List](mailto:marketing-subscribe@cloudstack.apache.org)


### Contributing

Discuss features development on the [Development Mailing List](mailto:dev-subscribe@cloudstack.apache.org).
Report issues on the `User` mailing list and open issue on [JIRA](http://issues.apache.org/jira/browse/CLOUDSTACK).

1. Fork the repository on Github
2. Create a named feature branch (like `add_component_x`)
3. Write your change
4. Write tests for your change (if applicable)
5. Run the tests, ensuring they all pass
6. Submit a Pull Request using Github


### License

Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.