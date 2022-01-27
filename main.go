package main

import "wordle-tg-bot/bot"

func main() {
	err := bot.Create()
	if err != nil {
		panic(err)
	}
}
