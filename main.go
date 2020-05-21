package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	tb "gopkg.in/tucnak/telebot.v2"
)

type Snail struct {
	Position int
	Candy    string
}

func (s *Snail) GetString() string {
	base := "_______________" + s.Candy
	out := base[:s.Position] + "🐌" + base[s.Position:]

	return out
}

var B *tb.Bot

var (
	ReplyMain = &tb.SendOptions{
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

	InlineBet = &tb.SendOptions{
		ParseMode: tb.ModeHTML,
		ReplyMarkup: &tb.ReplyMarkup{
			InlineKeyboard: [][]tb.InlineButton{
				{
					tb.InlineButton{Text: "Ставка 🍭 на Гери(🐌 №1)", Unique: "GeryBet"},
					tb.InlineButton{Text: "Ставка 🍓 на Боню(🐌 №2)", Unique: "BonyaBet"},
					tb.InlineButton{Text: "Ставка 🍏 на Васю(🐌 №3)", Unique: "VasyaBet"},
				},
			},
		},
	}

	InlineSnails = &tb.SendOptions{
		ParseMode: tb.ModeHTML,
		ReplyMarkup: &tb.ReplyMarkup{
			InlineKeyboard: [][]tb.InlineButton{
				{
					tb.InlineButton{Text: "🐌 Гери", Unique: "Gery"},
					tb.InlineButton{Text: "🐌 Боня", Unique: "Bonya"},
					tb.InlineButton{Text: "🐌 Вася", Unique: "Vasya"},
				},
			},
		},
	}
)

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
	B.Handle("\fGary", func(c *tb.Callback) { hSnails(c, "gary") })
	B.Handle("\fBonya", func(c *tb.Callback) { hSnails(c, "bonya") })
	B.Handle("\fVasya", func(c *tb.Callback) { hSnails(c, "vasya") })

	B.Handle("\fGaryBet", func(c *tb.Callback) { hBet(c, "gary") })
	B.Handle("\fBonyaBet", func(c *tb.Callback) { hBet(c, "bonya") })
	B.Handle("\fVasyaBet", func(c *tb.Callback) { hBet(c, "vasya") })

	//ConnectDataBase()
	//defer DB.Close()

	B.Start()
}

func hStart(m *tb.Message) {
	if !m.Private() {
		return
	}

	B.Send(m.Sender, "Стартовое сообщение", ReplyMain)
}

func hText(m *tb.Message) {
	if !m.Private() {
		return
	}

	if m.Text == "🏁 Гонка" {
		defPos := 0
		gery := Snail{Position: defPos, Candy: "🍭"}
		bonya := Snail{Position: defPos, Candy: "🍓"}
		vasya := Snail{Position: defPos, Candy: "🍏"}

		message := fmt.Sprintf(GetText("race"), "Ожидание ставки...",
			"Размер ставки - <b>10 BIP</b><br><b>Выигрыш - 20 BIP</b>",
			gery.GetString(),
			bonya.GetString(),
			vasya.GetString(),
		)

		B.Send(m.Sender, message, InlineBet)
	}
	if m.Text == "🐌 Улитки" {
		message := fmt.Sprintf(GetText("gary"))

		B.Send(m.Sender, message, InlineSnails)
	}
	if m.Text == "💰 Деньги" {

	}
	if m.Text == "❓ Помощь" {

	}
}

//GetText Получить сообщение из файла
func GetText(fileName string) string {
	content, err := ioutil.ReadFile("messages/" + fileName + ".html")
	if err != nil {
		fmt.Println("Проблема с вытаскиванием сообщения из файла", fileName, err)
	}

	return string(content)
}

func hBet(c *tb.Callback, snailName string) {
	B.Respond(c)
	B.Edit(c.Message, "Ставка принята", InlineBet)
}

func hSnails(c *tb.Callback, snailName string) {
	B.Respond(c)
	B.Edit(c.Message, GetText(snailName), InlineSnails)
}
