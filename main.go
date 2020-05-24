package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/go-pg/pg/v9"
	tb "gopkg.in/tucnak/telebot.v2"
)

type Snail struct {
	Position int
	Speed    int
	Score    int
	Adka     int
	Candy    string
}

func (s *Snail) GetString() string {
	base := "_________________________" + s.Candy
	out := base[:s.Position] + "🐌" + base[s.Position:]

	return out
}

var B *tb.Bot
var db *pg.DB
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
					tb.InlineButton{Text: "Ставка 🍭 на Гери(🐌 №1)", Unique: "GaryBet"},
				},
				{
					tb.InlineButton{Text: "Ставка 🍓 на Боню(🐌 №2)", Unique: "BonyaBet"},
				},
				{
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
					tb.InlineButton{Text: "🐌 Гери", Unique: "Gary"},
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

	ConnectDataBase()
	defer db.Close()

	B.Start()
}

func hStart(m *tb.Message) {
	if !m.Private() {
		return
	}
	p, isNewPlayer := NewDefaultPlayer(m.Sender.ID)

	if isNewPlayer {
		fmt.Printf("Новый игрок: @%s[%d]\n", m.Sender.Username, p.ID)

		B.Send(m.Sender, "Стартовое сообщение", ReplyMain)
	} else {
		B.Send(m.Sender, "Похоже, что ты уже играешь!", ReplyMain)
	}
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
			`Размер ставки - <b>50 BIP</b>
<b>Выигрыш - 100 BIP</b>`,
			gery.GetString(),
			bonya.GetString(),
			vasya.GetString(),
		)
		fmt.Println(message)
		B.Send(m.Sender, message, InlineBet)
	}
	if m.Text == "🐌 Улитки" {
		message := fmt.Sprintf(GetText("gary"))

		B.Send(m.Sender, message, InlineSnails)
	}
	if m.Text == "💰 Деньги" {

		B.Send(m.Sender, "💰 Деньги", ReplyMain)
	}
	if m.Text == "❓ Помощь" {

		B.Send(m.Sender, "❓ Помощь", ReplyMain)
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

	gery := Snail{Adka: Random(1, 4), Candy: "🍭"}
	bonya := Snail{Adka: Random(1, 4), Candy: "🍓"}
	vasya := Snail{Adka: Random(1, 4), Candy: "🍏"}

	gery.Speed = Random(60, 200)
	bonya.Speed = Random(60, 200)
	vasya.Speed = Random(60, 200)

	fmt.Println("Скорость Гери:", gery.Speed)
	fmt.Println("Скорость Бони:", bonya.Speed)
	fmt.Println("Скорость Васи:", vasya.Speed)

	win := false
	for !win {
		rSnail := Random(0, 3)
		var luckySnail *Snail
		if rSnail == 0 {
			luckySnail = &gery
		}
		if rSnail == 1 {
			luckySnail = &bonya
		}
		if rSnail == 2 {
			luckySnail = &vasya
		}
		randomka := Random(0, 100)

		if randomka < 20 {
			luckySnail.Adka = Random(1, 4)
		}

		gery.Score += gery.Adka
		bonya.Score += bonya.Adka
		vasya.Score += vasya.Adka

		fmt.Println("Скоры Гери:", gery.Score)
		fmt.Println("Скоры Бони:", bonya.Score)
		fmt.Println("Скоры Васи:", vasya.Score)

		isUpdateMessage := false
		if gery.Score > gery.Speed {
			gery.Position++
			gery.Score = 0
			isUpdateMessage = true
		}
		if bonya.Score > bonya.Speed {
			bonya.Position++
			bonya.Score = 0
			isUpdateMessage = true
		}
		if vasya.Score > vasya.Speed {
			vasya.Position++
			vasya.Score = 0
			isUpdateMessage = true
		}

		if gery.Position == 26 || bonya.Position == 26 || vasya.Position == 26 {
			win = true
		}

		if isUpdateMessage {

			message := fmt.Sprintf(GetText("race"), "ГОНКА",
				"",
				gery.GetString(),
				bonya.GetString(),
				vasya.GetString(),
			)

			fmt.Println("Позиция Гери:", gery.Position)
			fmt.Println("Позиция Бони:", bonya.Position)
			fmt.Println("Позиция Васи:", vasya.Position)

			B.Edit(c.Message, message, InlineBet)
		}
		time.Sleep(time.Millisecond * 10)
	}
}

func hSnails(c *tb.Callback, snailName string) {
	B.Respond(c)
	B.Edit(c.Message, GetText(snailName), InlineSnails)
}

type Player struct {
	ID         int
	Address    string
	PrivateKey string
}

func NewDefaultPlayer(id int) (Player, bool) {
	p := &Player{}
	p.ID = id
	p.Address, p.PrivateKey = CreateWallet()

	res, err := db.Model(p).OnConflict("DO NOTHING").Insert()
	if err != nil {
		panic(err)
	}

	if res.RowsAffected() > 0 {
		return *p, true
	}
	return *p, false
}
