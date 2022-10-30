package parser

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestParsingBasicProblemFile(t *testing.T) {
	// Arrange
	filename := "../../problems.csv"

	// Act
	_, err := ParseProblemFile(filename)

	// Assert
	assert.Equal(t, err, nil)
}
