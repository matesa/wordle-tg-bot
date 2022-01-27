package bot

import (
	"fmt"
	"wordle-tg-bot/configs"
	tele "gopkg.in/telebot.v3"
	"time"
)

var ANSWER = "İTAAT"
var ANSWER_MAP = map[rune]bool{'İ':false,'T':true,'A':true}
var SinglePlayer *GameStatus = &GameStatus{}

var (
	Bot *tele.Bot
)

type RuneStatus uint8
const (
	Unknown RuneStatus = 0
	WrongRune          = 1
	WrongSpot          = 2
	CorrectRune        = 3
)

type Rune struct {
	Rune rune
	Status RuneStatus
}

type Word struct {
	Saved bool
	Runes []Rune
}

type GameStatus struct {
	CurrentWord uint8
	Player int64
	CorrectWord string
	Words []Word
	Runes map[rune]Rune
	Status int8
}

func (g *GameStatus) InitGame() {
	for i := 0; i < 6; i++ {
		g.Words = append(g.Words, Word{Saved: false, Runes: make([]Rune, 0)})
	}
	g.Runes = map[rune]Rune{}
}

func Create() error {
	var err error
	Bot, err = tele.NewBot(tele.Settings{
		Token:  configs.Get("TG_BOT_TOKEN"),
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
		//Verbose: true,
	})
	if err == nil {
		Bot.Handle("/game", OnGame)
		Bot.Handle(tele.OnCallback, OnCallback)
		Bot.Start()
	}
	return err
}

func OnGame(c tele.Context) error {
	SinglePlayer.InitGame()
	a := SinglePlayer.GameReplyMarkup()
	return c.Reply(SinglePlayer.GameMessageText(), &tele.SendOptions{ReplyMarkup: &a, ParseMode: tele.ModeHTML})
}

func OnCallback(c tele.Context) error {
	ValidRunes := []rune{'E','R','T','Y','U','I','O','P','Ğ','Ü','A','S','D','F','G','H','J','K','L','Ş','İ','Z','C','V','B','N','M','Ö','Ç'}
	cb := c.Callback()
	if len(cb.Data) == 1 || len(cb.Data) == 2 {
		validRune := '0'
		for _, r := range ValidRunes {
			if string(r) == cb.Data {
				validRune = r
				break
			}
		}
		if validRune != '0' {
			i := SinglePlayer.CurrentWord
			if len(SinglePlayer.Words[i].Runes) < 5 {
				SinglePlayer.Words[i].Runes = append(SinglePlayer.Words[i].Runes, Rune{Rune: validRune, Status: Unknown})
				m := c.Message()
				c.Edit(SinglePlayer.GameMessageText(), &tele.SendOptions{ReplyMarkup: m.ReplyMarkup, ParseMode: tele.ModeHTML})
			}
			return c.Respond(&tele.CallbackResponse{})
		}
	} else if cb.Data == "back" {
		i := SinglePlayer.CurrentWord
		if len(SinglePlayer.Words[i].Runes) > 0 {
			SinglePlayer.Words[i].Runes = SinglePlayer.Words[i].Runes[:len(SinglePlayer.Words[i].Runes)-1]
			m := c.Message()
			c.Edit(SinglePlayer.GameMessageText(), &tele.SendOptions{ReplyMarkup: m.ReplyMarkup, ParseMode: tele.ModeHTML})
		}
	} else if cb.Data == "enter" {
		return SinglePlayer.OnEnter(c)
	}
	return c.Respond(&tele.CallbackResponse{})
}

func (r *Rune) KeyboardKeyTextFormat() string {
	text := string(r.Rune)
	switch r.Status {
	case Unknown:
		return fmt.Sprintf("%s", text)
	case WrongRune:
		return fmt.Sprintf("_%s_", text)
	case WrongSpot:
		return fmt.Sprintf("-%s-", text)
	case CorrectRune:
		return fmt.Sprintf("+%s+", text)
	}
	return text
}

func (c *GameStatus) GameMessageText() string {
	message := ""
	for i, word := range c.Words {
		message += fmt.Sprintf("%d. ", i+1)
		for _, run := range word.Runes {
			if run.Rune == 0 {
				message += fmt.Sprintf("<code>[_]</code> ")
				continue
			}
			switch run.Status {
			case Unknown:
				message += fmt.Sprintf("<code>[%c]</code> ", run.Rune)
			case WrongRune:
				message += fmt.Sprintf("<code>[%c]</code> ", run.Rune)
			case WrongSpot:
				message += fmt.Sprintf("<b>[%c]</b> ", run.Rune)
			case CorrectRune:
				message += fmt.Sprintf("<b>%c</b> ", run.Rune)
			}
		}
		if len(word.Runes) < 5 {
			for i := 0; i < 5-len(word.Runes); i++ {
				message += fmt.Sprintf("<code>[_]</code> ")
			}
		}
		if i == int(c.CurrentWord) {
			message += "👈"
		}
		message += "\r\n"
	}
	if c.Status != 0 {
		message += "\r\n"
		switch c.Status {
		case -1:
			message += "<b>Sonuç:</b> X/6"
		default:
			message += fmt.Sprintf("<b>Sonuç:</b> %d/6", c.Status)
		}
	}
	return message
}

