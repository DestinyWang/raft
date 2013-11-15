package raft

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
)

func LevelDBTestStore(t *testing.T) (string, *LevelDBLogStore, *LevelDBStableStore) {
	// Create a test dir
	dir, err := ioutil.TempDir("", "raft")
	if err != nil {
		t.Fatalf("err: %v ", err)
	}

	// New level
	logs, err := NewLevelDBLogStore(dir)
	if err != nil {
		t.Fatalf("err: %v ", err)
	}

	// New level
	stable, err := NewLevelDBStableStore(dir)
	if err != nil {
		t.Fatalf("err: %v ", err)
	}

	return dir, logs, stable
}

func TestLevelDB_StableStore(t *testing.T) {
	var l interface{} = &LevelDBStableStore{}
	_, ok := l.(StableStore)
	if !ok {
		t.Fatalf("LevelDBStableStore is not StableStore")
	}
}

func TestLevelDB_SetGet(t *testing.T) {
	// Create a test dir
	dir, err := ioutil.TempDir("", "raft")
	if err != nil {
		t.Fatalf("err: %v ", err)
	}
	defer os.RemoveAll(dir)

	// New level
	l, err := NewLevelDBStableStore(dir)
	if err != nil {
		t.Fatalf("err: %v ", err)
	}
	defer l.Close()

	// Get a bad key
	key := []byte("foobar")
	_, err = l.Get(key)
	if err.Error() != "not found" {
		t.Fatalf("err: %v ", err)
	}

	val := []byte("this is a test value")
	if err := l.Set(key, val); err != nil {
		t.Fatalf("err: %v ", err)
	}

	out, err := l.Get(key)
	if err != nil {
		t.Fatalf("err: %v ", err)
	}

	if bytes.Compare(val, out) != 0 {
		t.Fatalf("did not get result back: %v %v", val, out)
	}
}

func TestLevelDB_SetGetUint64(t *testing.T) {
	// Create a test dir
	dir, err := ioutil.TempDir("", "raft")
	if err != nil {
		t.Fatalf("err: %v ", err)
	}
	defer os.RemoveAll(dir)

	// New level
	l, err := NewLevelDBStableStore(dir)
	if err != nil {
		t.Fatalf("err: %v ", err)
	}
	defer l.Close()

	// Get a bad key
	key := []byte("dolla bills")
	_, err = l.GetUint64(key)
	if err.Error() != "not found" {
		t.Fatalf("err: %v ", err)
	}

	var val uint64 = 42000
	if err := l.SetUint64(key, val); err != nil {
		t.Fatalf("err: %v ", err)
	}

	out, err := l.GetUint64(key)
	if err != nil {
		t.Fatalf("err: %v ", err)
	}

	if out != val {
		t.Fatalf("did not get result back: %v %v", val, out)
	}
}

func TestLevelDB_LogStore(t *testing.T) {
	var l interface{} = &LevelDBLogStore{}
	_, ok := l.(LogStore)
	if !ok {
		t.Fatalf("LevelDBLogStore is not a LogStore")
	}
}

func TestLevelDB_Logs(t *testing.T) {
	// Create a test dir
	dir, err := ioutil.TempDir("", "raft")
	if err != nil {
		t.Fatalf("err: %v ", err)
	}
	defer os.RemoveAll(dir)

	// New level
	l, err := NewLevelDBLogStore(dir)
	if err != nil {
		t.Fatalf("err: %v ", err)
	}
	defer l.Close()

	// Should be no first index
	idx, err := l.FirstIndex()
	if err != nil {
		t.Fatalf("err: %v ", err)
	}
	if idx != 0 {
		t.Fatalf("bad idx: %d", idx)
	}

	// Should be no last index
	idx, err = l.LastIndex()
	if err != nil {
		t.Fatalf("err: %v ", err)
	}
	if idx != 0 {
		t.Fatalf("bad idx: %d", idx)
	}

	// Try a filed fetch
	var out Log
	if err := l.GetLog(10, &out); err.Error() != "log not found" {
		t.Fatalf("err: %v ", err)
	}

	// Write out a log
	log := Log{
		Index: 1,
		Term:  1,
		Type:  LogCommand,
		Data:  []byte("first"),
	}
	if err := l.StoreLog(&log); err != nil {
		t.Fatalf("err: %v", err)
	}
	log = Log{
		Index: 10,
		Term:  3,
		Type:  LogCommand,
		Data:  []byte("test"),
	}
	if err := l.StoreLog(&log); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Try to fetch
	if err := l.GetLog(10, &out); err != nil {
		t.Fatalf("err: %v ", err)
	}

	// Check the lowest index
	idx, err = l.FirstIndex()
	if err != nil {
		t.Fatalf("err: %v ", err)
	}
	if idx != 1 {
		t.Fatalf("bad idx: %d", idx)
	}

	// Check the highest index
	idx, err = l.LastIndex()
	if err != nil {
		t.Fatalf("err: %v ", err)
	}
	if idx != 10 {
		t.Fatalf("bad idx: %d", idx)
	}

	// Delete the entire range
	if err := l.DeleteRange(1, 10); err != nil {
		t.Fatalf("err: %v ", err)
	}

	// Index should be zero again
	idx, err = l.FirstIndex()
	if err != nil {
		t.Fatalf("err: %v ", err)
	}
	if idx != 0 {
		t.Fatalf("bad idx: %d", idx)
	}
	idx, err = l.LastIndex()
	if err != nil {
		t.Fatalf("err: %v ", err)
	}
	if idx != 0 {
		t.Fatalf("bad idx: %d", idx)
	}

	// Should not be able to fetch
	if err := l.GetLog(10, &out); err.Error() != "log not found" {
		t.Fatalf("err: %v ", err)
	}
}
