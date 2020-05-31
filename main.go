package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"strconv"
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
					tb.ReplyButton{Text: "💰 Кошелёк"},
					tb.ReplyButton{Text: "❓ Помощь"},
				},
			},
		},
	}

	ReplyOut = &tb.SendOptions{
		ParseMode: tb.ModeHTML,
		ReplyMarkup: &tb.ReplyMarkup{
			ResizeReplyKeyboard: true,
			ReplyKeyboard: [][]tb.ReplyButton{
				{
					tb.ReplyButton{Text: "❌ Отмена"},
					tb.ReplyButton{Text: "💰 Отправить всё"},
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

	InlineMoney = &tb.SendOptions{
		ParseMode: tb.ModeHTML,
		ReplyMarkup: &tb.ReplyMarkup{
			InlineKeyboard: [][]tb.InlineButton{
				{
					tb.InlineButton{Text: "📥 Пополнить", Unique: "MoneyIn"},
					tb.InlineButton{Text: "📤 Вывести", Unique: "MoneyOut"},
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

	B.Handle("\fMoneyIn", hMoneyIn)
	B.Handle("\fMoneyOut", hMoneyOut)

	ConnectDataBase()
	defer db.Close()

	B.Start()
}

func hMoneyIn(c *tb.Callback) {
	B.Respond(c)
	address, _ := GetWallet(c.Sender.ID)

	B.Send(c.Sender, "💰 Чтобы пополнить баланс, отправь BIP на этот адрес:")
	B.Send(c.Sender, "<code>"+address+"</code>", ReplyMain)
}
func hMoneyOut(c *tb.Callback) {
	B.Respond(c)

	address, _ := GetWallet(c.Sender.ID)
	if GetBalance(address) < 40.01 {
		B.Send(c.Sender, `🤯 Недостаточно средств для вывода!
<b>Минимальная сумма вывода:</b> 40 BIP`)
	} else {

		SetBotState(c.Sender.ID, "MinterAddressSend")
		B.Send(c.Sender, "💰 Куда будем отправлять монетки? <b>Пришли свой адрес в сети Minter</b>", &tb.SendOptions{ParseMode: tb.ModeHTML, ReplyMarkup: &tb.ReplyMarkup{ReplyKeyboardRemove: true}})
	}
}

func hStart(m *tb.Message) {
	if !m.Private() {
		return
	}
	p, isNewPlayer := NewDefaultPlayer(m.Sender.ID)

	if isNewPlayer {
		fmt.Printf("Новый игрок: @%s[%d]\n", m.Sender.Username, p.ID)

		message := GetText("start")
		B.Send(m.Sender, message, ReplyMain)
	} else {
		B.Send(m.Sender, "🤯 Похоже, что ты уже играешь!", ReplyMain)
	}
}

func hText(m *tb.Message) {
	if !m.Private() {
		return
	}

	botState := GetBotState(m.Sender.ID)

	if botState == "CoinNumSend" {
		if m.Text == "💰 Отправить всё" {
			adress, prKey := GetWallet(m.Sender.ID)
			outAdress := GetOutAddress(m.Sender.ID)
			minGasPrice, _ := minterClient.MinGasPrice()
			minGasPriceF, _ := strconv.ParseFloat(minGasPrice, 64)
			num := GetBalance(adress) - minGasPriceF*0.01
			_, err := SendCoin(num, adress, outAdress, prKey)

			if err != nil {
				B.Send(m.Sender, "🤯 Ошибка транзакции.", ReplyMain)
			} else {
				B.Send(m.Sender, "🎉 Монеты успешно отправлены!", ReplyMain)
			}
		} else if m.Text == "❌ Отмена" {
			B.Send(m.Sender, "❌ Вывод прерван", ReplyMain)
		} else {
			flyt, err := strconv.ParseFloat(m.Text, 64)
			if err != nil || flyt < 40 {
				B.Send(m.Sender, "🤯 Что-то не так... Нужно просто отправить число монет", ReplyMain)
			} else {
				adress, prKey := GetWallet(m.Sender.ID)
				outAdress := GetOutAddress(m.Sender.ID)
				_, err := SendCoin(flyt, adress, outAdress, prKey)
				if err != nil {
					B.Send(m.Sender, "🤯 Ошибка транзакции", ReplyMain)
				} else {
					B.Send(m.Sender, "🎉 Монеты успешно отправлены!", ReplyMain)
				}
			}
		}
		SetBotState(m.Sender.ID, "default")
	} else if botState == "MinterAddressSend" {

		_, err := minterClient.Address(m.Text)

		if err != nil {
			SetBotState(m.Sender.ID, "default")
			B.Send(m.Sender, "🤯 С этим адресом что-то не так. Перепроверь и попробуй ещё раз", ReplyMain)
		} else {
			address, _ := GetWallet(m.Sender.ID)
			bipBalance := GetBalance(address)

			minterClient.MaxGas()
			minGasPrice, _ := minterClient.MinGasPrice()
			minGasPriceF, _ := strconv.ParseFloat(minGasPrice, 64)

			max := bipBalance - minGasPriceF*0.01

			SetOutAddress(m.Sender.ID, m.Text)
			SetBotState(m.Sender.ID, "CoinNumSend")
			message := `💰 Сколько ты хочешь вывести? <b>Введи количество монет BIP</b>

<b>Доступно:</b> %.2f BIP
<b>Коммиссия на вывод средств:</b> %.2f BIP
<b>Минимальная сумма:</b> 40 BIP
<b>Максимальная сумма:</b> %.2f BIP`

			B.Send(m.Sender, fmt.Sprintf(message, bipBalance, minGasPriceF*0.01, max), ReplyOut)
		}

	} else {

		if m.Text == "🏁 Гонка" {
			defPos := 0
			gary := Snail{Position: defPos, Base: "_________________________🍭"}
			bonya := Snail{Position: defPos, Base: "_________________________🍓"}
			vasya := Snail{Position: defPos, Base: "_________________________🍏"}

			address, _ := GetWallet(m.Sender.ID)
			bipBalance := GetBalance(address)
			minGasPrice, _ := minterClient.MinGasPrice()
			minGasPriceF, _ := strconv.ParseFloat(minGasPrice, 64)

			message := fmt.Sprintf(GetText("race"), "💰 Ожидание ставки...", fmt.Sprintf(`
Баланс: <b>%.2f BIP</b>
Размер ставки - <b>50 BIP</b> + Комиссия - %.2f
<b>Выигрыш - 100 BIP</b>`, bipBalance, minGasPriceF*0.01),
				gary.GetString(),
				bonya.GetString(),
				vasya.GetString(),
			)

			B.Send(m.Sender, message, InlineBet)
		} else if m.Text == "🐌 Улитки" {
			message := fmt.Sprintf(GetText("gary"))

			B.Send(m.Sender, message, InlineSnails)
		} else if m.Text == "💰 Кошелёк" {
			winC, _ := GetRate(m.Sender.ID)

			address, _ := GetWallet(m.Sender.ID)
			bipBalance := GetBalance(address)
			usdBalance := GetBipPrice() * bipBalance

			message := fmt.Sprintf(GetText("winrate"), math.Round(bipBalance*100)/100, math.Round(usdBalance*100)/100, winC)

			B.Send(m.Sender, message, InlineMoney)
		} else if m.Text == "❓ Помощь" {
			message := GetText("help")
			B.Send(m.Sender, message, &tb.SendOptions{DisableWebPagePreview: true, ParseMode: tb.ModeHTML})
		} else {
			B.Send(m.Sender, "🤯 Жми на кнопки в меню!", ReplyMain)
		}
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
		B.Send(c.Sender, "🤯 Не хватает средств? Загляни в раздел <b>💰 Кошелёк</b>", tb.ModeHTML)
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
			B.Send(c.Sender, "🤯 ЭТОГО НЕ ДОЛЖНО БЫЛО СЛУЧИТСЯ! ВЫИГРЫШЬ НЕ ОТПРАВИЛСЯ!!!", ReplyMain)
		}
		fmt.Println(result)
		title := "Твоя улитка победила! Выигрышь - 100 BIP"

		message := fmt.Sprintf(messageRace, title,
			betka,
			snails[0].GetString(),
			snails[1].GetString(),
			snails[2].GetString(),
		)
		B.Edit(c.Message, message, tb.ModeHTML)
		B.Send(c.Sender, "Твоя ставка зашла! Не забудь поделиться с друзьями!", tb.ModeHTML)

		doWin(c.Sender.ID)
	} else {
		doLose(c.Sender.ID)

		title := "Эхх, неудача! Попробуй ещё раз!"

		message := fmt.Sprintf(messageRace, title,
			betka,
			snails[0].GetString(),
			snails[1].GetString(),
			snails[2].GetString(),
		)
		B.Edit(c.Message, message, tb.ModeHTML)
		B.Send(c.Sender, "Ты можешь почитать про 🐌 Улиток в особом разделе", tb.ModeHTML)
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
	BotState   string
	OutAddress string
}

func NewDefaultPlayer(id int) (Player, bool) {
	p := &Player{}
	p.ID = id
	p.Address, p.PrivateKey = CreateWallet()
	p.BotState = "default"

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
	p.WinCount, _ = GetRate(id)
	p.WinCount++

	db.Model(p).Set("win_count = ?", p.WinCount).Where("id = ?", p.ID).Update()
}
func doLose(id int) {
	p := &Player{}
	p.ID = id
	_, p.LoseCount = GetRate(id)
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

func GetOutAddress(id int) string {
	p := &Player{}
	p.ID = id
	err := db.Select(p)
	if err != nil {
		fmt.Println(err)
	}

	return p.OutAddress
}

func SetOutAddress(id int, outA string) {
	p := &Player{}
	p.ID = id
	p.OutAddress = outA

	db.Model(p).Set("out_address = ?", p.OutAddress).Where("id = ?", p.ID).Update()
}

func GetBotState(id int) string {
	p := &Player{}
	p.ID = id
	err := db.Select(p)
	if err != nil {
		fmt.Println(err)
	}

	return p.BotState
}

func SetBotState(id int, state string) {
	p := &Player{}
	p.ID = id
	p.BotState = state

	db.Model(p).Set("bot_state = ?", p.BotState).Where("id = ?", p.ID).Update()
}
