//go:build mage
// +build mage

package main

import (
	"github.com/magefile/mage/sh"
)

var (
	//Default = Build
	//targets = []string{"skinnyd", "skinnyctl"}
	//protos = []string{"consensus"}
	protos = []string{"BFTBaxos", "application"}
)

// Install build dependencies.
func BuildDeps() error {
	err := sh.RunV("protoc", "--version")
	if err != nil {
		return err
	}
	err = sh.RunV("go", "install", "google.golang.org/protobuf/cmd/protoc-gen-go@latest")
	if err != nil {
		return err
	}
	err = sh.RunV("go", "install", "google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest")
	if err != nil {
		return err
	}

	return nil
}

// Install dependencies.
func Deps() error {
	err := sh.RunV("go", "mod", "vendor")
	if err != nil {
		return err
	}

	return nil
}

// Generate code.
func Generate() error {
	for _, name := range protos {
		//protoc --go_out=./ --go-grpc_out=./ .\product.proto
		err := sh.RunV("protoc", "--go_out=./proto/"+name+"/", "--go-grpc_out=./proto/"+name+"/", "./proto/"+name+"/"+name+".proto")
		if err != nil {
			return err
		}
	}
	return nil
}

// Run tests.
//func Test() error {
//	mg.Deps(Generate)
//
//	return sh.RunV("go", "test", "-v", "./...")
//}

// Build binary executables.
//func Build() error {
//	for _, name := range targets {
//		err := sh.RunV("go", "build", "-v", "-o", "./bin/"+name, "./cmd/"+name)
//		if err != nil {
//			return err
//		}
//	}
//
//	return nil
//}

// Remove binary executables.
//func Clean() error {
//	return os.RemoveAll("bin")
//}
