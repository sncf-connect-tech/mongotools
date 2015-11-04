#!/bin/bash

if [  -d $GOPATH/src/com/vsct/dt/mongotools/ ]
then
	echo "build $GOPATH/src/com/vsct/dt/mongotools/"
else 
	echo "can't find $GOPATH/src/com/vsct/dt/mongotools/"
	echo "GOPATH=$GOPATH, check that \$GOPATH is setted and contains directory src/com/vsct/dt/mongotools/ (see README.md)"
	exit 1
fi

tools="mongooplog-tail mongooplog-window mongostat-lag mongostat-parser"
arch="386 amd64"
os="linux windows"

echo "clean build directory: 'rm -Rf ./build'"
rm -Rf build

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

if [ -f .git/refs/heads/master ]
then
	echo "add git hash `cat .git/refs/heads/master` to build/git-hash"
	cat .git/refs/heads/master > build/git-hash
else
	echo "can't find file .git/refs/heads/master"
	echo "you must be in git home directory to build the project"
fi

cd build
tar zvcf mongotools.tar.gz *
cd -
echo "end of build: ./mongotools.tar.gz"
