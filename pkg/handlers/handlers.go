package handlers

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/jackc/pgx/v5/pgxpool"
	data "github.com/ulugbek0217/taksi-bot/misc"
	hp "github.com/ulugbek0217/taksi-bot/pkg/helpers"
)

type Handlers struct {
	DB          *pgxpool.Pool
	Admin       int64
	GroupID     int64
	SendToGroup bool // Toggle flag: true = send to group, false = send to admin
	cache       Cache
}

type Cache struct {
	status    string
	phone     string
	direction string
	quantity  string
	isPackage bool
}

// SkipGroup is a middleware to skip group messages being answered
func (Handlers) SkipGroup(next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update.Message == nil ||
			(update.Message.Chat.Type == models.ChatTypeGroup ||
				update.Message.Chat.Type == models.ChatTypeSupergroup) {
			return
		}
		next(ctx, b, update)
	}
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

// ToggleToGroup toggles message sending to group
func (h *Handlers) ToggleToGroup(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil || update.Message.From.ID != h.Admin {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "❌ Bu buyruq faqat admin uchun!",
		})
		if err != nil {
			fmt.Printf("error on unauthorized toggle attempt: %v\n", err)
		}
		return
	}

	// Update memory
	h.SendToGroup = true

	// Save to database
	err := hp.SetSetting(ctx, h.DB, "send_to_group", "true")
	if err != nil {
		fmt.Printf("error saving toggle setting to database: %v\n", err)
		_, err = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "❌ Xatolik: Ma'lumotlar bazasiga saqlab bo'lmadi.",
		})
		if err != nil {
			fmt.Printf("err sending message: %v\n", err)
		}
		return
	}

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "✅ Buyurtmalar endi guruhga yuboriladi.",
	})
	if err != nil {
		fmt.Printf("error on toggle to group confirmation: %v\n", err)
	}

	fmt.Printf("Admin toggled message sending to GROUP (ID: %d) - SAVED TO DB\n", h.GroupID)
}

// ToggleToPM toggles message sending to admin PM
func (h *Handlers) ToggleToPM(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil || update.Message.From.ID != h.Admin {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "❌ Bu buyruq faqat admin uchun!",
		})
		if err != nil {
			fmt.Printf("error on unauthorized toggle attempt: %v\n", err)
		}
		return
	}

	// Update memory
	h.SendToGroup = false

	// Save to database
	err := hp.SetSetting(ctx, h.DB, "send_to_group", "false")
	if err != nil {
		fmt.Printf("error saving toggle setting to database: %v\n", err)
		_, err = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "❌ Xatolik: Ma'lumotlar bazasiga saqlab bo'lmadi.",
		})
		fmt.Printf("err sending message: %v\n", err)
		return
	}

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "✅ Buyurtmalar endi shaxsiy xabarga yuboriladi.",
	})
	if err != nil {
		fmt.Printf("error on toggle to PM confirmation: %v\n", err)
	}

	fmt.Printf("Admin toggled message sending to PRIVATE MESSAGES (Admin ID: %d) - SAVED TO DB\n", h.Admin)
}

