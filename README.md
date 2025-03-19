## CloudMonkey [![Build Status](https://travis-ci.com/apache/cloudstack-cloudmonkey.svg?branch=main)](https://travis-ci.com/apache/cloudstack-cloudmonkey)[![](https://images.microbadger.com/badges/version/apache/cloudstack-cloudmonkey.svg)](https://hub.docker.com/r/apache/cloudstack-cloudmonkey)[![cloudmonkey](https://snapcraft.io/cloudmonkey/badge.svg)](https://snapcraft.io/cloudmonkey)

`cloudmonkey` :cloud::monkey_face: is a command line interface (CLI) for
[Apache CloudStack](http://cloudstack.apache.org).
It can be used both as an interactive shell and as a command-line tool, simplifying Apache CloudStack configuration and management.

The modern CloudMonkey is a rewritten and simplified port in Go, compatible
with Apache CloudStack 4.9 and above. The legacy cloudmonkey written in Python
can be used with Apache CloudStack 4.0-incubating and above.

For documentation, kindly see the [wiki](https://github.com/apache/cloudstack-cloudmonkey/wiki).

### Development

To develop CloudMonkey, you need Go 1.11 or later and a unix-like
environment. You can use the following targets for building:

    $ make help

      all             Build program binary
      test-bench      Run benchmarks
      test-short      Run only short tests
      test-verbose    Run tests in verbose mode with coverage reporting
      test-race       Run tests with race detector
      check tests     Run tests
      test-xml        Run tests with xUnit output
      test-coverage   Run coverage tests
      debug           Runs a debuggable binary using dlv
      lint            Run golint
      fmt             Run gofmt on all source files

Build and run:

    $ make run

Build and run manually:

    $ make all
    $ ./bin/cmk

To build for all distros and platforms, run:

    $ make dist

### Community

You may join the relevant mailing list(s) for cloudmonkey related discussion:

[Development Mailing List](mailto:dev-subscribe@cloudstack.apache.org)

[Users Mailing List](mailto:users-subscribe@cloudstack.apache.org)

### Contribution

Discuss issue(s) and feature(s) on CloudStack [development mailing list](mailto:dev-subscribe@cloudstack.apache.org).
Report issue(s) on the `user` mailing list and/or open a Github [issue](https://github.com/apache/cloudstack-cloudmonkey/issues).

1. Fork the repository on Github
2. Create a named feature branch (like `add_component_x`)
3. Commit your change
4. Write tests for your change if applicable
5. Run the tests, ensuring they all pass
6. Submit a [Pull Request](https://github.com/apache/cloudstack-cloudmonkey/pull/new/main) using Github

### History

The original `cloudmonkey` was written in Python and contributed to Apache
CloudStack project by [Rohit Yadav](http://rohityadav.cloud) on 31 Oct 2012
under the Apache License 2.0.

Starting version 6.0.0, the modern cloudmonkey `cmk` is a fast and simplified
Go-port of the original tool with some backward incompatibilities and reduced
feature set. It ships as a standalone 64-bit [executable binary for several
platforms such as Linux, Mac and Windows](https://github.com/apache/cloudstack-cloudmonkey/releases).

**NOTE:**

If cloudmonkey is being upgraded from a version lower than v6.0.0, it must be noted
that the cloudmonkey configuration path is changed from `~/.cloudmonkey/config` to 
`~/.cmk/config` and a default `localcloud` profile is created. One must first set up basic configurations such as apikey/secretkey/username/password/url for the required profile(s) as required

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
