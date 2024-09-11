package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDownload(t *testing.T) {
	f, err := download()
	assert.Nil(t, err)
	assert.NotEmpty(t, f)
}

func TestJob(t *testing.T) {
	// dest = "./bin/tmp"
	// local = "./bin/xml/20240908_1524.zip"
	err := job()
	assert.Nil(t, err)
}
