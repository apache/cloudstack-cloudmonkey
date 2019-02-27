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

try:
    from setuptools import setup, find_packages
except ImportError:
    from distribute_setup import use_setuptools
    use_setuptools()
    from setuptools import setup, find_packages

import sys

from cloudmonkey import __version__, __description__
from cloudmonkey import __maintainer__, __maintaineremail__
from cloudmonkey import __project__, __projecturl__, __projectemail__

requires = [
              'jmespath',
              'Pygments>=1.5',
              'argcomplete',
              'dicttoxml',
              'prettytable>=0.6',
              'requests',
              'requests-toolbelt',
            ]

try:
    import readline
except ImportError:
    platform = str(sys.platform).lower()
    if 'win32' in platform or 'win64' in platform:
        requires.append('pyreadline')
    else:
        requires.append('readline')

# Upgrade notes for 5.3.0
print "If you're upgrading, run the following to enable parameter completion:"
print "  cloudmonkey sync"
print "  cloudmonkey set paramcompletion true"
print "Parameter completion may fail, if the above is not run!"

setup(
    name = 'cloudmonkey',
    version = __version__,
    author = __project__,
    author_email = __projectemail__,
    maintainer = __maintainer__,
    maintainer_email = __maintaineremail__,
    url = __projecturl__,
    description = __description__,
    long_description = "cloudmonkey is a CLI for Apache CloudStack",
    platforms = ("Any",),
    license = 'ASL 2.0',
    packages = find_packages(),
    install_requires = requires,
    include_package_data = True,
    zip_safe = False,
    classifiers = [
        "Development Status :: 5 - Production/Stable",
        "Environment :: Console",
        "Intended Audience :: Developers",
        "Intended Audience :: End Users/Desktop",
        "Operating System :: OS Independent",
        "Programming Language :: Python",
        "Topic :: Software Development :: Testing",
        "Topic :: Software Development :: Interpreters",
        "Topic :: Utilities",
    ],
    entry_points="""
    [console_scripts]
    cloudmonkey = cloudmonkey.cloudmonkey:main
    """,
)
