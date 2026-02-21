package services

import (
	"bytes"
	"fmt"
	"reflect"
	"time"

	"github.com/apache/arrow/go/v14/arrow"
	"github.com/apache/arrow/go/v14/arrow/array"
	"github.com/apache/arrow/go/v14/arrow/ipc"
	"github.com/apache/arrow/go/v14/arrow/memory"
)

// ArrowSerializer converts raw database rows into an Apache Arrow IPC stream
type ArrowSerializer struct {
	allocator memory.Allocator
}

// NewArrowSerializer creates a new ArrowSerializer
func NewArrowSerializer() *ArrowSerializer {
	return &ArrowSerializer{
		allocator: memory.NewGoAllocator(),
	}
}

// ToIPCStream takes column names and dynamic rows and returns an Arrow IPC byte buffer
func (s *ArrowSerializer) ToIPCStream(columns []string, rows [][]interface{}) ([]byte, error) {
	if len(columns) == 0 {
		return nil, fmt.Errorf("no columns provided")
	}

	// 1. Infer Arrow Schema based on the first row's data types
	fields := make([]arrow.Field, len(columns))
	typeInferrers := make([]arrow.DataType, len(columns))

	for i, colName := range columns {
		fields[i] = arrow.Field{Name: colName, Type: arrow.BinaryTypes.String, Nullable: true}
		typeInferrers[i] = arrow.BinaryTypes.String

		if len(rows) > 0 {
			val := rows[0][i]
			if val != nil {
				switch v := val.(type) {
				case int, int8, int16, int32, int64:
					fields[i].Type = arrow.PrimitiveTypes.Int64
					typeInferrers[i] = arrow.PrimitiveTypes.Int64
				case float32, float64:
					fields[i].Type = arrow.PrimitiveTypes.Float64
					typeInferrers[i] = arrow.PrimitiveTypes.Float64
				case bool:
					fields[i].Type = arrow.FixedWidthTypes.Boolean
					typeInferrers[i] = arrow.FixedWidthTypes.Boolean
				case time.Time:
					fields[i].Type = arrow.FixedWidthTypes.Timestamp_ms
					typeInferrers[i] = arrow.FixedWidthTypes.Timestamp_ms
				default:
					_ = v // keep as String
				}
			}
		}
	}

	schema := arrow.NewSchema(fields, nil)
	recordBuilder := array.NewRecordBuilder(s.allocator, schema)
	defer recordBuilder.Release()

	// 2. Append Data to Builders
	for _, row := range rows {
		for i, val := range row {
			if val == nil {
				recordBuilder.Field(i).AppendNull()
				continue
			}

			switch typeInferrers[i] {
			case arrow.PrimitiveTypes.Int64:
				builder := recordBuilder.Field(i).(*array.Int64Builder)
				v := reflect.ValueOf(val)
				if v.CanInt() {
					builder.Append(v.Int())
				} else {
					builder.AppendNull()
				}
			case arrow.PrimitiveTypes.Float64:
				builder := recordBuilder.Field(i).(*array.Float64Builder)
				v := reflect.ValueOf(val)
				if v.CanFloat() {
					builder.Append(v.Float())
				} else {
					builder.AppendNull()
				}
			case arrow.FixedWidthTypes.Boolean:
				builder := recordBuilder.Field(i).(*array.BooleanBuilder)
				if v, ok := val.(bool); ok {
					builder.Append(v)
				} else {
					builder.AppendNull()
				}
			case arrow.FixedWidthTypes.Timestamp_ms:
				builder := recordBuilder.Field(i).(*array.TimestampBuilder)
				if t, ok := val.(time.Time); ok {
					builder.Append(arrow.Timestamp(t.UnixMilli()))
				} else {
					builder.AppendNull()
				}
			default:
				builder := recordBuilder.Field(i).(*array.StringBuilder)
				builder.Append(fmt.Sprintf("%v", val))
			}
		}
	}

	record := recordBuilder.NewRecord()
	defer record.Release()

	// 3. Serialize to IPC Stream format
	var buf bytes.Buffer
	writer := ipc.NewWriter(&buf, ipc.WithSchema(schema))

	if err := writer.Write(record); err != nil {
		writer.Close()
		return nil, fmt.Errorf("failed to write arrow record: %v", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close arrow writer: %v", err)
	}

	return buf.Bytes(), nil
}