// GetToggleStatus shows current toggle status with detailed info (admin only)
func (h *Handlers) GetToggleStatus(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil || update.Message.From.ID != h.Admin {
		return
	}

	var statusMessage string

	if h.SendToGroup {
		// Get group info
		groupInfo, err := b.GetChat(ctx, &bot.GetChatParams{ChatID: h.GroupID})
		var groupName string
		if err != nil {
			groupName = "Guruh nomi olinmadi"
			fmt.Printf("Error getting group info: %v\n", err)
		} else {
			if groupInfo.Title != "" {
				groupName = groupInfo.Title
			} else {
				groupName = "Nomsiz guruh"
			}
		}

		statusMessage = fmt.Sprintf(`📊 JORIY HOLAT

🎯 Buyurtmalar jo'natilayotgan joy:
📢 GURUH

📋 Guruh ma'lumotlari:
• Nomi: %s
• ID: %d
• Turi: %s

✅ Barcha yangi buyurtmalar ushbu guruhga yuboriladi.

🔄 O'zgartirish uchun:
• /pm_msg - Shaxsiy xabarga o'tkazish`,
			groupName, h.GroupID, groupInfo.Type)

	} else {
		// Get admin info
		adminInfo, err := b.GetChat(ctx, &bot.GetChatParams{ChatID: h.Admin})
		var adminName, adminUsername string
		if err != nil {
			adminName = "Admin nomi olinmadi"
			adminUsername = "Username olinmadi"
			fmt.Printf("Error getting admin info: %v\n", err)
		} else {
			adminName, adminUsername = hp.GetUserDisplayName(
				adminInfo.FirstName,
				adminInfo.LastName,
				adminInfo.Username,
			)
		}

		statusMessage = fmt.Sprintf(`📊 JORIY HOLAT

🎯 Buyurtmalar jo'natilayotgan joy:
👤 SHAXSIY XABAR

📋 Admin ma'lumotlari:
• Ism: %s
• Username: %s
• ID: %d

✅ Barcha yangi buyurtmalar shaxsiy xabarga yuboriladi.

🔄 O'zgartirish uchun:
• /group_msg - Guruhga o'tkazish`,
			adminName, adminUsername, h.Admin)
	}

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   statusMessage,
	})
	if err != nil {
		fmt.Printf("error on status check: %v\n", err)
	}
}

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

// MainHandler commits the order
func (h *Handlers) MainHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}
	// fmt.Printf("Chat type: %s\n", update.Message.Chat.Type)
	// if update.Message.Chat.Type == models.ChatTypeGroup || update.Message.Chat.Type == models.ChatTypeSupergroup {
	// 	return
	// }
	switch h.cache.status {
	case "":
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "⚠️ Iltimos tugmalardan foydalaning.",
		})

		if err != nil {
			fmt.Printf("error on func MainHandler on use keyboard message: %v\n", err)
			return
		}
		return
	case "quantity":
		h.CustomersQuantity(ctx, b, update)
		return
	case "phone":
		h.cache.phone = update.Message.Text
	case "bulkMessage":
		h.BulkMessageSender(ctx, b, update)
		return
	}

	// Get user info for better formatting
	var userName string

	if update.Message.From.FirstName != "" {
		userName = update.Message.From.FirstName
		if update.Message.From.LastName != "" {
			userName += " " + update.Message.From.LastName
		}
	} else {
		userName = "Noma'lum"
	}
	var proflink string
	if update.Message.From.Username != "" {
		proflink += fmt.Sprintf("(@%s)", update.Message.From.Username)
	}

	var infoMessage string
	if h.cache.isPackage {
		infoMessage = fmt.Sprintf(`🚚 YANGI POCHTA BUYURTMASI

📦 Xizmat turi: Pochta yetkazish
📍 Yo'nalish: %s
📞 Telefon raqami: %s
👤 Buyurtmachi ismi: <a href='tg://user?id=%d'>%s</a> %s`,
			h.cache.direction, h.cache.phone, update.Message.From.ID, userName, proflink)
	} else {
		infoMessage = fmt.Sprintf(`🚖 YANGI TAXI BUYURTMASI

🚗 Xizmat turi: Taxi (yo'lovchi)
📍 Yo'nalish: %s
👥 Yo'lovchilar soni: %s
📞 Telefon raqami: %s
👤 Buyurtmachi ismi: <a href='tg://user?id=%d'>%s</a> %s`,
			h.cache.direction, h.cache.quantity, h.cache.phone, update.Message.From.ID, userName, proflink)
	}

	// Determine where to send the message based on toggle
	var targetChatID int64
	if h.SendToGroup {
		targetChatID = h.GroupID
	} else {
		targetChatID = h.Admin
	}

	h.cache = Cache{}

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    targetChatID,
		Text:      infoMessage,
		ParseMode: models.ParseModeHTML,
	})
	if err != nil {
		fmt.Printf("error on func MainHandler on sending message to target (%d): %v\n", targetChatID, err)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Buyurtmani amalga oshirishda xatolik yuz berdi.",
		})
		return
	}

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "✅ Buyurtmangiz qabul qilindi. Siz bilan tez orada bog'lanamiz.",
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
		fmt.Printf("error on func MainHandler on order received message: %v\n", err)
	}

	h.cache.status = ""

	// Log where the message was sent
	if h.SendToGroup {
		fmt.Printf("MainHandler sent to GROUP (ID: %d)\n", h.GroupID)
	} else {
		fmt.Printf("MainHandler sent to ADMIN PM (ID: %d)\n", h.Admin)
	}
}

