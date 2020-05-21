package main

import (
	"log"
	"os"

	tb "gopkg.in/tucnak/telebot.v2"
)

var B *tb.Bot

var ReplyMain = &tb.SendOptions{
	ParseMode: tb.ModeHTML,
	ReplyMarkup: &tb.ReplyMarkup{
		ResizeReplyKeyboard: true,
		ReplyKeyboard: [][]tb.ReplyButton{
			{
				tb.ReplyButton{Text: "🏁 Гонка"},
			},
			{
				tb.ReplyButton{Text: "🐌 Улитки"},
				tb.ReplyButton{Text: "💰 Деньги"},
				tb.ReplyButton{Text: "❓ Помощь"},
			},
		},
	},
}

func main() {

	var (
		port      = os.Getenv("PORT")
		publicURL = os.Getenv("PUBLIC_URL")
		token     = os.Getenv("TOKEN")
		err       error
	)

	poller := &tb.Webhook{
		Listen:   ":" + port,
		Endpoint: &tb.WebhookEndpoint{PublicURL: publicURL},
	}

	B, err = tb.NewBot(tb.Settings{
		Token:  token,
		Poller: poller,
	})

	if err != nil {
		log.Fatalln(err)
	}

	B.Handle("/start", hStart)
	B.Handle(tb.OnText, hText)

	//ConnectDataBase()
	//defer DB.Close()

	B.Start()
}

func hStart(m *tb.Message) {
	if !m.Private() {
		return
	}

	B.Send(m.Sender, "")
}

func hText(m *tb.Message) {
	if !m.Private() {
		return
	}

	if m.Text == "🏁 Гонка" {
		message := `
		НАЧАЛО ГОНКИ
		
		1 🐌_______________🍭 Гери
		2 🐌_______________🍓 Боня
		3 🐌_______________🍏 Вася`

		B.Send(m.Sender, message)
	}
	if m.Text == "🐌 Улитки" {

	}
	if m.Text == "💰 Деньги" {

	}
	if m.Text == "❓ Помощь" {

	}
}
