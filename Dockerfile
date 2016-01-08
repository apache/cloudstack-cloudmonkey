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
#
#
FROM python:2

MAINTAINER "Apache CloudStack" <dev@cloudstack.apache.org>
LABEL Description="Apache CloudStack CloudMonkey; Python based CloudStack command line interface"
LABEL Vendor="Apache.org"
LABEL License=ApacheV2
LABEL Version=5.3.3

COPY . /cloudstack-cloudmonkey
RUN pip install requests
RUN (cd /cloudstack-cloudmonkey; python setup.py build)
RUN (cd /cloudstack-cloudmonkey; python setup.py install)

RUN mkdir -p /cloudmonkey
WORKDIR /cloudmonkey
COPY config.docker /cloudmonkey/config

ENTRYPOINT ["cloudmonkey", "-c", "/cloudmonkey/config"]