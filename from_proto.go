package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/dynamicpb"
)

var fromProtoCommand = flag.NewFlagSet("from-proto", flag.ExitOnError)

var outputFormat string
var protoSrcPath string

func init() {
	fromProtoCommand.StringVar(&outputFormat, "to", "json", "output data format: json|text")
	fromProtoCommand.StringVar(&protoSrcPath, "in", "", "[required] proto source path")
}

func fromProto() error {
	if fromProtoCommand.NArg() < 2 {
		fmt.Printf("Usage: %s from-proto [options] PROTO_MESSAGE_NAME PROTO_FILES...\n", os.Args[0])
		fromProtoCommand.PrintDefaults()
		os.Exit(1)
	}

	if len(protoSrcPath) == 0 {
		return errors.New("--from option is required")
	}

	messageName := fromProtoCommand.Args()[0]
	paths := fromProtoCommand.Args()[1:]

	desc, err := resolveMessageDescriptor(messageName, paths)
	if err != nil {
		return fmt.Errorf("fail to resolve message descriptor: %w", err)
	}

	input, err := os.ReadFile(protoSrcPath)
	if err != nil {
		return fmt.Errorf("fail to open %s: %w", protoSrcPath, err)
	}

	msg := dynamicpb.NewMessage(desc)

	pm := msg.New().Interface()
	if err := proto.Unmarshal(input, pm); err != nil {
		return fmt.Errorf("fail to unmarshal proto: %w", err)
	}

	switch format(outputFormat) {
	case formatJSON:
		fmt.Println(protojson.Format(pm))
	case formatText:
		fmt.Println(prototext.Format(pm))
	default:
		fmt.Fprintf(os.Stderr, "unsupported output format: %s", outputFormat)
	}
	return nil
}
