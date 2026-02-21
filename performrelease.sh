#!/bin/sh
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

version='TESTBUILD'
sourcedir=~/cloudstack-cloudmonkey
outputdir=/tmp/cloudstack-cloudmonkey-build/
branch='main'
tag='no'
certid='X'
committosvn='X'

usage(){
    echo "usage: $0 -v version [-b branch] [-s source dir] [-o output dir] [-t] [-u] [-c] [-h]"
    echo "  -v sets the version"
    echo "  -b sets the branch (defaults to 'main')"
    echo "  -s sets the source directory (defaults to $sourcedir)"
    echo "  -o sets the output directory (defaults to $outputdir)"
    echo "  -t tags the git repo with the version"
    echo "  -u sets the certificate ID to sign with (if not provided, the default key is attempted)"
    echo "  -c commits build artifacts to cloudstack dev dist dir in svn"
    echo "  -h"
}

while getopts v:s:o:b:u:tch opt
do
    case "$opt" in
      v)  version="$OPTARG";;
      s)  sourcedir="$OPTARG";;
      o)  outputdir="$OPTARG";;
      b)  branch="$OPTARG";;
      t)  tag="yes";;
      u)  certid="$OPTARG";;
      c)  committosvn="yes";;
      h)  usage
          exit 0;;
      /?)       # unknown flag
          usage
          exit 1;;
    esac
done
shift `expr $OPTIND - 1`

if [ $version == "TESTBUILD" ]; then
    echo >&2 "A version must be specified with the -v option: build_asf.sh -v 4.0.0.RC1"
    exit 1
fi

echo "Using version: $version"
echo "Using source directory: $sourcedir"
echo "Using output directory: $outputdir"
echo "Using branch: $branch"
if [ "$tag" == "yes" ]; then
    if [ "$certid" == "X" ]; then
        echo "Tagging the branch with the version number, and signing the branch with your default certificate."
    else
        echo "Tagging the branch with the version number, and signing the branch with certificate ID $certid."
    fi
else
    echo "The branch will not be tagged.  You should consider doing this."
fi

if [ -d "$outputdir" ]; then
    rm -r $outputdir/*
else
    mkdir $outputdir
fi

cd $sourcedir

echo 'checking out correct branch'
git checkout $branch

git clean -f

export commitsh=`git show HEAD | head -n 1 | cut -d ' ' -f 2`
echo "releasing as $commitsh"

echo 'archiving'
git archive --format=tar --prefix=apache-cloudstack-cloudmonkey-$version-src/ $branch > $outputdir/apache-cloudstack-cloudmonkey-$version-src.tar
bzip2 $outputdir/apache-cloudstack-cloudmonkey-$version-src.tar

cd $outputdir
echo 'armor'
if [ "$certid" == "X" ]; then
  gpg -v --armor --output apache-cloudstack-cloudmonkey-$version-src.tar.bz2.asc --detach-sig apache-cloudstack-cloudmonkey-$version-src.tar.bz2
else
  gpg -v --default-key $certid --armor --output apache-cloudstack-cloudmonkey-$version-src.tar.bz2.asc --detach-sig apache-cloudstack-cloudmonkey-$version-src.tar.bz2
fi

echo 'sha512'
sha512sum apache-cloudstack-cloudmonkey-$version-src.tar.bz2 > apache-cloudstack-cloudmonkey-$version-src.tar.bz2.sha512

echo 'verify'
gpg -v --verify apache-cloudstack-cloudmonkey-$version-src.tar.bz2.asc apache-cloudstack-cloudmonkey-$version-src.tar.bz2

if [ "$tag" == "yes" ]; then
  echo 'tag'
  cd $sourcedir
  if [ "$certid" == "X" ]; then
      git tag -s $version -m "Tagging release $version on branch $branch."
  else
      git tag -u $certid -s $version -m "Tagging release $version on branch $branch."
  fi
fi

if [ "$committosvn" == "yes" ]; then
  echo 'committing artifacts to svn'
  rm -Rf /tmp/cloudstack-dev-dist
  cd /tmp
  svn co https://dist.apache.org/repos/dist/dev/cloudstack/ cloudstack-dev-dist
  cd cloudstack-dev-dist
  if [ -d "cloudmonkey-$version" ]; then
    cd cloudmonkey-$version
    svn rm *
  else
    mkdir cloudmonkey-$version
    svn add cloudmonkey-$version
    cd cloudmonkey-$version
  fi
  cp $outputdir/apache-cloudstack-cloudmonkey-$version-src.tar.bz2 .
  cp $outputdir/apache-cloudstack-cloudmonkey-$version-src.tar.bz2.asc .
  cp $outputdir/apache-cloudstack-cloudmonkey-$version-src.tar.bz2.sha512 .
  svn add apache-cloudstack-cloudmonkey-$version-src.tar.bz2
  svn add apache-cloudstack-cloudmonkey-$version-src.tar.bz2.asc
  svn add apache-cloudstack-cloudmonkey-$version-src.tar.bz2.sha512
  svn commit -m "Committing release candidate artifacts for $version to dist/dev/cloudstack in preparation for release vote"
fi

echo "completed.  use commit-sh of $commitsh when starting the VOTE thread"
