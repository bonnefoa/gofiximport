gofiximport
========

gofiximport remove unused imports and add missing imports in a golang source file.


Installation
------------

    go get github.com/bonnefoa/gofiximport

gofiximport should be available in your $GOPATH/bin

Usage
------------

To output result in stdout, just use:

    gofiximport -file <file>

To write modification in the source file:

    gofiximport -file <file> -w
