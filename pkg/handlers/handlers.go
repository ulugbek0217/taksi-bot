package handlers

import (
	"context"
	"fmt"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/jackc/pgx/v5"
	data "github.com/ulugbek0217/taksi-bot/misc"
	hp "github.com/ulugbek0217/taksi-bot/pkg/helpers"
)

type Handlers struct {
	DB    *pgx.Conn
	Admin int64
	cache Cache
}

type Cache struct {
	status    string
	phone     string
	direction string
	quantity  string
	isPackage bool
}

// Start is the entry point
func (h *Handlers) Start(ctx context.Context, b *bot.Bot, update *models.Update) {
	h.cache = Cache{}
	if update.Message == nil {
		return
	}
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("Sizga TAXI xizmati zarur bo'lsa quyidagi raqamlarga murojaat qiling:\n\n%s\n%s\n\nYoki bot orqali buyurtma qilish uchun quyidagi yo'nalishlardan birini tanlang:", data.PhoneNumbers[0], data.PhoneNumbers[1]),
		ReplyMarkup: models.ReplyKeyboardMarkup{
			Keyboard: [][]models.KeyboardButton{
				{
					{
						Text: data.Directions[0],
					},
				},
				{
					{
						Text: data.Directions[1],
					},
				},
				{
					{
						Text: data.Directions[2],
					},
				},
				{
					{
						Text: data.Directions[3],
					},
				},
				{
					{
						Text: data.Directions[4],
					},
				},
				{
					{
						Text: data.Directions[5],
					},
				},
				{
					{
						Text: "🏠 Bosh sahifa",
					},
				},
			},
			ResizeKeyboard: true,
		},
	})

	if err != nil {
		fmt.Printf("error on func start on send message: %v\n", err)
		return
	}

	if !hp.UserExists(h.DB, update.Message.From.ID) {
		hp.RegMin(h.DB, update.Message.From.ID)
	}
}

// func (h Handlers) Passenger(ctx context.Context, b *bot.Bot, update *models.Update) {
// 	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
// 		ChatID: update.Message.Chat.ID,
// 		Text:   fmt.Sprintf("Sizga TAXI xizmati zarur bo'lsa quyidagi raqamlarga murojaat qiling:\n\n%s\n%s\n\nYoki bot orqali buyurtma qilish uchun quyidagi yo'nalishlardan birini tanlang:", data.PhoneNumbers[0], data.PhoneNumbers[1]),
// 		ReplyMarkup: models.ReplyKeyboardMarkup{
// 			Keyboard: [][]models.KeyboardButton{
// 				{
// 					{
// 						Text: data.Directions[0],
// 					},
// 				},
// 				{
// 					{
// 						Text: data.Directions[1],
// 					},
// 				},
// 				{
// 					{
// 						Text: data.Directions[2],
// 					},
// 				},
// 				{
// 					{
// 						Text: data.Directions[3],
// 					},
// 				},
// 				{
// 					{
// 						Text: data.Directions[4],
// 					},
// 				},
// 				{
// 					{
// 						Text: data.Directions[5],
// 					},
// 				},
// 				{
// 					{
// 						Text: "🏠 Bosh sahifa",
// 					},
// 				},
// 			},
// 			ResizeKeyboard: true,
// 		},
// 	})

// 	if err != nil {
// 		fmt.Printf("error on func start on send message: %v\n", err)
// 		return
// 	}
// }

// Direction handles the direction and asks for customers quantity or a package
func (h *Handlers) Direction(ctx context.Context, b *bot.Bot, update *models.Update) {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "👥 Nechta yo'lovchi yoki 📦pochta bor?",
		ReplyMarkup: models.ReplyKeyboardMarkup{
			Keyboard: [][]models.KeyboardButton{
				{
					{
						Text: "📦Pochta bor",
					},
				},
				{
					{
						Text: "1 kishi",
					},
					{
						Text: "2 kishi",
					},
				},
				{
					{
						Text: "3 kishi",
					},
					{
						Text: "4 kishi",
					},
				},
				{
					{
						Text: "🏠 Bosh sahifa",
					},
				},
			},
		},
	})
	if err != nil {
		fmt.Printf("error on func Direction on send message: %v\n", err)
		return
	}

	h.cache.direction = update.Message.Text
	h.cache.status = "quantity"
}

func (h *Handlers) Package(ctx context.Context, b *bot.Bot, update *models.Update) {

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "📞 Endi telefon raqamingizni yuboring.\n\nNamuna: +998901234567",
	})
	if err != nil {
		fmt.Printf("error on func Package on send message: %v\n", err)
		return
	}

	h.cache.isPackage = true
	h.cache.status = "phone"

}

func (h *Handlers) CustomersQuantity(ctx context.Context, b *bot.Bot, update *models.Update) {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "📞 Endi telefon raqamingizni yuboring.\n\nNamuna: +998901234567",
	})
	if err != nil {
		fmt.Printf("error on func CustomersQuantity on send message: %v\n", err)
		return
	}

	h.cache.quantity = update.Message.Text
	h.cache.status = "phone"

}

// Order commits the order
func (h *Handlers) Order(ctx context.Context, b *bot.Bot, update *models.Update) {
	if h.cache.status == "" {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "⚠️ Iltimos tugmalardan foydalaning.",
		})

		if err != nil {
			fmt.Printf("error on func Order on use keyboard message: %v\n", err)
			return
		}
		return
	} else if h.cache.status == "quantity" {
		h.CustomersQuantity(ctx, b, update)
		return
	} else if h.cache.status == "phone" {
		h.cache.phone = update.Message.Text
	}

	var infoMessage string
	if h.cache.isPackage {
		infoMessage = fmt.Sprintf("////////////////\n|| Pochta ||\n////////////////\nYo'nalish: %s\nBuyurtma beruvchi raqami: %s", h.cache.direction, h.cache.phone)
	} else {
		infoMessage = fmt.Sprintf("////////////////////\n|| Yo'lovchi ||\n////////////////////\nYo'nalish: %s\nYo'lovchilar: %s\nBuyurtma beruvchi raqami: %s", h.cache.direction, h.cache.quantity, h.cache.phone)
	}

	h.cache = Cache{}

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: h.Admin,
		Text:   infoMessage,
	})
	if err != nil {
		fmt.Printf("error on func Order on sending message to admin: %v\n", err)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Buyurtmani amalga oshirishda xatolik yuz berdi.",
		})
		return
	}
	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Buyurtmangiz qabul qilindi. Siz bilan tez orada bog'lanamiz.",
		ReplyMarkup: models.ReplyKeyboardMarkup{
			Keyboard: [][]models.KeyboardButton{
				{
					{
						Text: "🏠 Bosh sahifa",
					},
				},
			},
			ResizeKeyboard: true,
		},
	},
	)

	if err != nil {
		fmt.Printf("error on func Order on order received message: %v\n", err)
	}
}

// BotUsersQuantity sends to the chat the number of registered users
func (h Handlers) BotUsersQuantity(ctx context.Context, b *bot.Bot, update *models.Update) {
	count, err := hp.CountUsers(ctx, h.DB)
	if err != nil {
		fmt.Printf("error on func BotUsersQuantity on counting users: %v\n", err)
		return
	}

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("Ma'lumotlar bazasidagi foydalanuvchilar soni: %d\n", count),
	})

	if err != nil {
		fmt.Printf("error on func BotUsersQuantity on send users quantity message\n")
	}
}
