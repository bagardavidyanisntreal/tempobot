package telegram

import (
	"fmt"

	"github.com/bagardavidyanisntreal/tempobot/internal/model"
)

func BuildEventText(event *model.Event, stats map[string]int) string {
	return fmt.Sprintf(
		"🏃 %s\n\n%s\n\n"+
			"🏃 Побегут — %d\n"+
			"🤔 Возможно — %d\n"+
			"❌ Не смогут — %d",
		event.Title,
		event.Description,
		stats["going"],
		stats["maybe"],
		stats["no"],
	)
}
