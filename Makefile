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

all: build

runtests:
	nosetests -v --verbosity=3

build: buildcache
	python setup.py build
	python setup.py sdist

check:
	pep8 cloudmonkey/*.py

buildcache:
	python cloudmonkey/cachemaker.py
	mv -f precache.py cloudmonkey/

install: clean
	python setup.py sdist
	pip install dist/cloudmonkey-*.tar.gz

clean:
	rm -frv build dist *egg-info
