package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/korber710/golang_playground/quiz-game/pkg/quiz"
	"github.com/urfave/cli/v2"
)

func main() {
	var problemFile string

	app := &cli.App{
		Name:     "quiz-game",
		Version:  "v0.0.1",
		Compiled: time.Now(),
		Authors: []*cli.Author{
			&cli.Author{
				Name:  "Steve Korber",
				Email: "korbersa@outlook.com",
			},
		},
		Copyright: "(c) 2022 Korber Solutions",
		Usage:     "a simple math quiz in go!",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "file",
				Value:       "problems.csv",
				Usage:       "read problems from `FILE`",
				Destination: &problemFile,
			},
		},
		Action: func(cCtx *cli.Context) error {
			fmt.Println("Welcome to the game!")
			fmt.Println("Problem file:", problemFile)

			// Create a new quiz
			q, err := quiz.NewQuiz(problemFile)
			if err != nil {
				return err
			}

			// Loop through and ask questions until the user stops
			gameActive := ""
			for gameActive != "stop" {
				// Check with user if they want to play
				reader := bufio.NewReader(os.Stdin)
				fmt.Printf("Do you want to play? (enter `stop` to quit): ")
				gameActive, _ = reader.ReadString('\n')

				// Convert CRLF to LF
				gameActive = strings.Replace(gameActive, "\n", "", -1)

				// Play the game with the user
				if gameActive != "stop" {
					q.AskQuestion(os.Stdin)
				}
			}

			fmt.Println("Thanks for playing!")
			fmt.Println("Total questions asked:", q.QuestionsAsked)
			fmt.Println("Total correct answers:", q.CorrectAnswers)

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
