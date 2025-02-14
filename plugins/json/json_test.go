package main

import (
	"fmt"
	"io"
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/falcosecurity/plugin-sdk-go/pkg/sdk"
)

type testEventReader struct {
	num      uint64
	jsonData string
	time     time.Time
}

func (t *testEventReader) EventNum() uint64 {
	return t.num
}

func (t *testEventReader) Timestamp() uint64 {
	return uint64(t.time.UnixNano())
}

func (t *testEventReader) Reader() io.ReadSeeker {
	return strings.NewReader(t.jsonData)
}

type testExtractRequest struct {
	fieldID   uint64
	fieldType uint32
	arg       string
	field     string
	value     interface{}
}

func (t *testExtractRequest) FieldID() uint64 {
	return t.fieldID
}

func (t *testExtractRequest) FieldType() uint32 {
	return t.fieldType
}

func (t *testExtractRequest) Field() string {
	return t.field
}

func (t *testExtractRequest) Arg() string {
	return t.arg
}

func (t *testExtractRequest) SetValue(v interface{}) {
	t.value = v
}

func (t *testExtractRequest) SetPtr(unsafe.Pointer) {
	// do nothing
}

func TestExtractValue(t *testing.T) {
	var s string
	var ok bool
	var err error
	testEvent := &testEventReader{
		num:  1,
		time: time.Now(),
		jsonData: `{
			"list":[
				{
					"intvalue":1,
					"floatvalue":2.5
				}
			],
			"value":"hello",
			"~/escaped":"hello\"2"
		}`,
	}
	testRequest := &testExtractRequest{
		fieldID:   0,
		field:     "json.value",
		arg:       "",
		fieldType: sdk.FieldTypeCharBuf,
		value:     nil,
	}
	e := &MyPlugin{}

	// invalid json pointer
	testRequest.arg = "invalid_pointer"
	if err := e.Extract(testRequest, testEvent); err != nil {
		t.Error(err)
	}

	// valid json pointer with string value
	testRequest.arg = "/value"
	err = e.Extract(testRequest, testEvent)
	if err != nil {
		t.Error(err)
	}
	if s, ok = testRequest.value.(string); !ok {
		t.Errorf("expected string value")
	}
	if s != "hello" {
		t.Errorf("expected value %s, but found %s", "hello", s)
	}

	// valid json pointer with string value, nesting, and int conversion
	testRequest.arg = "/list/0/intvalue"
	err = e.Extract(testRequest, testEvent)
	if err != nil {
		t.Error(err)
	}
	if s, ok = testRequest.value.(string); !ok {
		t.Errorf("expected string value")
	}
	if s != "1" {
		t.Errorf("expected value %s, but found %s", "1", s)
	}

	// valid json pointer with u64 value, nesting, and float conversion
	testRequest.arg = "/list/0/floatvalue"
	err = e.Extract(testRequest, testEvent)
	if err != nil {
		t.Error(err)
	}
	if s, ok = testRequest.value.(string); !ok {
		t.Errorf("expected string value")
	}
	if s != "2.5" {
		t.Errorf("expected value %s, but found %s", "2.5", s)
	}

	// test jevt.value alias too
	testRequest.fieldID = 3
	testRequest.field = "jevt.value"

	// json pointer with escaping
	testRequest.arg = "/~0~1escaped"
	err = e.Extract(testRequest, testEvent)
	if err != nil {
		t.Error(err)
	}
	if s, ok = testRequest.value.(string); !ok {
		t.Errorf("expected string value")
	}
	if s != "hello\"2" {
		t.Errorf("expected value %s, but found %s", "hello\"2", s)
	}
}

func TestExtractObject(t *testing.T) {
	var s string
	var ok bool
	var err error
	testIndentedJSON := "{\n  \"value\": \"hello\"\n}"
	testEvent := &testEventReader{
		num:      1,
		time:     time.Now(),
		jsonData: `{"value":"hello"}`,
	}
	testRequest := &testExtractRequest{
		fieldID:   1,
		field:     "json.object",
		arg:       "",
		fieldType: sdk.FieldTypeCharBuf,
		value:     nil,
	}
	e := &MyPlugin{}

	// extract object with json.obj
	err = e.Extract(testRequest, testEvent)
	if err != nil {
		t.Error(err)
	}
	if s, ok = testRequest.value.(string); !ok {
		t.Errorf("expected string value")
	}
	if s != testIndentedJSON {
		t.Errorf("expected value '%s', but found '%s'", testIndentedJSON, s)
	}

	// extract object with jevt.obj
	testRequest.fieldID = 4
	testRequest.field = "jevt.obj"
	err = e.Extract(testRequest, testEvent)
	if err != nil {
		t.Error(err)
	}
	if s, ok = testRequest.value.(string); !ok {
		t.Errorf("expected string value")
	}
	if s != testIndentedJSON {
		t.Errorf("expected value '%s', but found '%s'", testIndentedJSON, s)
	}
}

func TestExtractRawtime(t *testing.T) {
	var ts string
	var ok bool
	var err error
	testEvent := &testEventReader{
		num:      1,
		time:     time.Now(),
		jsonData: "{}",
	}
	testTimeStr := fmt.Sprintf("%d", testEvent.time.UnixNano())
	testRequest := &testExtractRequest{
		fieldID:   2,
		field:     "json.rawtime",
		arg:       "",
		fieldType: sdk.FieldTypeUint64,
		value:     nil,
	}
	e := &MyPlugin{}

	// extract object with json.rawtime
	err = e.Extract(testRequest, testEvent)
	if err != nil {
		t.Error(err)
	}
	if ts, ok = testRequest.value.(string); !ok {
		t.Errorf("expected uint64 value")
	}
	if ts != testTimeStr {
		t.Errorf("expected value '%s', but found '%s'", testTimeStr, ts)
	}

	// extract object with jevt.rawtime
	testRequest.fieldID = 5
	testRequest.field = "jevt.rawtime"
	err = e.Extract(testRequest, testEvent)
	if err != nil {
		t.Error(err)
	}
	if ts, ok = testRequest.value.(string); !ok {
		t.Errorf("expected uint64 value")
	}
	if ts != testTimeStr {
		t.Errorf("expected value '%s', but found '%s'", testTimeStr, ts)
	}
}
