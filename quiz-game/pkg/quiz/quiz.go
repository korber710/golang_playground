package quiz

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"

	"github.com/korber710/golang_playground/quiz-game/pkg/parser"
)

type QuizEntry struct {
	Question string
	Answer   int
}

type QuizGame struct {
	Problems       []QuizEntry
	CorrectAnswers int
	TotalQuestions int
	QuestionsAsked int
	questionIndex  int
}

func NewQuiz(filename string) (q *QuizGame, err error) {
	// Create new QuizGame object
	q = new(QuizGame)

	// Parse the filename into an array
	fmt.Println("filename:", filename)
	datas, err := parser.ParseProblemFile(filename)
	if err != nil {
		log.Fatal(err)
	}

	// Add data to quiz
	err = q.ExtractEntries(datas)

	return q, err
}

func (q *QuizGame) ExtractEntries(datas [][]string) (err error) {
	// Loop through all entries
	for _, data := range datas {
		// Create new QuizEntry
		qEntry := new(QuizEntry)

		// Add the question
		qEntry.Question = data[0]

		// Parse the answer as an integer
		trimmedData := strings.Trim(data[1], " ")
		qEntry.Answer, err = strconv.Atoi(trimmedData)
		if err != nil {
			log.Fatal(err)
		}

		// Add the entry
		q.Problems = append(q.Problems, *qEntry)
	}

	// Set total problems
	q.TotalQuestions = len(q.Problems)

	fmt.Printf("%+v\n", q)

	return err
}

func (q *QuizGame) AskQuestion(stdin io.Reader) {
	reader := bufio.NewReader(stdin)

	fmt.Printf("Please answer %s: ", q.Problems[q.questionIndex].Question)
	text, _ := reader.ReadString('\n')

	// Convert CRLF to LF
	text = strings.Replace(text, "\n", "", -1)

	// fmt.Println("text:", text)

	// Convert input to integer
	userAnswer, err := strconv.Atoi(text)
	if err != nil {
		log.Fatal(err)
	}

	// Compare answers
	if userAnswer == q.Problems[q.questionIndex].Answer {
		q.CorrectAnswers += 1
	}

	// Increment question
	q.QuestionsAsked += 1
	q.questionIndex += 1
}

func (q *QuizGame) CheckIndex() {
	fmt.Println("questionIndex:", q.questionIndex)
}