// BotUsersQuantity sends to the chat the number of registered users
func (h *Handlers) BotUsersQuantity(ctx context.Context, b *bot.Bot, update *models.Update) {
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

// PrepareBulkMessage prepares users state to send bulk message
func (h *Handlers) PrepareBulkMessage(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil || update.Message.From.ID != h.Admin {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "❌ Bu buyruq faqat admin uchun!",
		})
		if err != nil {
			fmt.Printf("error on unauthorized toggle attempt: %v\n", err)
		}
		return
	}

	h.cache.status = "bulkMessage"

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Foydalanuvchilar uchun xabarni keltiring",
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
	})

	if err != nil {
		log.Printf("error on func PrepareBulkMessage on bulk send message: %v\n", err)
	}
}

// BulkMessageSender sends given message to all users
func (h *Handlers) BulkMessageSender(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Yuborish boshlandi",
	})

	if err != nil {
		log.Printf("error sending message: %v\n", err)
	}

	// conn, err := pgxpool.New(context.Background(), os.Getenv("DB_PATH"))
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
	// 	os.Exit(1)
	// }
	// defer conn.Close()

	rows, err := h.DB.Query(context.Background(), "SELECT user_id FROM users WHERE status = 0")
	if err != nil {
		fmt.Printf("conn error while getting users %v\n", err)
	}
	defer rows.Close()

	var count int

	for rows.Next() {
		if count >= 30 {
			count = 0
			time.Sleep(time.Second * 1)
		}
		var id int64
		err := rows.Scan(&id)
		if err != nil {
			fmt.Printf("error scanning a row: %v\n", err)
		}

		_, err = b.CopyMessage(ctx, &bot.CopyMessageParams{
			ChatID:     id,
			FromChatID: update.Message.From.ID, // fmt.Sprintf("%d", update.Message.Chat.ID)
			MessageID:  update.Message.ID,
		})
		if err != nil {
			//_, err := h.DB.Exec(ctx, "DELETE FROM users WHERE user_id = $1", id)
			//if err != nil {
			//	fmt.Printf("error deleting from conn: %v\n", err)
			//}
			log.Printf("error copying message: %v\n", err)
		} else {
			count++
		}

		_, err = h.DB.Exec(ctx, "UPDATE users SET status = 1 WHERE user_id = $1", id)
		if err != nil {
			fmt.Printf("error setting status: %v\n", err)
		}

	}

	var sent_to int
	res := h.DB.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE status = 1")
	res.Scan(&sent_to)

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("Foydalanuvchilarga xabar yuborish tugatildi\nUmumiy hisobda %d ta foydalanuvchiga xabar yuborildi.", sent_to),
	})
	if err != nil {
		fmt.Printf("error sending message: %v", err)
	}

	_, err = h.DB.Exec(ctx, "UPDATE users SET status = 0 WHERE status = 1")
	if err != nil {
		fmt.Printf("error setting status: %v\n", err)
	}

	h.cache.status = ""
}
