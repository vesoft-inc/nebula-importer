package base

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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
	assert.True(t, IsValidType("String"))
	assert.True(t, IsValidType("STRING"))
	assert.True(t, IsValidType("sTring"))
	assert.True(t, IsValidType("int"))
	assert.True(t, IsValidType("float"))
	assert.True(t, IsValidType("date"))
	assert.False(t, IsValidType("byte"))
	assert.True(t, IsValidType("datetime"))
	assert.True(t, IsValidType("bool"))
	assert.True(t, IsValidType("timestamp"))
	assert.True(t, IsValidType("double"))
	assert.True(t, IsValidType("geography"))
	assert.True(t, IsValidType("geography(point)"))
	assert.True(t, IsValidType("geography(linestring)"))
	assert.True(t, IsValidType("geography(polygon)"))
}
