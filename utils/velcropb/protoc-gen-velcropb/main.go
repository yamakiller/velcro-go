package main

import (
	"fmt"
	"os"
	"path/filepath"

	gengo "google.golang.org/protobuf/cmd/protoc-gen-go/internal_gengo"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/yamakiller/velcro-go/utils/velcropb"
	genvelcropb "github.com/yamakiller/velcro-go/utils/velcropb/protoc-gen-velcropb/generator"
)

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--version" {
		fmt.Fprintf(os.Stdout, "%s %s\n", filepath.Base(os.Args[0]), velcropb.Version)
		os.Exit(0)
	}
	if len(os.Args) == 2 && os.Args[1] == "--help" {
		fmt.Fprintf(os.Stdout, "See %s for usage information.\n", velcropb.Home)
		os.Exit(0)
	}

	protogen.Options{}.Run(func(gen *protogen.Plugin) error {
		// check: only support proto3 now
		for _, f := range gen.Files {
			if f.Desc.Syntax() != protoreflect.Proto3 {
				return nil
			}

			// 这里开始构建代码
			for _, f := range gen.Files {
				if f.Generate {
					genvelcropb.GenerateFile(gen, f)
				}
			}
			gen.SupportedFeatures = gengo.SupportedFeatures
		}
		return nil
	})
}
