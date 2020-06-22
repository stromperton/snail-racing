package main

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-pg/pg/v9"
	"golang.org/x/crypto/sha3"
	tb "gopkg.in/tucnak/telebot.v2"
)

const (
	maxScore        = 200
	winPos          = 26
	changeSpeedProb = 1
)

var (
	appWallet   string
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

	InlineBetNum = &tb.SendOptions{
		ParseMode: tb.ModeHTML,
		ReplyMarkup: &tb.ReplyMarkup{
			InlineKeyboard: [][]tb.InlineButton{
				{
					tb.InlineButton{Text: "10 BIP", Unique: "BetNum", Data: "10"},
					tb.InlineButton{Text: "25 BIP", Unique: "BetNum", Data: "25"},
				},
				{
					tb.InlineButton{Text: "50 BIP", Unique: "BetNum", Data: "50"},
					tb.InlineButton{Text: "100 BIP", Unique: "BetNum", Data: "100"},
				},
				{
					tb.InlineButton{Text: "🤑 Халява", Unique: "BetNum", Data: "haliava"},
				},
			},
		},
	}
	InlineBetNumHaliava = &tb.SendOptions{
		ParseMode: tb.ModeHTML,
		ReplyMarkup: &tb.ReplyMarkup{
			InlineKeyboard: [][]tb.InlineButton{
				{
					tb.InlineButton{Text: "1 🤑", Unique: "BetNum", Data: "1h"},
					tb.InlineButton{Text: "5 🤑", Unique: "BetNum", Data: "5h"},
				},
				{
					tb.InlineButton{Text: "10 🤑", Unique: "BetNum", Data: "10h"},
					tb.InlineButton{Text: "50 🤑", Unique: "BetNum", Data: "50h"},
				},
				{
					tb.InlineButton{Text: "🤑 Монеты", Unique: "BetNum", Data: "monety"},
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
		ParseMode:             tb.ModeHTML,
		DisableWebPagePreview: true,
		ReplyMarkup: &tb.ReplyMarkup{
			InlineKeyboard: [][]tb.InlineButton{
				{
					tb.InlineButton{Text: "🤑 Халява", Unique: "MoneyGive"},
				},
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
	B.Handle("/sender", hSender)
	B.Handle(tb.OnText, hText)
	B.Handle("\fGary", func(c *tb.Callback) { hSnails(c, "gary") })
	B.Handle("\fBonya", func(c *tb.Callback) { hSnails(c, "bonya") })
	B.Handle("\fVasya", func(c *tb.Callback) { hSnails(c, "vasya") })

	B.Handle("\fGaryBet", func(c *tb.Callback) { hBet(c, "gary") })
	B.Handle("\fBonyaBet", func(c *tb.Callback) { hBet(c, "bonya") })
	B.Handle("\fVasyaBet", func(c *tb.Callback) { hBet(c, "vasya") })

	B.Handle("\fBetNum", hBetNum)

	B.Handle("\fMoneyIn", hMoneyIn)
	B.Handle("\fMoneyOut", hMoneyOut)
	B.Handle("\fMoneyGive", hMoneyGive)

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
	if GetBalance(address) < 10.01 {
		B.Send(c.Sender, `🤯 Недостаточно средств для вывода!
<b>Минимальная сумма вывода:</b> 10 BIP`, tb.ModeHTML)
	} else {

		SetBotState(c.Sender.ID, "MinterAddressSend")
		B.Send(c.Sender, "💰 Куда будем отправлять монетки? <b>Пришли свой адрес в сети Minter</b>", &tb.SendOptions{ParseMode: tb.ModeHTML, ReplyMarkup: &tb.ReplyMarkup{ReplyKeyboardRemove: true}})
	}
}

func hMoneyGive(c *tb.Callback) {
	B.Respond(c)

	B.Send(c.Sender, `<b>Хочешь получить халяву?</b>
<i>Расскажи о боте друзьям!</i>
	
<b>За каждую ставку</b> твоих друзей ты получишь <b>один халявный заезд</b>
с возможностью выиграть <b>10 BIP</b>

<b>Например</b>
По ссылке Саши в бота зашли 10 человек. Предположим, что все они не сильно азартны и каждый сыграл всего по 3 раза.
Итого Саша получает <b>30 бесплатных</b> билетов. Если играть на них, то, согласно теории вероятности, сумма выигрышей составит примерно <code>30 * 1/3 * 10 = </code><b>100 BIP<b>
Таким образом Саша получает по <b>10 BIP с каждого</b> привлеченного игрока`, tb.ModeHTML)
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

func hSender(m *tb.Message) {
	if !m.Private() || m.Sender.ID != 303629013 || !m.IsReply() {
		return
	}

	var players []Player
	err := db.Model(&players).Select()

	if err != nil {
		fmt.Println(err)
	}

	for _, v := range players {
		B.Send(&tb.Chat{ID: int64(v.ID)}, m.ReplyTo.Text, tb.ModeHTML)
	}

}

/*
func hCheck(c *tb.Callback) {
	B.Respond(c)

	seed, _ := strconv.ParseInt(c.Data, 10, 64)
	myR := rand.New(rand.NewSource(seed))

	snails := [3]Snail{
		{Base: "_________________________🍭", Name: "gary"},
		{Base: "_________________________🍓", Name: "bonya"},
		{Base: "_________________________🍏", Name: "vasya"},
	}

	for i, _ := range snails {
		snails[i].Adka = Random(myR, 1, 10)
	}

	win := "nil"
	var winnersArray []string

	mess := fmt.Sprintf(messageRace, "tt", "ff",
		snails[0].GetString(),
		snails[1].GetString(),
		snails[2].GetString(),
	)
	B.Edit(c.Message, mess, tb.ModeHTML)
	for win == "nil" {

		isUpdateMessage := false
		for i := 0; i < 3; i++ {
			isUpdate, winner := snails[i].Hodik(myR)
			if isUpdate {
				isUpdateMessage = true
			}
			if winner {
				winnersArray = append(winnersArray, snails[i].Name)
			}
		}

		if len(winnersArray) > 0 {
			winInd := Random(myR, 0, len(winnersArray))

			for i, snailName := range winnersArray {
				if i == winInd {
					win = snailName
				} else {
					snails[i].Position--
				}
			}
		}

		if isUpdateMessage {
			message := fmt.Sprintf(messageRace, "tt", "fsdfdf",
				snails[0].GetString(),
				snails[1].GetString(),
				snails[2].GetString(),
			)
			B.Edit(c.Message, message, tb.ModeHTML)
		}
		time.Sleep(time.Millisecond * 10)

	}
	inlineCheck := &tb.SendOptions{
		ParseMode:             tb.ModeHTML,
		DisableWebPagePreview: true,
		ReplyMarkup: &tb.ReplyMarkup{
			InlineKeyboard: [][]tb.InlineButton{
				{
					tb.InlineButton{Text: "Проверка заезда", Data: strconv.FormatInt(seed, 10), Unique: "Check"},
				},
			},
		},
	}
	message := fmt.Sprintf(messageRace, "tt", "fsdfdf",
		snails[0].GetString(),
		snails[1].GetString(),
		snails[2].GetString(),
	)
	B.Edit(c.Message, message, inlineCheck)
}*/

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
			if err != nil || flyt < 10 {
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
<b>Минимальная сумма:</b> 10 BIP
<b>Максимальная сумма:</b> %.2f BIP`

			B.Send(m.Sender, fmt.Sprintf(message, bipBalance, minGasPriceF*0.01, max), ReplyOut)
		}

	} else {

		if m.Text == "🏁 Гонка" {
			if GetBotState(m.Sender.ID) == "race" {
				B.Send(m.Sender, "🤯 Гонка уже идёт!", tb.ModeHTML)
			} else {
				lastRaceMessage := GetLastRaceMessage(m.Sender.ID)
				if lastRaceMessage != nil {
					_, err := B.EditReplyMarkup(lastRaceMessage, nil)
					if err != nil {
						fmt.Println("Ну шо же", err)
					}
				}
				defPos := 0
				gary := Snail{Position: defPos, Base: "_________________________🍭"}
				bonya := Snail{Position: defPos, Base: "_________________________🍓"}
				vasya := Snail{Position: defPos, Base: "_________________________🍏"}

				message := fmt.Sprintf(GetText("race"), "🐌 Ожидание ставки...", "Кто победит? На кого будешь ставить?",
					gary.GetString(),
					bonya.GetString(),
					vasya.GetString(),
				)

				lastRaceMessage, _ = B.Send(m.Sender, message, InlineBet)
				SetLastRaceMessage(m.Sender.ID, lastRaceMessage)
			}
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

func hBetNum(c *tb.Callback) {
	B.Respond(c)
	var betNum float64
	var betNumHaliava int
	var betka string

	var supers string
	var jjj string

	betSnailName := GetBetSnailName(c.Sender.ID)

	if c.Data[0] == ' ' {
		jjj = c.Data[7:]
	} else {
		jjj = c.Data
	}
	switch jjj {
	case "10":
		betNum = 10
	case "25":
		betNum = 25
	case "50":
		betNum = 50
	case "100":
		betNum = 100
	case "haliava":
		B.EditReplyMarkup(c.Message, InlineBetNumHaliava.ReplyMarkup)
		return
	case "monety":
		B.EditReplyMarkup(c.Message, InlineBetNum.ReplyMarkup)
		return
	case "1h":
		betNumHaliava = 1
	case "5h":
		betNumHaliava = 5
	case "10h":
		betNumHaliava = 10
	case "50h":
		betNumHaliava = 50
	default:
		B.Send(c.Sender, "🤯 Что-то пошло не так "+c.Data+" "+jjj, tb.ModeHTML)
		return
	}

	if betNum > 0 {
		address, key := GetWallet(c.Sender.ID)
		result, err := SendCoin(betNum-float64(0.01), address, appWallet, key)
		if err != nil {
			fmt.Println("Ошибка отправки транзакции", err)
			B.Send(c.Sender, "🤯 Не хватает средств? Загляни в раздел <b>💰 Кошелёк</b>", tb.ModeHTML)
			return
		}
		supers = result.Hash

	} else if betNumHaliava > 0 {
		if GetHaliava(c.Sender.ID) >= betNumHaliava {
			haliavaChange(c.Sender.ID, -1*betNumHaliava)

			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			supers = "x0haliava" + strconv.Itoa(Random(r, 1, 100000000000)) + "end"
		} else {
			B.Send(c.Sender, "🤯 Нужна халява? Загляни в раздел <b>💰 Кошелёк</b>", tb.ModeHTML)
			return
		}
	}

	hash := strings.ToLower(supers)

	SetBotState(c.Sender.ID, "race")
	fmt.Println("Ставка "+c.Data+" BIP ", hash)

	h := sha3.NewLegacyKeccak256()
	seed := int64(binary.BigEndian.Uint64(h.Sum([]byte(hash))))
	myR := rand.New(rand.NewSource(seed))

	snails := [3]Snail{
		{Adka: Random(myR, 1, 10), Base: "_________________________🍭", Name: "gary"},
		{Adka: Random(myR, 1, 10), Base: "_________________________🍓", Name: "bonya"},
		{Adka: Random(myR, 1, 10), Base: "_________________________🍏", Name: "vasya"},
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
			isUpdate, winner := snails[i].Hodik(myR)
			if isUpdate {
				isUpdateMessage = true
			}
			if winner {
				winnersArray = append(winnersArray, snails[i].Name)
			}
		}

		if len(winnersArray) > 0 {
			winInd := Random(myR, 0, len(winnersArray))

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
	inlineCheck := &tb.SendOptions{
		ParseMode:             tb.ModeHTML,
		DisableWebPagePreview: true,
		ReplyMarkup: &tb.ReplyMarkup{
			InlineKeyboard: [][]tb.InlineButton{
				{
					tb.InlineButton{Text: "Проверка заезда", URL: "https://play.golang.org/p/Sc5JMnGFnIv"},
				},
			},
		},
	}

	B.Send(c.Sender, "<code>Mt"+hash+"</code>", inlineCheck)

	if win == betSnailName {
		var title string
		var betNumWin float64
		if betNum > 0 {
			betNumWin = betNum
		} else if betNumHaliava > 0 {
			betNumWin = float64(betNumHaliava) * 5
		}

		address, _ := GetWallet(c.Sender.ID)
		result, err := SendCoin(betNumWin*2, appWallet, address, GetPrivateKeyFromMnemonic(os.Getenv("MNEMONIC")))

		if err != nil {
			fmt.Println("Ошибка отправки транзакции", err)
			B.Send(c.Sender, "🤯 ЭТОГО НЕ ДОЛЖНО БЫЛО СЛУЧИТСЯ! ВЫИГРЫШ НЕ ОТПРАВИЛСЯ!!!", ReplyMain)
		}
		fmt.Println(result)

		title = fmt.Sprintf("Твоя улитка победила! Выигрыш - %.0f BIP!", betNumWin*2)

		message := fmt.Sprintf(messageRace, title,
			betka,
			snails[0].GetString(),
			snails[1].GetString(),
			snails[2].GetString(),
		)
		_, err = B.Edit(c.Message, message, tb.ModeHTML)
		fmt.Println(err)
		B.Send(c.Sender, "<b>🎉 Твоя ставка зашла!</b> <i>Не забудь поделиться с друзьями!</i>", tb.ModeHTML)

		doWin(c.Sender.ID)
	} else {
		doLose(c.Sender.ID)

		title := "К сожалению, твоя улитка проиграла..."

		message := fmt.Sprintf(messageRace, title,
			betka,
			snails[0].GetString(),
			snails[1].GetString(),
			snails[2].GetString(),
		)
		_, err := B.Edit(c.Message, message, tb.ModeHTML)
		fmt.Println(err)
		B.Send(c.Sender, "Эхх, неудача! <b>Попробуй ещё раз!</b>", tb.ModeHTML)
	}
	//B.Send(c.Sender, "Ты всегда можешь <a href='https://play.golang.org/p/2uElqjxMZca'>проверить бота на честность</a>, используя транзакцию заезда:", tb.ModeHTML)

	SetBotState(c.Sender.ID, "default")
}

func hBet(c *tb.Callback, betSnailName string) {
	B.Respond(c)
	var betka string

	SetBetSnailName(c.Sender.ID, betSnailName)

	snails := [3]Snail{
		{Base: "_________________________🍭", Name: "gary"},
		{Base: "_________________________🍓", Name: "bonya"},
		{Base: "_________________________🍏", Name: "vasya"},
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

	address, _ := GetWallet(c.Sender.ID)
	bipBalance := GetBalance(address)

	message := fmt.Sprintf(messageRace, "💰 Ожидание ставки...", fmt.Sprintf(`
	Баланс: <b>%.2f BIP</b>
	`+betka+`
	Выигрыш = <b>Размер ставки × 2</b>`, bipBalance),
		snails[0].GetString(),
		snails[1].GetString(),
		snails[2].GetString(),
	)
	B.Edit(c.Message, message, InlineBetNum)
}

func hSnails(c *tb.Callback, snailName string) {
	B.Respond(c)
	B.Edit(c.Message, GetText(snailName), InlineSnails)
}

type Player struct {
	ID              int
	Address         string
	PrivateKey      string
	WinCount        int `pg:"win_count,use_zero,notnull"`
	LoseCount       int `pg:"lose_count,use_zero,notnull"`
	BotState        string
	OutAddress      string
	BetSnailName    string
	LastRaceMessage *tb.Message

	Haliava int `pg:"haliava,use_zero,notnull"`
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

func haliavaChange(id int, change int) {
	p := &Player{}
	p.ID = id
	p.Haliava = GetHaliava(id)
	p.Haliava += change

	db.Model(p).Set("haliava = ?", p.Haliava).Where("id = ?", p.ID).Update()
}

func GetHaliava(id int) int {
	p := &Player{}
	p.ID = id
	err := db.Select(p)
	if err != nil {
		fmt.Println(err)
	}

	return p.Haliava
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

func GetBetSnailName(id int) string {
	p := &Player{}
	p.ID = id
	err := db.Select(p)
	if err != nil {
		fmt.Println(err)
	}

	return p.BetSnailName
}

func GetLastRaceMessage(id int) *tb.Message {
	p := &Player{}
	p.ID = id
	err := db.Select(p)
	if err != nil {
		fmt.Println(err)
	}

	return p.LastRaceMessage
}

func SetLastRaceMessage(id int, message *tb.Message) {
	p := &Player{}
	p.ID = id
	p.LastRaceMessage = message

	db.Model(p).Set("last_race_message = ?", p.LastRaceMessage).Where("id = ?", p.ID).Update()
}

func SetBetSnailName(id int, bsn string) {
	p := &Player{}
	p.ID = id
	p.BetSnailName = bsn

	db.Model(p).Set("bet_snail_name = ?", p.BetSnailName).Where("id = ?", p.ID).Update()
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
