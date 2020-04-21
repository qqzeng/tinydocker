package main

import (
	"os"
	"testing"
)

func TestPahExists(t *testing.T) {
	tmpPath := "/var/run/a/"
	exist, err := PathExists(tmpPath)
	t.Logf("exist : %v, err: %v", exist, err)

	os.MkdirAll(tmpPath, 0622)
	exist, err = PathExists(tmpPath)
	t.Logf("exist : %v, err: %v", exist, err)
}

func PathExists(url string) (bool, error) {
	_, err := os.Stat(url)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
