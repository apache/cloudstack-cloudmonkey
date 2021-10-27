# Licensed to the Apache Software Foundation (ASF) under one
# or more contributor license agreements.  See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership.  The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License.  You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied.  See the License for the
# specific language governing permissions and limitations
# under the License.

FROM debian:sid as builder

MAINTAINER "Apache CloudStack" <dev@cloudstack.apache.org>
LABEL Description="Apache CloudStack CloudMonkey; Go based CloudStack command line interface"
LABEL Vendor="Apache.org"
LABEL License=ApacheV2
LABEL Version=6.3.0

WORKDIR /work/
RUN apt -y update && apt -y install git golang-go build-essential && \
    git clone https://github.com/apache/cloudstack-cloudmonkey.git && \
    go version && \
    cd cloudstack-cloudmonkey && \
    make all && \
    pwd && \
    ls -alh ./bin/cmk

FROM debian:stable
COPY --from=builder /work/cloudstack-cloudmonkey/bin/cmk /usr/bin/
RUN apt-get -y update && uname -a && mkdir -p /root/.cmk/ &&\
    cmk version && cmk help && ls -alh /root/
