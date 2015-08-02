package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"path"
	"runtime"
	"time"

	"github.com/jh-bate/fantail"
	"github.com/tucnak/telebot"
)

type config struct {
	BotToken string `json:"botToken"`
}

type fantailBot struct {
	bot     *telebot.Bot
	current *telebot.Message
	api     *fantail.Api
}

const (
	bg       = "/bg"
	mood     = "/mood"
	note     = "/note"
	exercise = "/exercise"
	chat     = "/chat"
)

type step struct {
	done bool
	fn   func(msg telebot.Message)
}

type walk []step

type steps struct {
	Name    string
	UserId  int
	ToWalk  walk
	Walking bool
}

//the various steps we can `walk`

func (b fantailBot) initBgSteps(userId int) *steps {
	return &steps{
		Name:    bg,
		Walking: false,
		UserId:  userId,
		ToWalk: walk{
			{false, b.greet},
			{false, b.bg},
			{false, b.bgFeedback},
			{false, b.note},
			{false, b.thanks},
		},
	}
}

func (b fantailBot) initNoteSteps(userId int) *steps {
	return &steps{
		Name:    note,
		Walking: false,
		UserId:  userId,
		ToWalk: walk{
			{false, b.greet},
			{false, b.note},
			{false, b.thanks},
		},
	}
}

func (b fantailBot) initMoodSteps(userId int) *steps {
	return &steps{
		Name:    mood,
		Walking: false,
		UserId:  userId,
		ToWalk: walk{
			{false, b.greet},
			{false, b.mood},
			{false, b.note},
			{false, b.thanks},
		},
	}
}

func (b fantailBot) initExerciseSteps(userId int) *steps {
	return &steps{
		Name:    exercise,
		Walking: false,
		UserId:  userId,
		ToWalk: walk{
			{false, b.greet},
			{false, b.exercise},
			{false, b.note},
			{false, b.thanks},
		},
	}
}

func (b fantailBot) initChatSteps(userId int) *steps {
	return &steps{
		Name:    chat,
		Walking: false,
		UserId:  userId,
		ToWalk: walk{
			{false, b.greet},
			{false, b.mood},
			{false, b.exercise},
			{false, b.bg},
			{false, b.bgFeedback},
			{false, b.note},
			{false, b.thanks},
		},
	}
}

var fBot *fantailBot

func loadConfig() *config {

	_, filename, _, _ := runtime.Caller(1)
	configFile, err := ioutil.ReadFile(path.Join(path.Dir(filename), "botConfig.json"))

	if err != nil {
		log.Panic("could not load config ", err.Error())
	}
	var botConf config
	err = json.Unmarshal(configFile, &botConf)
	if err != nil {
		log.Panic("could not load config")
	}
	return &botConf
}

func initBot() *fantailBot {
	botConfig := loadConfig()

	bot, err := telebot.NewBot(botConfig.BotToken)
	if err != nil {
		return nil
	}

	return &fantailBot{bot: bot, api: fantail.InitApi()}
}

func (b fantailBot) mood(msg telebot.Message) {

	fBot.bot.SendMessage(msg.Chat, "How would you say you are feeling at the moment?",
		&telebot.SendOptions{
			ReplyMarkup: telebot.ReplyMarkup{
				ForceReply:      true,
				CustomKeyboard:  [][]string{[]string{":)", ":|"}, []string{":(", ">:("}},
				OneTimeKeyboard: true,
			},
		})
}

func (b fantailBot) bg(msg telebot.Message) {

	fBot.bot.SendMessage(msg.Chat, "Its all about that BG ... well not totally but", nil)

	fBot.bot.SendMessage(msg.Chat, "Where did your last BG fall?", &telebot.SendOptions{
		ReplyMarkup: telebot.ReplyMarkup{
			ForceReply:      true,
			CustomKeyboard:  [][]string{[]string{"above"}, []string{"in range"}, []string{"below"}},
			OneTimeKeyboard: true,
		},
	})

	return
}

func (b fantailBot) exercise(msg telebot.Message) {

	fBot.bot.SendMessage(msg.Chat, "So have you managed to get in any excerise recently?", &telebot.SendOptions{
		ReplyMarkup: telebot.ReplyMarkup{
			ForceReply:      true,
			CustomKeyboard:  [][]string{[]string{"yeah"}, []string{"nah"}},
			OneTimeKeyboard: true,
		},
	})

	return
}
func (b fantailBot) bgFeedback(msg telebot.Message) {

	feedbackText := ""

	if msg.Text == "above" {
		feedbackText = "just remember everyone has a high BG at times - you just don't want that to be the norm :)"
	}

	if msg.Text == "in range" {
		feedbackText = "** high five ** thats how we roll!!"
	}

	if msg.Text == "below" {
		feedbackText = "lows - they can really suck! remember to treat them properly so you aren't riding the roller coaster all day!"
	}

	fBot.bot.SendMessage(msg.Chat, feedbackText, nil)
	return
}

func (b fantailBot) note(msg telebot.Message) {

	fBot.bot.SendMessage(msg.Chat, "Anything worth noting since we last chatted?", &telebot.SendOptions{
		ReplyMarkup: telebot.ReplyMarkup{
			ForceReply: true,
		},
	})
	return
}

func (b fantailBot) help(msg telebot.Message) {
	fBot.bot.SendMessage(msg.Chat, "Hi, "+msg.Chat.FirstName+" we will be in touch", nil)
	return
}

func (b fantailBot) greet(msg telebot.Message) {
	fBot.bot.SendMessage(msg.Chat, "Hello, "+msg.Chat.FirstName+"!", nil)
	return
}

func (b fantailBot) thanks(msg telebot.Message) {
	fBot.bot.SendMessage(msg.Chat, "Thanks "+msg.Chat.FirstName+" its always nice to chat!", nil)
	return
}

func (theSteps steps) walkSteps(message telebot.Message) {
	for step := range theSteps.ToWalk {
		if theSteps.ToWalk[step].done == false {
			theSteps.ToWalk[step].done = true
			theSteps.ToWalk[step].fn(message)
			return
		}
		if step == len(theSteps.ToWalk) {
			theSteps.Walking = false
		}
	}
	return
}

func (theSteps steps) inPlay(typeOf string) bool {
	return theSteps.Name == typeOf && theSteps.Walking
}

func main() {

	fBot = initBot()

	messages := make(chan telebot.Message)
	fBot.bot.Listen(messages, 1*time.Second)

	var currentSteps *steps

	for message := range messages {

		log.Println("message", message.Text, "from", message.Chat.FirstName, "id", message.Chat.ID)
		//

		if message.Text == chat || currentSteps != nil && currentSteps.inPlay(chat) {
			if currentSteps == nil {
				currentSteps = fBot.initChatSteps(message.Chat.ID)
			}
			log.Println("walking ...", currentSteps.Name)
			currentSteps.Walking = true
			currentSteps.walkSteps(message)
		}

		if message.Text == note || currentSteps != nil && currentSteps.inPlay(note) {
			if currentSteps == nil {
				currentSteps = fBot.initNoteSteps(message.Chat.ID)
			}
			log.Println("walking ...", currentSteps.Name)
			currentSteps.Walking = true
			currentSteps.walkSteps(message)
		}

		if message.Text == bg || currentSteps != nil && currentSteps.inPlay(bg) {
			if currentSteps == nil {
				currentSteps = fBot.initBgSteps(message.Chat.ID)
			}
			log.Println("walking ...", currentSteps.Name)
			currentSteps.Walking = true
			currentSteps.walkSteps(message)
		}
	}
}
