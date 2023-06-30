package telegram

import (
	"context"
	"net/url"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/zhashkevych/go-pocket-sdk"
)

const (
	commandStart = "start"
	// replyStartTemplate     = "Hi! In order to save links in your Pocket account, you first need to give me access. To do this, follow the link:\n %s"
	// replyAlreadyAuthorized = "You are already authorized. Just send link to save"
	// msgSaved               = "Link successfully saved"
)

func (b *Bot) handleCommand(message *tgbotapi.Message) error {
	switch message.Command() {
	case commandStart:
		return b.handleStartCommand(message)
	default:
		return b.handleUnknownCommand(message)
	}
}

func (b *Bot) handleMessage(message *tgbotapi.Message) error {
	_, err := url.ParseRequestURI(message.Text) // Проверка на валидность ссылки
	if err != nil {
		return errInvalidURL
	}

	accessToken, err := b.getAccessToken(message.Chat.ID) // Проверка на авторизацию
	if err != nil {
		return errUnauthorized
	}

	if pocketErr := b.pocketClient.Add(context.Background(), pocket.AddInput{ // Сохранение ссылки в Покет
		AccessToken: accessToken,
		URL:         message.Text,
	}); pocketErr != nil {
		return errUnableToSave
	}

	// msg.ReplyToMessageID = message.MessageID // Используется для reply-сообщений
	// log.Printf("[%s] %s", message.From.UserName, message.Text)
	msg := tgbotapi.NewMessage(message.Chat.ID, b.messages.SavedSuccessfully)
	_, err = b.bot.Send(msg)
	return err
}

// handleStartCommand - выведет сгенерированную ссылку для авторизации в Pocket
// После предоставления доступа произойдёт redirect по RedirectURL
// Сперва проверит наличие токена. Если его нет, сгенерирует
func (b *Bot) handleStartCommand(message *tgbotapi.Message) error {
	_, err := b.getAccessToken(message.Chat.ID)
	if err != nil {
		return b.initAuthorizationProcess(message) // Если токена нет, сгенерируй его
	}
	msg := tgbotapi.NewMessage(message.Chat.ID, b.messages.AlreadyAuthorized)
	_, sendErr := b.bot.Send(msg)
	if err != nil {
		return sendErr
	}
	return err
}

func (b *Bot) handleUnknownCommand(message *tgbotapi.Message) error {
	msg := tgbotapi.NewMessage(message.Chat.ID, b.messages.UnknownCommand) // Default message
	_, err := b.bot.Send(msg)
	return err
}
