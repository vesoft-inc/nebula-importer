package base

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestFileExists(t *testing.T) {
	fileName := "test.csv"
	var isExist bool
	isExist = FileExists(fileName)
	assert.False(t, isExist)
	file := MustCreateFile(fileName)
	isExist = FileExists(fileName)
	assert.True(t, isExist)
	file.Close()
	os.Remove(fileName)
}

func TestIsValidType(t *testing.T) {
	assert.True(t, IsValidType("string"))
	assert.True(t, IsValidType("int"))
	assert.True(t, IsValidType("float"))
	assert.False(t, IsValidType("date"))
	assert.False(t, IsValidType("byte"))
	assert.False(t, IsValidType("datetime"))
	assert.True(t, IsValidType("bool"))
	assert.True(t, IsValidType("timestamp"))
	assert.True(t, IsValidType("double"))
}
