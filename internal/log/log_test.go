package log

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/gogo/protobuf/proto"
	api "github.com/lisp-ceo/dlog/api/v1"
	"github.com/stretchr/testify/require"
)

func TestLog(t *testing.T) {
	for scenario, fn := range map[string]func(
		t *testing.T, log *Log,
	){
		"append and read a record suceeds": testAppendRead,
		"offset out of range error":        testOutOfRangeErr,
		"init with existing segments":      testInitExisting,
		"reader":                           testReader,
		"truncate":                         testTruncate,
	} {
		t.Run(scenario, func(t *testing.T) {
			dir, err := ioutil.TempDir("", "store-test")
			require.NoError(t, err)

			defer os.RemoveAll(dir)
			c := Config{}
			c.Segment.MaxStoreBytes = 32

			log, err := NewLog(dir, c)
			require.NoError(t, err)

			fn(t, log)
		})
	}

}

func testTruncate(t *testing.T, log *Log) {
	r := &api.Record{
		Value: []byte("hello world"),
	}
	for i := 0; i < 3; i++ {
		_, err := log.Append(r)
		require.NoError(t, err)
	}

	err := log.Truncate(1)
	require.NoError(t, err)

	// Reading a truncated record returns `offset out of range`
	_, err = log.Read(0)
	require.Error(t, err)
}

func testReader(t *testing.T, log *Log) {
	r := &api.Record{
		Value: []byte("hello world"),
	}
	off, err := log.Append(r)
	require.NoError(t, err)
	require.Equal(t, uint64(0), off)

	reader := log.Reader()
	b, err := ioutil.ReadAll(reader)
	require.NoError(t, err)

	read := &api.Record{}
	err = proto.Unmarshal(b[lenWidth:], read)
	require.NoError(t, err)
	// Record read back from reader is equal to one saved
	require.Equal(t, r, read)
}

func testInitExisting(t *testing.T, log *Log) {
	r := &api.Record{
		Value: []byte("hello world"),
	}

	// Write 3 entries to the log and close.
	for i := 0; i < 3; i++ {
		_, err := log.Append(r)
		require.NoError(t, err)
	}
	require.NoError(t, log.Close())

	off, err := log.LowestOffset()
	require.NoError(t, err)
	require.Equal(t, uint64(0), off)

	off, err = log.HighestOffset()
	require.NoError(t, err)
	require.Equal(t, uint64(2), off)

	// Re-initializing the log using existing state on disk
	n, err := NewLog(log.Dir, log.Config)
	require.NoError(t, err)

	off, err = n.LowestOffset()
	require.NoError(t, err)
	require.Equal(t, uint64(0), off)

	off, err = n.HighestOffset()
	require.NoError(t, err)
	require.Equal(t, uint64(2), off)
}

func testOutOfRangeErr(t *testing.T, log *Log) {
	read, err := log.Read(1)
	require.Nil(t, read)

	apiErr := err.(api.ErrOffsetOutOfRange)
	require.Equal(t, uint64(1), apiErr.Offset)
}

func testAppendRead(t *testing.T, log *Log) {
	r := &api.Record{
		Value: []byte("hello world"),
	}
	off, err := log.Append(r)
	require.NoError(t, err)
	require.Equal(t, uint64(0), off)

	read, err := log.Read(off)
	require.NoError(t, err)
	require.Equal(t, r, read)
}