func (c *GameStatus) GameReplyMarkup() tele.ReplyMarkup {
	if c.Status != 0 {
		return tele.ReplyMarkup{}
	}
	Keyboard := [][]string{
		{"A","B","C","Ç","D","E","F","G"},
		{"Ğ","H","I","İ","J","K","L","M"},
		{"N","O","Ö","P","R","S","Ş","T"},
		{"U","Ü","V","Y","Z"},
	}
	menu := tele.ReplyMarkup{}
	myrows := make([][]tele.InlineButton, 0)
	for row := range Keyboard {
		mycolumns := make([]tele.InlineButton, 0)
		for column := range Keyboard[row] {
			if _, ok := c.Runes[rune(Keyboard[row][column][0])]; !ok {
				mycolumns = append(mycolumns, tele.InlineButton{Text: fmt.Sprintf("%s", Keyboard[row][column]), Data: Keyboard[row][column]})
			} else {
				val := c.Runes[rune(Keyboard[row][column][0])]
				mycolumns = append(mycolumns, tele.InlineButton{Text: val.KeyboardKeyTextFormat(), Data: Keyboard[row][column]})
			}
		}
		myrows = append(myrows, mycolumns)
	}
	myrows = append(myrows, []tele.InlineButton{tele.InlineButton{Text: "Enter", Data: "enter"}, tele.InlineButton{Text: "Geri", Data: "back"}})
	menu.InlineKeyboard = myrows
	return menu
}

func UpdateKey(r Rune) {
	if v, ok := SinglePlayer.Runes[r.Rune]; ok {
		if v.Status < r.Status {
			v.Status = r.Status
			SinglePlayer.Runes[r.Rune] = v
		}
	} else {
		SinglePlayer.Runes[r.Rune] = r
	}
}

func (g *GameStatus) OnEnter(c tele.Context) error {
	i := g.CurrentWord
	if len(g.Words[i].Runes) != 5 {
		return c.Respond(&tele.CallbackResponse{Text: "Kelimeyi tamamlayın!", ShowAlert: true})
	}
	if true { // is valid word
		answer := []rune(ANSWER)
		correctr := map[rune]int{}
		correctc := 0
		for x, run := range g.Words[i].Runes {
			if run.Rune == answer[x] {
				g.Words[i].Runes[x].Status = CorrectRune
				correctr[run.Rune] = x
				correctc++
			} else if _, ok := ANSWER_MAP[run.Rune]; ok {
				g.Words[i].Runes[x].Status = WrongSpot
			} else {
				g.Words[i].Runes[x].Status = WrongRune
			}
			UpdateKey(g.Words[i].Runes[x])
		}
		for x, run := range g.Words[i].Runes {
			if v, ok := correctr[run.Rune]; ok && ANSWER_MAP[run.Rune] == false && v != x {
				g.Words[i].Runes[x].Status = WrongRune
			}
		}
		if correctc == 5 {
			g.Status = int8(g.CurrentWord)+1
			g.CurrentWord = 6
			mrk := SinglePlayer.GameReplyMarkup()
			c.Edit(SinglePlayer.GameMessageText(), &tele.SendOptions{ReplyMarkup: &mrk, ParseMode: tele.ModeHTML})
			return c.Respond(&tele.CallbackResponse{Text: "Tebrikler!", ShowAlert: true})
		}
		g.Words[i].Saved = true
		g.CurrentWord++
		if g.CurrentWord == 6 {
			g.Status = -1
			mrk := SinglePlayer.GameReplyMarkup()
			c.Edit(SinglePlayer.GameMessageText(), &tele.SendOptions{ReplyMarkup: &mrk, ParseMode: tele.ModeHTML})
			return c.Respond(&tele.CallbackResponse{Text: "Kazanamadınız!", ShowAlert: true})
		}
		mrk := SinglePlayer.GameReplyMarkup()
		c.Edit(SinglePlayer.GameMessageText(), &tele.SendOptions{ReplyMarkup: &mrk, ParseMode: tele.ModeHTML})
		return c.Respond(&tele.CallbackResponse{})
	}
	return c.Respond(&tele.CallbackResponse{Text: "Geçersiz kelime!", ShowAlert: true})
}