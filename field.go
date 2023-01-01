package main

import (
	"encoding/base64"
	"errors"
	"fmt"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

func fillMessageFields(input map[string]any, desc protoreflect.MessageDescriptor) (*dynamicpb.Message, error) {
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
		nested, err := fillMessageFields(m, f.Message())
		if err != nil {
			return protoValue, fmt.Errorf("fail to fill nested message fields: %w", err)
		}
		value = nested

	case protoreflect.GroupKind:
		return protoValue, errors.New("the groups feature unsupported")
	}

	return protoreflect.ValueOf(value), nil
}
