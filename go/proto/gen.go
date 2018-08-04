// +build ignore

//go:generate go run gen.go

package main

import (
	"log"
	"os"
	"path"
	"runtime"
	"text/template"
	"time"
)

// Any proto which will be marshalled as a single capnp message needs to be listed here.
var RootTypes = []string{
	"APNAPkt",
	"ASEntry",
	"CtrlPld",
	"PathSegment",
	"PathSegmentSignedData",
	"SCIONDMsg",
	"SignedBlob",
	"SignedCtrlPld",
}

func main() {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("No caller information")
	}
	f, err := os.Create(path.Join(path.Dir(filename), "structs.gen.go"))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	fileTmpl.Execute(f, struct {
		Timestamp time.Time
		RootTypes []string
	}{
		time.Now(),
		RootTypes,
	})
}

var fileTmpl = template.Must(template.New("").Parse(`// Code generated by go generate; DO NOT EDIT.
// This file was generated by gen.go @ {{ .Timestamp }}

package proto

import (
	"zombiezen.com/go/capnproto2"

	"github.com/scionproto/scion/go/lib/common"
)

// NewRootStruct calls the appropriate NewRoot<x> function corresponding to the capnp proto type ID,
// and returns the inner capnp.Struct that it receives. This allows the helper
// functions in cereal.go to support generic capnp root struct types.
func NewRootStruct(id ProtoIdType, seg *capnp.Segment) (capnp.Struct, error) {
	var blank capnp.Struct
	switch id {
	{{- range .RootTypes }}
	case {{.}}_TypeID:
		v, err := NewRoot{{.}}(seg)
		if err != nil {
			return blank, common.NewBasicError("Error creating new {{.}} capnp struct", err)
		}
		return v.Struct, nil
	{{- end }}
	}
	return blank, common.NewBasicError(
		"Unsupported capnp struct type (i.e. not listed in go/proto/gen.go:RootTypes)",
		nil,
		"id", id,
	)
}
{{ range .RootTypes }}
func (s {{.}}) GetStruct() capnp.Struct {
	return s.Struct
}
{{- end }}
`))
