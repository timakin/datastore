package testsuite

import (
	"context"
	"fmt"
	"testing"

	"go.mercari.io/datastore"
)

func PutAndGet(t *testing.T, ctx context.Context, client datastore.Client) {
	defer func() {
		err := client.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	type TestEntity struct {
		String string
	}

	key := client.IncompleteKey("Test", nil)
	t.Log(key)
	newKey, err := client.Put(ctx, key, &TestEntity{String: "Test"})
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("new key: %s", newKey.String())

	entity := &TestEntity{}
	err = client.Get(ctx, newKey, entity)
	if err != nil {
		t.Fatal(err)
	}

	if v := entity.String; v != "Test" {
		t.Errorf("unexpected: %v", v)
	}
}

func PutAndDelete(t *testing.T, ctx context.Context, client datastore.Client) {
	defer func() {
		err := client.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	type TestEntity struct {
		String string
	}

	key := client.IncompleteKey("Test", nil)
	t.Log(key)
	newKey, err := client.Put(ctx, key, &TestEntity{String: "Test"})
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("new key: %s", newKey.String())

	err = client.Delete(ctx, newKey)
	if err != nil {
		t.Fatal(err)
	}

	entity := &TestEntity{}
	err = client.Get(ctx, newKey, entity)
	if err != datastore.ErrNoSuchEntity {
		t.Fatal(err)
	}
}

func PutAndGet_ObjectHasObjectSlice(t *testing.T, ctx context.Context, client datastore.Client) {
	if IsAEDatastoreClient(ctx) {
		// flatten options must required in ae.
		t.SkipNow()
	}

	defer func() {
		err := client.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	type Inner struct {
		A string
		B string
	}

	type Data struct {
		Slice []Inner // `datastore:",flatten"` // If flatten removed, aedatastore env will fail.
	}

	key := client.NameKey("Test", "a", nil)
	_, err := client.Put(ctx, key, &Data{
		Slice: []Inner{
			Inner{A: "A1", B: "B1"},
			Inner{A: "A2", B: "B2"},
			Inner{A: "A3", B: "B3"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	obj := &Data{}
	err = client.Get(ctx, key, obj)
	if err != nil {
		t.Fatal(err)
	}

	if v := len(obj.Slice); v != 3 {
		t.Errorf("unexpected: %v", v)
	}

	for idx, s := range obj.Slice {
		if v := s.A; v != fmt.Sprintf("A%d", idx+1) {
			t.Errorf("unexpected: %v", v)
		}
		if v := s.B; v != fmt.Sprintf("B%d", idx+1) {
			t.Errorf("unexpected: %v", v)
		}
	}
}

func PutAndGet_ObjectHasObjectSliceWithFlatten(t *testing.T, ctx context.Context, client datastore.Client) {
	defer func() {
		err := client.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	type Inner struct {
		A string
		B string
	}

	type Data struct {
		Slice []Inner `datastore:",flatten"`
	}

	key := client.NameKey("Test", "a", nil)
	_, err := client.Put(ctx, key, &Data{
		Slice: []Inner{
			Inner{A: "A1", B: "B1"},
			Inner{A: "A2", B: "B2"},
			Inner{A: "A3", B: "B3"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	obj := &Data{}
	err = client.Get(ctx, key, obj)
	if err != nil {
		t.Fatal(err)
	}

	if v := len(obj.Slice); v != 3 {
		t.Errorf("unexpected: %v", v)
	}

	for idx, s := range obj.Slice {
		if v := s.A; v != fmt.Sprintf("A%d", idx+1) {
			t.Errorf("unexpected: %v", v)
		}
		if v := s.B; v != fmt.Sprintf("B%d", idx+1) {
			t.Errorf("unexpected: %v", v)
		}
	}
}

func PutEntityType(t *testing.T, ctx context.Context, client datastore.Client) {
	defer func() {
		err := client.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	type Inner struct {
		A string
		B string
	}

	type DataA struct {
		C *Inner
	}

	type DataB struct {
		C *Inner `datastore:",flatten"`
	}

	key := client.IncompleteKey("Test", nil)
	_, err := client.Put(ctx, key, &DataA{
		C: &Inner{
			A: "a",
			B: "b",
		},
	})
	if IsAEDatastoreClient(ctx) {
		if err != datastore.ErrInvalidEntityType {
			t.Fatal(err)
		}
	} else {
		if err != nil {
			t.Fatal(err)
		}
	}

	_, err = client.Put(ctx, key, &DataB{
		C: &Inner{
			A: "a",
			B: "b",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func PutAndGetNilKey(t *testing.T, ctx context.Context, client datastore.Client) {
	defer func() {
		err := client.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	type Data struct {
		KeyA datastore.Key
		KeyB datastore.Key
	}

	key := client.IncompleteKey("Test", nil)
	key, err := client.Put(ctx, key, &Data{
		KeyA: client.NameKey("Test", "a", nil),
		KeyB: nil,
	})
	if err != nil {
		t.Fatal(err)
	}

	obj := &Data{}
	err = client.Get(ctx, key, obj)
	if err != nil {
		t.Fatal(err)
	}

	if v := obj.KeyA; v == nil {
		t.Errorf("unexpected: %v", v)
	}
	if v := obj.KeyB; v != nil {
		t.Errorf("unexpected: %v", v)
	}
}

func PutAndGetNilKeySlice(t *testing.T, ctx context.Context, client datastore.Client) {
	defer func() {
		err := client.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	type Data struct {
		Keys []datastore.Key
	}

	key := client.IncompleteKey("Test", nil)
	key, err := client.Put(ctx, key, &Data{
		Keys: []datastore.Key{
			client.NameKey("Test", "a", nil),
			nil,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	obj := &Data{}
	err = client.Get(ctx, key, obj)
	if err != nil {
		t.Fatal(err)
	}

	if v := len(obj.Keys); v != 2 {
		t.Fatalf("unexpected: %v", v)
	}
	if v := obj.Keys[0]; v == nil {
		t.Errorf("unexpected: %v", v)
	}
	if v := obj.Keys[1]; v != nil {
		t.Errorf("unexpected: %v", v)
	}
}

type EntityInterface interface {
	Kind() string
	ID() string
}

type PutInterfaceTest struct {
	kind string
	id   string
}

func (e *PutInterfaceTest) Kind() string {
	return e.kind
}
func (e *PutInterfaceTest) ID() string {
	return e.id
}

func PutInterface(t *testing.T, ctx context.Context, client datastore.Client) {
	defer func() {
		err := client.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	var e EntityInterface
	e = &PutInterfaceTest{}

	key := client.IncompleteKey("Test", nil)
	_, err := client.Put(ctx, key, e)
	if err != nil {
		t.Fatal(err)
	}
}
