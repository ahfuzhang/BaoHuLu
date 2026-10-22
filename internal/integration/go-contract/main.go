package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reflect"

	data "github.com/ahfuzhang/BaoHuLu/examples/Demo"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(arguments []string) error {
	if len(arguments) != 2 || arguments[0] != "write" && arguments[0] != "verify" {
		return fmt.Errorf("usage: go run . write|verify <vector-directory>")
	}
	root := arguments[1]
	sample := data.AllTypes{
		DoubleValue: 1.25, FloatValue: -2.5, Int32Value: math.MinInt32, Int64Value: math.MinInt64,
		Uint32Value: math.MaxInt32, Uint64Value: math.MaxUint64, Sint32Value: -123, Sint64Value: -456,
		Fixed32Value: math.MaxInt32, Fixed64Value: math.MaxInt64, Sfixed32Value: -100, Sfixed64Value: -200,
		BoolValue: true, StringValue: "hello \"世界\" \\\n", BytesValue: []byte{0, 1, 127, 255}, Status: data.StatusActive,
		Child: data.Child{ChildID: 9, ChildName: "child"}, Numbers: []int32{-1, 0, math.MaxInt32}, Tags: []string{"first", "二"},
		Scores: map[string]int32{"score": 3}, BigTag: 1234,
		M1: map[int32]int32{-1: 1}, M2: map[int64]int32{math.MinInt64: 2}, M3: map[uint32]int32{math.MaxUint32: 3},
		M4: map[uint64]int32{math.MaxInt64: 4}, M5: map[int32]int32{-5: 5}, M6: map[int64]int32{-6: 6},
		M7: map[uint32]int32{7: 7}, M8: map[uint64]int32{8: 8}, M9: map[int32]int32{-9: 9}, M10: map[int64]int32{-10: 10},
		M13: map[string][]byte{"bytes": {0, 255}}, M14: map[string]data.Status{"state": data.StatusDisabled},
		Values: map[string]*data.ValueTypes{"value": {Int64Value: math.MinInt64, Uint64Value: math.MaxUint64, BoolValue: true}},
		Next:   &data.AllTypes{Int32Value: 2}, Children: map[uint64]*data.AllTypes{1: {Int32Value: 3}},
	}
	expected := sample.ToJSON(nil)
	if arguments[0] == "write" {
		if err := os.MkdirAll(root, 0755); err != nil {
			return err
		}
		for name, content := range map[string][]byte{"go.pb": sample.ToProtobuf(nil), "go.json": expected} {
			if err := os.WriteFile(filepath.Join(root, name), content, 0644); err != nil {
				return err
			}
		}
		sample.M4 = map[uint64]int32{math.MaxUint64: 4}
		sample.Fixed64Value = math.MaxUint64
		sample.Uint32Value = math.MaxUint32
		sample.Fixed32Value = math.MaxUint32
		if err := os.WriteFile(filepath.Join(root, "go-unsigned-map.pb"), sample.ToProtobuf(nil), 0644); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(root, "go-unsigned.json"), sample.ToJSON(nil), 0644); err != nil {
			return err
		}
		fmt.Println("Go AllTypes protobuf/JSON vectors written")
		return nil
	}
	for _, name := range []string{"csharp.pb", "oracle.pb", "csharp.json"} {
		input, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			return err
		}
		reader := data.ReadonlyAllTypes{}
		if filepath.Ext(name) == ".json" {
			err = reader.FromJSON(input)
		} else {
			err = reader.FromProtobuf(input)
		}
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		clone := reader.Clone(nil)
		reader.Reset()
		for index := range input {
			input[index] = 255
		}
		if err := equalJSON(expected, clone.ToJSON(nil)); err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
	}
	fmt.Println("AllTypes: Go/C# JSON+Protobuf, Google.Protobuf oracle, Clone/Reset passed")
	return nil
}

func equalJSON(expected, actual []byte) error {
	decode := func(content []byte) (any, error) {
		decoder := json.NewDecoder(bytes.NewReader(content))
		decoder.UseNumber()
		var value any
		err := decoder.Decode(&value)
		return value, err
	}
	left, err := decode(expected)
	if err != nil {
		return err
	}
	right, err := decode(actual)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(left, right) {
		return fmt.Errorf("JSON differs:\nexpected %s\nactual %s", expected, actual)
	}
	return nil
}
