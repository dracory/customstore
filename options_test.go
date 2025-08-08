// Package customstore_test provides black-box tests for the customstore package.
package customstore_test

import (
    "reflect"
    "testing"

    "github.com/dracory/customstore"
)

func TestWithID(t *testing.T) {
    r := customstore.NewRecord("test", customstore.WithID("my-id-123"))
    if r.ID() != "my-id-123" {
        t.Fatalf("WithID not applied, expected my-id-123, got %s", r.ID())
    }
}

func TestWithMemo(t *testing.T) {
    r := customstore.NewRecord("test", customstore.WithMemo("hello memo"))
    if r.Memo() != "hello memo" {
        t.Fatalf("WithMemo not applied, expected 'hello memo', got %s", r.Memo())
    }
}

func TestWithPayload(t *testing.T) {
    json := `{"a":1,"b":"two"}`
    r := customstore.NewRecord("test", customstore.WithPayload(json))
    if r.Payload() != json {
        t.Fatalf("WithPayload not applied, expected %s, got %s", json, r.Payload())
    }
}

func TestWithPayloadMap(t *testing.T) {
    mp := map[string]any{"x": 9, "y": "yes"}
    r := customstore.NewRecord("test", customstore.WithPayloadMap(mp))
    got, err := r.PayloadMap()
    if err != nil {
        t.Fatalf("PayloadMap returned error: %v", err)
    }
    if !reflect.DeepEqual(got, map[string]any{"x": float64(9), "y": "yes"}) {
        t.Fatalf("WithPayloadMap not applied, expected %v, got %v", mp, got)
    }
}

func TestWithMetas(t *testing.T) {
    metas := map[string]string{"k1": "v1", "k2": "v2"}
    r := customstore.NewRecord("test", customstore.WithMetas(metas))
    got, err := r.Metas()
    if err != nil {
        t.Fatalf("Metas returned error: %v", err)
    }
    if !reflect.DeepEqual(got, metas) {
        t.Fatalf("WithMetas not applied, expected %v, got %v", metas, got)
    }
}
