package exporter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
)

type (
	JSONHandler interface {
		Build(ctx context.Context, writer *JSONWriter) error
	}

	JSONWriter struct {
		data any
	}

	jsonExporter struct {
		fileName string
		handler  JSONHandler
	}
)

func NewJSONExporter(fileName string, handler JSONHandler) Exporter {
	return &jsonExporter{
		fileName: fileName,
		handler:  handler,
	}
}

func (e *jsonExporter) FileName() string {
	return e.fileName
}

func (e *jsonExporter) Export(ctx context.Context, w Writer) (*bytes.Buffer, error) {
	jsonWriter, ok := w.(*JSONWriter)
	if !ok {
		return nil, fmt.Errorf("invalid writer type for JSONExporter")
	}
	if err := e.handler.Build(ctx, jsonWriter); err != nil {
		return nil, err
	}
	return jsonWriter.WriteToBuffer()
}

func NewJSONWriter(data any) *JSONWriter {
	return &JSONWriter{data: data}
}

func (w *JSONWriter) WriteToBuffer() (*bytes.Buffer, error) {
	b, err := json.MarshalIndent(w.data, "", "  ")
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(b), nil
}
