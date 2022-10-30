package quiz

import (
	"strings"
	"testing"

	"gotest.tools/v3/assert"
)

func TestNewQuizGame(t *testing.T) {
	// Arrange
	programFile := "../../problems.csv"

	// Act
	q, err := NewQuiz(programFile)

	// Assert
	assert.Equal(t, err, nil)
	assert.Equal(t, q.CorrectAnswers, 0)
	assert.Equal(t, q.TotalQuestions, 13)
}

func TestPlayingBasicGame(t *testing.T) {
	// Arrange
	programFile := "../../problems.csv"
	q, err := NewQuiz(programFile)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	// Act
	input := strings.NewReader("10\n")
	q.AskQuestion(input)
	input = strings.NewReader("1\n")
	q.AskQuestion(input)

	// Assert
	assert.Equal(t, q.QuestionsAsked, 2)
	assert.Equal(t, q.CorrectAnswers, 1)
}
