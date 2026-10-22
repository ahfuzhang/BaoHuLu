package codegen

import (
	"fmt"
	"sort"
)

func memoryLayout(values []any) (map[string]any, error) {
	fields := make([]map[string]any, len(values))
	for index, value := range values {
		field, ok := value.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("layout requires dictionaries")
		}
		for _, key := range []string{"Size", "Align", "Ptrdata", "Number"} {
			if _, ok := field[key].(int); !ok {
				return nil, fmt.Errorf("layout requires integer %s", key)
			}
		}
		size, align, ptrdata := field["Size"].(int), field["Align"].(int), field["Ptrdata"].(int)
		if size < 0 || size > 1<<30 || align <= 0 || align > 1<<16 || align&(align-1) != 0 || ptrdata < 0 || ptrdata > size {
			return nil, fmt.Errorf("invalid layout dimensions")
		}
		fields[index] = field
	}
	sort.SliceStable(fields, func(left, right int) bool {
		first, second := fields[left], fields[right]
		firstSize, secondSize := first["Size"].(int), second["Size"].(int)
		if (firstSize == 0) != (secondSize == 0) {
			return firstSize == 0
		}
		if first["Align"] != second["Align"] {
			return first["Align"].(int) > second["Align"].(int)
		}
		firstPtr, secondPtr := first["Ptrdata"].(int), second["Ptrdata"].(int)
		if (firstPtr == 0) != (secondPtr == 0) {
			return secondPtr == 0
		}
		if firstPtr > 0 && firstSize-firstPtr != secondSize-secondPtr {
			return firstSize-firstPtr < secondSize-secondPtr
		}
		if firstSize != secondSize {
			return firstSize > secondSize
		}
		return first["Number"].(int) < second["Number"].(int)
	})
	offset, ptrEnd, maxAlign := 0, 0, 1
	ordered := make([]any, 0, len(fields))
	for _, field := range fields {
		align := field["Align"].(int)
		offset = (offset + align - 1) &^ (align - 1)
		maxAlign = max(maxAlign, align)
		if ptrdata := field["Ptrdata"].(int); ptrdata > 0 {
			ptrEnd = max(ptrEnd, offset+ptrdata)
		}
		offset += field["Size"].(int)
		if offset > 1<<30 {
			return nil, fmt.Errorf("layout exceeds size limit")
		}
		ordered = append(ordered, field)
	}
	offset = (offset + maxAlign - 1) &^ (maxAlign - 1)
	return map[string]any{"Fields": ordered, "Size": offset, "Ptrdata": ptrEnd}, nil
}
