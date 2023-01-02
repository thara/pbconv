package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
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

		msg, err = toProtoMessage(obj, desc)
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

func toProtoMessage(input map[string]any, desc protoreflect.MessageDescriptor) (*dynamicpb.Message, error) {
	msg := dynamicpb.NewMessage(desc)

	fields := desc.Fields()
	for k, v := range input {
		f := fields.ByJSONName(k)
		if f == nil {
			return nil, fmt.Errorf("not found field: %s", k)
		}
		switch {
		case f.IsList():
			a, ok := v.([]any)
			if !ok {
				return nil, errors.New("invalid value as list")
			}
			protoList := msg.Mutable(f).List()
			for _, v := range a {
				fieldValue, err := protoFieldValueOf(f, v)
				if err != nil {
					return nil, fmt.Errorf("fail to convert value in %s: %w", k, err)
				}
				protoList.Append(fieldValue)
			}
			msg.Set(f, protoreflect.ValueOf(protoList))
		case f.IsMap():
			m, ok := v.(map[string]any)
			if !ok {
				return nil, errors.New("invalid value as list")
			}
			protoMap := msg.Mutable(f).Map()
			for k, v := range m {
				fieldValue, err := protoFieldValueOf(f.MapValue(), v)
				if err != nil {
					return nil, fmt.Errorf("fail to convert value in %s: %w", k, err)
				}
				protoMap.Set(protoreflect.ValueOf(k).MapKey(), fieldValue)
			}
			msg.Set(f, protoreflect.ValueOf(protoMap))
		default:
			fieldValue, err := protoFieldValueOf(f, v)
			if err != nil {
				return nil, fmt.Errorf("fail to convert value in %s", k)
			}
			msg.Set(f, fieldValue)
		}
	}

	return msg, nil
}

func protoFieldValueOf(f protoreflect.FieldDescriptor, jsonValue any) (protoValue protoreflect.Value, err error) {
	var value any
	switch f.Kind() {
	case protoreflect.BoolKind:
		v, ok := jsonValue.(bool)
		if !ok {
			return protoValue, errors.New("invalid value as bool")
		}
		value = v
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		v, ok := jsonValue.(float64)
		if !ok {
			return protoValue, errors.New("invalid value as float64")
		}
		value = int32(v)
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		v, ok := jsonValue.(float64)
		if !ok {
			return protoValue, errors.New("invalid value as float64")
		}
		value = int64(v)
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		v, ok := jsonValue.(float64)
		if !ok {
			return protoValue, errors.New("invalid value as float64")
		}
		value = uint32(v)
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		v, ok := jsonValue.(float64)
		if !ok {
			return protoValue, errors.New("invalid value as float64")
		}
		value = uint64(v)
	case protoreflect.FloatKind:
		v, ok := jsonValue.(float64)
		if !ok {
			return protoValue, errors.New("invalid value as float64")
		}
		value = float32(v)
	case protoreflect.DoubleKind:
		v, ok := jsonValue.(float64)
		if !ok {
			return protoValue, errors.New("invalid value as float64")
		}
		value = v
	case protoreflect.StringKind:
		v, ok := jsonValue.(string)
		if !ok {
			return protoValue, errors.New("invalid value as string")
		}
		value = v
	case protoreflect.BytesKind:
		v, ok := jsonValue.(string)
		if !ok {
			return protoValue, errors.New("invalid value as string")
		}
		decoded, err := base64.StdEncoding.DecodeString(v)
		if err != nil {
			return protoValue, errors.New("invalid base64 string")
		}
		value = decoded
	case protoreflect.EnumKind:
		v, ok := jsonValue.(float64)
		if ok {
			value = protoreflect.EnumNumber(v)
			break
		}
		s, ok := jsonValue.(string)
		if !ok {
			return protoValue, errors.New("invalid value as string or float64")
		}
		ev := f.Enum().Values().ByName(protoreflect.Name(s))
		if ev == nil {
			return protoValue, fmt.Errorf("not found enum value: %s", s)
		}
		value = ev.Number()
	case protoreflect.MessageKind:
		m, ok := jsonValue.(map[string]any)
		if !ok {
			return protoValue, errors.New("invalid value as object")
		}
		nested, err := toProtoMessage(m, f.Message())
		if err != nil {
			return protoValue, fmt.Errorf("fail to fill nested message fields: %w", err)
		}
		value = nested

	case protoreflect.GroupKind:
		return protoValue, errors.New("the groups feature unsupported")
	}

	return protoreflect.ValueOf(value), nil
}
