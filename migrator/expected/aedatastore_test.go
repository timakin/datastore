package fixture

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"testing"

	"go.mercari.io/datastore"
	netcontext "golang.org/x/net/context"
	"google.golang.org/api/iterator"
	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
)

type AEDatastoreStruct struct {
	Test string
}

func newContext() (context.Context, func(), error) {
	inst, err := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
	if err != nil {
		return nil, nil, err
	}
	r, err := inst.NewRequest("GET", "/", nil)
	if err != nil {
		return nil, nil, err
	}
	ctx := appengine.NewContext(r)
	return ctx, func() { inst.Close() }, nil
}

func TestAEDatastore_Put(t *testing.T) {
	ctx, close, err := newContext()
	if err != nil {
		t.Fatal(err)
	}
	defer close()

	key := client.IncompleteKey("AEDatastoreStruct", nil)
	key, err = client.Put(ctx, key, &AEDatastoreStruct{"Hi!"})
	if err != nil {
		t.Fatal(err.Error())
	}

	t.Logf("key: %s", key.String())
}

func TestAEDatastore_GetMulti(t *testing.T) {
	ctx, close, err := newContext()
	if err != nil {
		t.Fatal(err)
	}
	defer close()

	type Data struct {
		Str string
	}

	key1, err := client.Put(ctx, client.IDKey("Data", 1, nil), &Data{"Data1"})
	if err != nil {
		t.Fatal(err.Error())
	}
	key2, err := client.Put(ctx, client.IDKey("Data", 2, nil), &Data{"Data2"})
	if err != nil {
		t.Fatal(err.Error())
	}

	list := make([]*Data, 2)
	err = client.GetMulti(ctx, []datastore.Key{key1, key2}, list)
	if err != nil {
		t.Fatal(err.Error())
	}

	if v := len(list); v != 2 {
		t.Fatalf("unexpected: %v", v)
	}
}

func TestAEDatastore_TransactionDeleteAndGet(t *testing.T) {
	ctx, close, err := newContext()
	if err != nil {
		t.Fatal(err)
	}
	defer close()

	type Data struct {
		Str string
	}

	key, err := client.Put(ctx, client.IncompleteKey("Data", nil), &Data{"Data"})
	if err != nil {
		t.Fatal(err.Error())
	}
	commit, err = client.RunInTransaction(ctx, func(tx datastore.Transaction) error {
		err := tx.Delete(key)
		if err != nil {
			return err
		}

		obj := &Data{}
		err = tx.Get(key, obj)
		if err != nil {
			t.Fatalf("unexpected: %v", err)
		}

		return nil
	})

	if err != nil {
		t.Fatal(err.Error())
	}
}

func TestAEDatastore_Query(t *testing.T) {
	ctx, close, err := newContext()
	if err != nil {
		t.Fatal(err)
	}
	defer close()

	type Data struct {
		Str string
	}

	_, err = client.Put(ctx, client.IncompleteKey("Data", nil), &Data{"Data1"})
	if err != nil {
		t.Fatal(err.Error())
	}
	_, err = client.Put(ctx, client.IncompleteKey("Data", nil), &Data{"Data2"})
	if err != nil {
		t.Fatal(err.Error())
	}

	q := client.NewQuery("Data").Filter("Str =", "Data2")
	var list []*Data
	_, err = client.GetAll(ctx, q, &list)
	if err != nil {
		t.Fatal(err.Error())
	}

	if v := len(list); v != 1 {
		t.Fatalf("unexpected: %v", v)
	}
}

func TestAEDatastore_QueryCursor(t *testing.T) {
	ctx, close, err := newContext()
	if err != nil {
		t.Fatal(err)
	}
	defer close()

	type Data struct {
		Str string
	}

	{
		var keys []datastore.Key
		var entities []*Data
		for i := 0; i < 100; i++ {
			keys = append(keys, client.IncompleteKey("Data", nil))
			entities = append(entities, &Data{Str: fmt.Sprintf("#%d", i+1)})
		}
		_, err = client.PutMulti(ctx, keys, entities)
		if err != nil {
			t.Fatal(err)
		}
	}

	var cur datastore.Cursor
	var dataList []*Data
	const limit = 3
outer:
	for {
		q := client.NewQuery("Data").Order("Str").Limit(limit)
		if cur.String() != "" {
			q = q.Start(cur)
		}
		it := client.Run(ctx, q)

		count := 0
		for {
			obj := &Data{}
			_, err := it.Next(obj)
			if err == iterator.Done {
				break
			} else if err != nil {
				t.Fatal(err)
			}

			dataList = append(dataList, obj)
			count++
		}
		if count != limit {
			break
		}

		cur, err = it.Cursor()
		if err != nil {
			t.Fatal(err)
		}
		if cur.String() == "" {
			break outer
		}
	}

	if v := len(dataList); v != 100 {
		t.Errorf("unexpected: %v", v)
	}
}

