package telegram

import "strconv"

func InlineKeyboard(eventID int64, stats map[string]int) map[string]any {
	id := strconv.FormatInt(eventID, 10)

	return map[string]any{
		"inline_keyboard": [][]map[string]string{
			{
				{
					"text":          "🏃 Побегу (" + strconv.Itoa(stats["going"]) + ")",
					"callback_data": "event:" + id + ":going",
				},
				{
					"text":          "🤔 Возможно (" + strconv.Itoa(stats["maybe"]) + ")",
					"callback_data": "event:" + id + ":maybe",
				},
				{
					"text":          "❌ Не смогу (" + strconv.Itoa(stats["no"]) + ")",
					"callback_data": "event:" + id + ":no",
				},
			},
		},
	}
}
