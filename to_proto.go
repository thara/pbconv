package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/dynamicpb"
)

var toProtoCommand = flag.NewFlagSet("to-proto", flag.ExitOnError)

var inputFormat string
var protoDstPath string

func init() {
	toProtoCommand.StringVar(&inputFormat, "from", "json", "input data format: json|text")
	toProtoCommand.StringVar(&protoDstPath, "out", "out.proto", "proto destination path")
}

func toProto() error {
	if toProtoCommand.NArg() < 2 {
		fmt.Printf("Usage: %s to-proto [options] PROTO_MESSAGE_NAME PROTO_FILES...\n", os.Args[0])
		toProtoCommand.PrintDefaults()
		fmt.Println("\nRead input data from stdin.")
		os.Exit(1)
	}

	r := bufio.NewReader(os.Stdin)
	input, err := r.ReadString('\n')
	if err != nil && err != io.EOF {
		return fmt.Errorf("fail to read stdin: %v", err)
	}

	messageName := toProtoCommand.Args()[0]
	paths := toProtoCommand.Args()[1:]

	desc, err := resolveMessageDescriptor(messageName, paths)
	if err != nil {
		return fmt.Errorf("fail to resolve message descriptor: %w", err)
	}

	var msg *dynamicpb.Message
	switch format(inputFormat) {
	case formatJSON:
		if !json.Valid([]byte(input)) {
			return errors.New("invalid json input")
		}

		var obj map[string]any
		err := json.Unmarshal([]byte(input), &obj)
		if err != nil {
			return fmt.Errorf("fail to unmarshal json: %w", err)
		}

		msg, err = fillMessageFields(obj, desc)
		if err != nil {
			return fmt.Errorf("fail to message fields: %w", err)
		}
	case formatText:
		//TODO
	default:
		return fmt.Errorf("unsupported input format: %s", inputFormat)
	}

	result, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("fail to marshal proto: %w", err)
	}

	f, err := os.Create(protoDstPath)
	if err != nil {
		return fmt.Errorf("fail to create a file: %w", err)
	}
	defer f.Close()

	_, err = f.Write(result)
	if err != nil {
		return fmt.Errorf("fail to write a file: %w", err)
	}
	return nil
}