func TestAEDatastore_ErrConcurrentTransaction(t *testing.T) {
	ctx, close, err := newContext()
	if err != nil {
		t.Fatal(err)
	}
	defer close()

	type Data struct {
		Str string
	}

	key := client.NameKey("Data", "a", nil)
	_, err = client.Put(ctx, key, &Data{})
	if err != nil {
		t.Fatal(err)
	}
	commit, // ErrConcurrent will be occur
		err = client.RunInTransaction(ctx, func(tx datastore.Transaction) error {
		err := tx.Get(txCtx1, key, &Data{})
		if err != nil {
			return err
		}
		commit, err = client.RunInTransaction(ctx, func(tx datastore.Transaction) error {
			err := tx.Get(txCtx2, key, &Data{})
			if err != nil {
				return err
			}

			_, err = tx.Put(txCtx2, key, &Data{Str: "#2"})
			return err
		})
		if err != nil {
			return err
		}

		_, err = tx.Put(txCtx1, key, &Data{Str: "#1"})
		return err
	})
	if err != datastore.ErrConcurrentTransaction {
		t.Fatal(err)
	}
}

func TestAEDatastore_ObjectHasObjectSlice(t *testing.T) {
	type Inner struct {
		A string
		B string
	}

	type Data struct {
		Slice []Inner
	}

	ps, err := datastore.SaveStruct(ctx, &Data{
		Slice: []Inner{
			Inner{A: "A1", B: "B1"},
			Inner{A: "A2", B: "B2"},
			Inner{A: "A3", B: "B3"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if v := len(ps); v != 6 {
		t.Fatalf("unexpected: %v", v)
	}

	sort.SliceStable(ps, func(i, j int) bool {
		a := ps[i]
		b := ps[j]
		if v := strings.Compare(a.Name, b.Name); v < 0 {
			return true
		}
		if v := strings.Compare(a.Value.(string), b.Value.(string)); v < 0 {
			return true
		}

		return false
	})

	expects := []struct {
		Name     string
		Value    string
		Multiple bool
	}{
		{"Slice.A", "A1", true},
		{"Slice.A", "A2", true},
		{"Slice.A", "A3", true},
		{"Slice.B", "B1", true},
		{"Slice.B", "B2", true},
		{"Slice.B", "B3", true},
	}
	for idx, expect := range expects {
		t.Logf("idx: %d", idx)
		p := ps[idx]
		if v := p.Name; v != expect.Name {
			t.Errorf("unexpected: %v", v)
		}
		if v := p.Value.(string); v != expect.Value {
			t.Errorf("unexpected: %v", v)
		}
		if v := p.Multiple; v != expect.Multiple {
			t.Errorf("unexpected: %v", v)
		}
	}
}

func TestAEDatastore_GeoPoint(t *testing.T) {
	ctx, close, err := newContext()
	if err != nil {
		t.Fatal(err)
	}
	defer close()

	type Data struct {
		A datastore.GeoPoint
		// B *appengine.GeoPoint
		C []datastore.GeoPoint
		// D []*appengine.GeoPoint
	}

	obj := &Data{
		A: datastore.GeoPoint{1.1, 2.2},
		// B: &appengine.GeoPoint{3.3, 4.4},
		C: []datastore.GeoPoint{
			{5.5, 6.6},
			{7.7, 8.8},
		},
		/*
			D: []*appengine.GeoPoint{
				{9.9, 10.10},
				{11.11, 12.12},
			},
		*/
	}

	key, err := client.Put(ctx, client.IncompleteKey("Data", nil), obj)
	if err != nil {
		t.Fatal(err)
	}

	obj = &Data{}
	err = client.Get(ctx, key, obj)
	if err != nil {
		t.Fatal(err)
	}

	if v := obj.A.Lat; v != 1.1 {
		t.Errorf("unexpected: %v", v)
	}
	if v := obj.A.Lng; v != 2.2 {
		t.Errorf("unexpected: %v", v)
	}

	if v := len(obj.C); v != 2 {
		t.Fatalf("unexpected: %v", v)
	}
	if v := obj.C[0].Lat; v != 5.5 {
		t.Errorf("unexpected: %v", v)
	}
	if v := obj.C[0].Lng; v != 6.6 {
		t.Errorf("unexpected: %v", v)
	}
	if v := obj.C[1].Lat; v != 7.7 {
		t.Errorf("unexpected: %v", v)
	}
	if v := obj.C[1].Lng; v != 8.8 {
		t.Errorf("unexpected: %v", v)
	}
}