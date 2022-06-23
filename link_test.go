package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInterestingLink(t *testing.T) {

	var follow, err = interestingLink("https://github.com/pricing")
	assert.NoError(t, err)
	assert.False(t, follow)

	follow, err = interestingLink("https://github.com/kmulvey/trashmap/blob/main/go.mod")
	assert.NoError(t, err)
	assert.False(t, follow)

	follow, err = interestingLink("https://github.com/kmulvey/trashmap/blob/0528fe74a2ad29d592ea0a367914dd1822e3e402/.goreleaser.yaml")
	assert.NoError(t, err)
	assert.False(t, follow)

	follow, err = interestingLink("https://github.com/sindresorhus/awesome")
	assert.NoError(t, err)
	assert.True(t, follow)

	follow, err = interestingLink("https://github.com/kmulvey/trashmap/blob/main/README.md")
	assert.NoError(t, err)
	assert.True(t, follow)
}
