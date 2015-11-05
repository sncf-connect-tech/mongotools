#!/bin/bash

if [  -d $GOPATH/src/com/vsct/dt/mongotools/ ]
then
	echo "build $GOPATH/src/com/vsct/dt/mongotools/"
else 
	echo "can't find $GOPATH/src/com/vsct/dt/mongotools/"
	echo "GOPATH=$GOPATH, check that \$GOPATH is setted and contains directory src/com/vsct/dt/mongotools/ (see README.md)"
	exit 1
fi

echo "clean build directory: 'rm -Rf ./build'"
rm -Rf build
mkdir build

if [ -f .git/HEAD ]
then
	githead=`cat .git/HEAD | cut -d\  -f2`
	echo "HEAD=$githead"
	githash=`cat .git/$githead`
	echo "add git hash $githash to build/git-hash"
	echo "$githash" > ./build/git-hash
else
	echo "can't find file .git/HEAD"
	echo "you must be in git home directory to build the project"
	exit 1
fi

while [[ $# > 1 ]]
do
	key="$1"
	case $key in
		-sb|--skip-build)
        	SKIP_BUILD="$2"
        	echo "SKIP_BUILD=$2"
		;;
		-r|--release)
        	RELEASE="$2"
        	echo "RELEASE=$2"
		;;
		-n|--nexus-url)
        	NEXUS_URL="$2"
        	echo "NEXUS_URL=$2"
		;;
		-nu|--nexus-user)
        	NEXUS_USER="$2"
        	echo "NEXUS_USER=$2"
		;;
		-np|--nexus-password)
        	NEXUS_PASSWORD="$2"
        	echo "NEXUS_PASSWORD=****"
		;;
		-nr|--nexus-repo)
        	NEXUS_REPO="$2"
        	echo "NEXUS_REPO=$2"
		;;
		*)
		;;
	esac
shift
done

if [ "true" -eq $SKIP_BUILD ];
then
	#tools="mongooplog-tail mongooplog-window mongostat-lag mongostat-parser"
	#arch="386 amd64"
	#os="linux windows"

	tools="mongooplog-tail mongooplog-window mongostat-lag mongostat-parser"
	arch="386"
	os="linux"


	for o in $os; do
		for a in $arch; do
			mkdir -p build/$o/$a/
			for tool in $tools; do
				echo "build/$o/$a/$tool"
				GOARCH=$a
				GOOS=$o
				go build -o build/$o/$a/$tool -a -v -buildmode exe com/vsct/dt/mongotools/$tool
			done;
		done;
	done;
fi

cd build
tar zvcf mongotools.tar.gz *
cd -
echo "end of build: ./mongotools.tar.gz"


if [ ! -z "$RELEASE" ];
then
       	echo "RELEASE=$RELEASE"
       	echo "NEXUS_URL=$NEXUS_URL"
       	echo "NEXUS_USER=$NEXUS_USER"
       	echo "NEXUS_PASSWORD=****"
       	echo "NEXUS_REPO=$NEXUS_REPO"
	if [ -z "$NEXUS_URL"  ] ||  [ -z "$NEXUS_USER" ] || [ -z "$NEXUS_PASSWORD" ] || [ -z "$NEXUS_REPO"  ];
	then
		echo "missing some parameters for a release"
		exit 1
	else
		echo "upload ./build/mongotools.tar.gz to nexus"
		curl -v -F r=$NEXUS_REPO -F hasPom=false -F e=tar.gz -F g=com.vsct.dt.mongotools -F a=mongotools -F v=$RELEASE -F p=tar.gz -F file=@build/mongotools.tar.gz -u $NEXUS_USER:$NEXUS_PASSWORD $NEXUS_URL
	fi
fi
