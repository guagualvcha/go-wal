package wal

import (
	"io/ioutil"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSIGHUP(t *testing.T) {
	// First, create an AutoFile writing to a tempfile dir
	file, err := ioutil.TempFile("", "sighup_test")
	require.NoError(t, err)
	err = file.Close()
	require.NoError(t, err)
	name := file.Name()

	// Here is the actual AutoFile
	af, err := OpenAutoFile(name)
	require.NoError(t, err)

	// Write to the file.
	_, err = af.Write([]byte("Line 1\n"))
	require.NoError(t, err)
	_, err = af.Write([]byte("Line 2\n"))
	require.NoError(t, err)

	// Move the file over
	err = os.Rename(name, name+"_old")
	require.NoError(t, err)

	// Send SIGHUP to self.
	syscall.Kill(syscall.Getpid(), syscall.SIGHUP)

	// Wait a bit... signals are not handled synchronously.
	time.Sleep(time.Millisecond * 10)

	// Write more to the file.
	_, err = af.Write([]byte("Line 3\n"))
	require.NoError(t, err)
	_, err = af.Write([]byte("Line 4\n"))
	require.NoError(t, err)
	err = af.Close()
	require.NoError(t, err)

	// Both files should exist
	fileBytes, err := ioutil.ReadFile(name + "_old")
	require.NoError(t, err)

	if string(fileBytes) != "Line 1\nLine 2\n" {
		t.Errorf("Unexpected body %s", fileBytes)
	}
	fileBytes, err = ioutil.ReadFile(name)
	require.NoError(t, err)

	if string(fileBytes) != "Line 3\nLine 4\n" {
		t.Errorf("Unexpected body %s", fileBytes)
	}
}

func TestAutoFileSize(t *testing.T) {
	// First, create an AutoFile writing to a tempfile dir
	f, err := ioutil.TempFile("", "sighup_test")
	require.NoError(t, err)
	err = f.Close()
	require.NoError(t, err)

	// Here is the actual AutoFile.
	af, err := OpenAutoFile(f.Name())
	require.NoError(t, err)

	// 1. Empty file
	size, err := af.Size()
	require.Zero(t, size)
	require.NoError(t, err)

	// 2. Not empty file
	data := []byte("Maniac\n")
	_, err = af.Write(data)
	require.NoError(t, err)
	size, err = af.Size()
	require.EqualValues(t, len(data), size)
	require.NoError(t, err)

	// 3. Not existing file
	err = af.Close()
	require.NoError(t, err)
	err = os.Remove(f.Name())
	require.NoError(t, err)
	size, err = af.Size()
	require.EqualValues(t, 0, size, "Expected a new file to be empty")
	require.NoError(t, err)

	// Cleanup
	_ = os.Remove(f.Name())
}
