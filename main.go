package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/bufbuild/protocompile"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type format string

const (
	formatJSON  format = "json"
	formatText  format = "text"
	formatProto format = "proto"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s to-proto|from-proto ...\n", os.Args[0])
		os.Exit(1)
	}

	subCommandName := os.Args[1]

	if err := run(subCommandName); err != nil {
		log.Fatal(err)
	}
}

func run(subCommandName string) error {
	switch subCommandName {
	case "to-proto":
		if err := toProtoCommand.Parse(os.Args[2:]); err != nil {
			return fmt.Errorf("fail to parse flags: %v", err)
		}
		return toProto()
	case "from-proto":
		if err := fromProtoCommand.Parse(os.Args[2:]); err != nil {
			return fmt.Errorf("fail to parse flags: %v", err)
		}
		return fromProto()
	default:
		fmt.Printf("Usage: %s to-proto|from-proto ...\n", os.Args[0])
	}
	return nil
}

func resolveMessageDescriptor(messageName string, paths []string) (protoreflect.MessageDescriptor, error) {
	compiler := protocompile.Compiler{
		Resolver: &protocompile.SourceResolver{},
	}

	ctx := context.Background()

	fs, err := compiler.Compile(ctx, paths...)
	if err != nil {
		return nil, fmt.Errorf("fail to compile proto: %w", err)
	}

	var desc protoreflect.MessageDescriptor
	for _, ds := range fs {
		ms := ds.Messages()
		d := ms.ByName(protoreflect.Name(messageName))
		if d != nil {
			desc = d
		}
	}
	if desc == nil {
		return nil, fmt.Errorf("not found message: %s", messageName)
	}
	return desc, nil
}
