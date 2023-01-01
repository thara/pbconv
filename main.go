package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/bufbuild/protocompile"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var verbose bool

func init() {
	flag.BoolVar(&verbose, "v", false, "verbose mode")
}

func main() {
	flag.Parse()

	messageName := flag.Args()[0]
	paths := flag.Args()[1:]

	r := bufio.NewReader(os.Stdin)
	s, err := r.ReadString('\n')
	if err != nil && err != io.EOF {
		log.Fatalf("fail to read stdin: %v", err)
	}
	if !json.Valid([]byte(s)) {
		log.Fatal("invalid json input")
	}

	if err := run(s, messageName, paths...); err != nil {
		log.Fatal(err)
	}
}

func run(jsonString string, messageName string, paths ...string) error {
	var input map[string]any
	if err := json.Unmarshal([]byte(jsonString), &input); err != nil {
		return fmt.Errorf("fail to unmarshal json: %w", err)
	}

	compiler := protocompile.Compiler{
		Resolver: &protocompile.SourceResolver{},
	}

	ctx := context.Background()

	fs, err := compiler.Compile(ctx, paths...)
	if err != nil {
		return fmt.Errorf("fail to compile proto: %w", err)
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
		return fmt.Errorf("not found message: %s", messageName)
	}

	msg, err := fillMessageFields(input, desc)
	if err != nil {
		return fmt.Errorf("fail to message fields: %w", err)
	}

	if verbose {
		fmt.Fprintln(os.Stderr, prototext.Format(msg))
	}

	result, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("fail to marshal proto: %w", err)
	}
	os.Stdout.Write(result)

	return nil
}
