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
	Name     string
	Position int
	Speed    int
	Score    int
	Adka     int
	Base     string
}

func (s *Snail) GetString() string {
	var out string
	if s.Position == winPos {
		out = "_________________________🐌🥇"
	} else {
		out = s.Base[:s.Position] + "🐌" + s.Base[s.Position:]

	}
	return out
}

func (s *Snail) Hodik() (bool, bool) {
	randomka := Random(0, 100)

	if randomka < changeSpeedProb {
		s.Adka = Random(1, 10)
	}

	s.Score += s.Adka
	//fmt.Println("Скоры "+s.Name+":", gary.Score)
	if s.Score > maxScore {
		s.Position++
		s.Score = 0

		if s.Position == winPos {
			return true, true
		}

		return true, false
	}
	return false, false
}

var (
	maxScore        int
	winPos          int
	changeSpeedProb int

	appWallet string

	messageRace string
)

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
					tb.InlineButton{Text: "🍭 На Гери(🐌 №1)", Unique: "GaryBet"},
				},
				{
					tb.InlineButton{Text: "🍓 На Боню(🐌 №2)", Unique: "BonyaBet"},
				},
				{
					tb.InlineButton{Text: "🍏 На Васю(🐌 №3)", Unique: "VasyaBet"},
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

	maxScore = GetInt("MAX_SCORE")
	winPos = GetInt("WIN_POS")
	changeSpeedProb = GetInt("CHANGE_SPEED_PROB")

	appWallet = os.Getenv("APP_WALLET")

	messageRace = GetText("race")

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
		gary := Snail{Position: defPos, Base: "_________________________🍭"}
		bonya := Snail{Position: defPos, Base: "_________________________🍓"}
		vasya := Snail{Position: defPos, Base: "_________________________🍏"}

		message := fmt.Sprintf(GetText("race"), "Ожидание ставки...",
			`Размер ставки - <b>50 BIP</b>
<b>Выигрыш - 100 BIP</b>`,
			gary.GetString(),
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
		winC, _ := GetRate(m.Sender.ID)

		address, _ := GetWallet(m.Sender.ID)
		bipBalance := GetBalance(address)
		usdBalance := GetBipPrice() * bipBalance

		message := fmt.Sprintf(GetText("winrate"), address, bipBalance, usdBalance, winC, "0", 0)

		B.Send(m.Sender, message, ReplyMain)
	}
	if m.Text == "❓ Помощь" {
		message := "Помощь..."
		B.Send(m.Sender, message, ReplyMain)
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

func hBet(c *tb.Callback, betSnailName string) {
	B.Respond(c)
	var betka string

	address, key := GetWallet(c.Sender.ID)
	result, err := SendCoin(50, address, appWallet, key)
	if err != nil {
		fmt.Println("Ошибка отправки транзакции", err)
		return
	}

	fmt.Println(result)

	snails := [3]Snail{
		{Adka: Random(1, 10), Base: "_________________________🍭", Name: "gary"},
		{Adka: Random(1, 10), Base: "_________________________🍓", Name: "bonya"},
		{Adka: Random(1, 10), Base: "_________________________🍏", Name: "vasya"},
	}

	if betSnailName == snails[0].Name {
		betka = "Ставка: 🐌 <b>Гери</b> 🍭"
	}
	if betSnailName == snails[1].Name {
		betka = "Ставка: 🐌 <b>Боня</b> 🍓"
	}
	if betSnailName == snails[2].Name {
		betka = "Ставка: 🐌 <b>Вася</b> 🍏"
	}

	win := "nil"
	var winnersArray []string
	for win == "nil" {

		isUpdateMessage := false
		for i := 0; i < 3; i++ {
			isUpdate, winner := snails[i].Hodik()
			if isUpdate {
				isUpdateMessage = true
			}
			if winner {
				winnersArray = append(winnersArray, snails[i].Name)
			}
		}

		if len(winnersArray) > 0 {
			winInd := Random(0, len(winnersArray))

			for i, snailName := range winnersArray {
				if i == winInd {
					win = snailName
				} else {
					snails[i].Position--
				}
			}
		}

		if isUpdateMessage {
			title := "Гонка!"

			message := fmt.Sprintf(messageRace, title,
				betka,
				snails[0].GetString(),
				snails[1].GetString(),
				snails[2].GetString(),
			)
			B.Edit(c.Message, message, tb.ModeHTML)
		}
		time.Sleep(time.Millisecond * 10)
	}
	if win == betSnailName {
		address, _ := GetWallet(c.Sender.ID)
		result, err := SendCoin(100, appWallet, address, GetPrivateKeyFromMnemonic(os.Getenv("MNEMONIC")))
		if err != nil {
			fmt.Println("Ошибка отправки транзакции", err)
		}
		fmt.Println(result)

		doWin(c.Sender.ID)
	} else {
		doLose(c.Sender.ID)
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
	WinCount   int `pg:"win_count,use_zero,notnull"`
	LoseCount  int `pg:"lose_count,use_zero,notnull"`
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

func doWin(id int) {
	p := &Player{}
	p.ID = id
	p.WinCount++

	db.Model(p).Set("win_count = ?", p.WinCount).Where("id = ?", p.ID).Update()
}
func doLose(id int) {
	p := &Player{}
	p.ID = id
	p.LoseCount++

	db.Model(p).Set("lose_count = ?", p.LoseCount).Where("id = ?", p.ID).Update()
}

func GetRate(id int) (int, int) {
	p := &Player{}
	p.ID = id
	err := db.Select(p)
	if err != nil {
		fmt.Println(err)
	}

	return p.WinCount, p.LoseCount
}

func GetWallet(id int) (string, string) {
	p := &Player{}
	p.ID = id
	err := db.Select(p)
	if err != nil {
		fmt.Println(err)
	}

	return p.Address, p.PrivateKey
}
