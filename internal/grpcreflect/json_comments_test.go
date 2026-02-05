package grpcreflect

import (
	"strings"
	"testing"
)

func TestGenerateJSONExampleWithComments(t *testing.T) {
	tests := []struct {
		name     string
		msg      *MessageInfo
		expected []string
	}{
		{
			name: "simple oneof fields",
			msg: &MessageInfo{
				Name: "TestMessage",
				Fields: []FieldInfo{
					{
						Name:       "regular_field",
						Type:       "string",
						OneofGroup: "",
					},
					{
						Name:       "oneof_field1",
						Type:       "string",
						OneofGroup: "my_oneof",
					},
					{
						Name:       "oneof_field2",
						Type:       "int32",
						OneofGroup: "my_oneof",
					},
				},
			},
			expected: []string{
				`"regular_field": ""`,
				`// oneof my_oneof (choose one):`,
				`"oneof_field1": ""`,
				`"oneof_field2": 0`,
			},
		},
		{
			name: "nested message with oneof",
			msg: &MessageInfo{
				Name: "OuterMessage",
				Fields: []FieldInfo{
					{
						Name: "nested",
						Type: "NestedMessage",
						Message: &MessageInfo{
							Name: "NestedMessage",
							Fields: []FieldInfo{
								{
									Name:       "field1",
									Type:       "string",
									OneofGroup: "nested_oneof",
								},
								{
									Name:       "field2",
									Type:       "string",
									OneofGroup: "nested_oneof",
								},
							},
						},
					},
				},
			},
			expected: []string{
				`"nested": {`,
				`// oneof nested_oneof (choose one):`,
				`"field1": ""`,
				`"field2": ""`,
			},
		},
		{
			name: "proto3 optional field (single field in oneof)",
			msg: &MessageInfo{
				Name: "MessageWithOptional",
				Fields: []FieldInfo{
					{
						Name:       "status",
						Type:       "Status",
						IsEnum:     true,
						OneofGroup: "_status",
						EnumValues: []EnumValueInfo{
							{Name: "STATUS_UNKNOWN", Number: 0},
							{Name: "STATUS_ACTIVE", Number: 1},
						},
					},
				},
			},
			expected: []string{
				`"status": "STATUS_UNKNOWN"`,
			},
		},
		{
			name: "multiple oneof groups",
			msg: &MessageInfo{
				Name: "MultiOneofMessage",
				Fields: []FieldInfo{
					{
						Name:       "field1",
						Type:       "string",
						OneofGroup: "group1",
					},
					{
						Name:       "field2",
						Type:       "string",
						OneofGroup: "group1",
					},
					{
						Name:       "field3",
						Type:       "int32",
						OneofGroup: "group2",
					},
					{
						Name:       "field4",
						Type:       "int32",
						OneofGroup: "group2",
					},
				},
			},
			expected: []string{
				`// oneof group1 (choose one):`,
				`"field1": ""`,
				`"field2": ""`,
				`// oneof group2 (choose one):`,
				`"field3": 0`,
				`"field4": 0`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateJSONExampleWithComments(tt.msg)

			for _, expected := range tt.expected {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected to find %q in result, but it wasn't there.\nResult:\n%s", expected, result)
				}
			}

			lines := strings.Split(result, "\n")
			oneofCommentCount := 0
			for _, line := range lines {
				if strings.Contains(line, "// oneof") {
					oneofCommentCount++
				}
			}

			t.Logf("Generated JSON:\n%s", result)
		})
	}
}

func TestGenerateJSONExampleWithComments_RealWorld(t *testing.T) {
	msg := &MessageInfo{
		Name: "GetMessagesListRequest",
		Fields: []FieldInfo{
			{
				Name:       "application_id",
				Type:       "string",
				OneofGroup: "",
			},
			{
				Name:       "message_id",
				Type:       "int64",
				OneofGroup: "_message_id",
			},
			{
				Name:       "limit",
				Type:       "uint32",
				OneofGroup: "_limit",
			},
			{
				Name:       "sort",
				Type:       "SortOptions",
				OneofGroup: "_sort",
				Message: &MessageInfo{
					Name: "SortOptions",
					Fields: []FieldInfo{
						{
							Name:       "order_by_date",
							Type:       "SortOrder",
							IsEnum:     true,
							OneofGroup: "sort_field",
							EnumValues: []EnumValueInfo{
								{Name: "SORT_UNSPECIFIED", Number: 0},
								{Name: "SORT_ASCENDING", Number: 1},
								{Name: "SORT_DESCENDING", Number: 2},
							},
						},
						{
							Name:       "order_by_segments_count",
							Type:       "SortOrder",
							IsEnum:     true,
							OneofGroup: "sort_field",
							EnumValues: []EnumValueInfo{
								{Name: "SORT_UNSPECIFIED", Number: 0},
								{Name: "SORT_ASCENDING", Number: 1},
								{Name: "SORT_DESCENDING", Number: 2},
							},
						},
					},
				},
			},
		},
	}

	result := GenerateJSONExampleWithComments(msg)

	expectedPatterns := []string{
		`"application_id": ""`,
		`"message_id": 0`,
		`"limit": 0`,
		`"sort": {`,
		`// oneof sort_field (choose one):`,
		`"order_by_date": "SORT_UNSPECIFIED"`,
		`"order_by_segments_count": "SORT_UNSPECIFIED"`,
	}

	unexpectedPatterns := []string{
		`// oneof _message_id`,
		`// oneof _limit`,
		`// oneof _sort`,
	}

	for _, pattern := range expectedPatterns {
		if !strings.Contains(result, pattern) {
			t.Errorf("Expected to find %q in result", pattern)
		}
	}

	for _, pattern := range unexpectedPatterns {
		if strings.Contains(result, pattern) {
			t.Errorf("Did NOT expect to find %q in result (proto3 optional should not have comments)", pattern)
		}
	}

	t.Logf("Generated real-world JSON:\n%s", result)
}
