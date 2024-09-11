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
	dest = "./tmp"
	local = "20240909_1459.zip"
	err := job()
	assert.Nil(t, err)
}
