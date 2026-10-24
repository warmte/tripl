package common

import (
	"strconv"
	"strings"
	"time"
	"unicode"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/warmte/tripl/lang"
)

func BMain(cfg *lang.Cfg) [][]string {
	return [][]string{
		{cfg.Buttons.Add},
		{cfg.TripObjects.Ticket.Many, cfg.TripObjects.Booking.Many, cfg.TripObjects.Event.Many, cfg.TripObjects.Trip.Many},
		{cfg.Buttons.Settings}}
}

func BList(listType ListType, cfg *lang.Cfg, objectCfg *lang.TripObjectCfg) [][]string {
	first := objectCfg.Actual
	if listType == ArchiveListType {
		first = objectCfg.Archive
	}
	return [][]string{{first, cfg.Buttons.Back}}
}

func BListAll(cfg *lang.Cfg, objectCfg *lang.TripObjectCfg) [][]string {
	return [][]string{{objectCfg.Actual, objectCfg.Archive, cfg.Buttons.Back}}
}

func BuildKeyboard(buttons [][]string, cfg *lang.Cfg) [][]tgbotapi.KeyboardButton {
	raws := make([][]tgbotapi.KeyboardButton, len(buttons))
	for i, r := range buttons {
		cols := make([]tgbotapi.KeyboardButton, len(r))
		for j, b := range buttons[i] {
			cols[j] = tgbotapi.NewKeyboardButton(b)
		}
		raws[i] = tgbotapi.NewKeyboardButtonRow(cols...)
	}
	return raws
}

func BuildMsgWithKeyboard(chatId int64, text string, buttons [][]string, cfg *lang.Cfg) tgbotapi.Chattable {
	raws := BuildKeyboard(buttons, cfg)
	msg := tgbotapi.NewMessage(chatId, lang.EscapeMarkdownV2(text))
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(raws...)
	msg.ParseMode = tgbotapi.ModeMarkdownV2
	return msg
}

func cropStr(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n])
}

func Normalize(value *string) *string {
	if value != nil && (*value == "null" || *value == "") {
		return nil
	} else if value != nil {
		r := cropStr(strings.ReplaceAll(*value, "\n", " "), 256)
		return &r
	} else {
		return value
	}
}

func NormalizeCost(raw string) string {
	s := strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, raw)
	if i := strings.LastIndex(s, ","); i >= 0 && !strings.Contains(s, ".") {
		if len(s)-i-1 == 3 {
			s = strings.ReplaceAll(s, ",", "")
		} else {
			s = strings.Replace(s, ",", ".", 1)
		}
	}
	if _, err := strconv.ParseFloat(s, 64); err != nil {
		return raw
	}
	return s
}

func NormalizeCurrency(rawStr string) string {
	raw := cropStr(strings.TrimSpace(rawStr), 256)
	s := strings.ToLower(strings.TrimSpace(raw))
	if code, ok := currencyAliases[s]; ok {
		return code
	}
	return strings.ToUpper(s)
}

var currencyAliases = map[string]string{
	"₽": "RUB", "р": "RUB", "р.": "RUB", "руб": "RUB", "руб.": "RUB", "рубль": "RUB", "рублей": "RUB", "rub": "RUB",
	"$": "USD", "usd": "USD", "долл": "USD", "долл.": "USD", "доллар": "USD", "долларов": "USD", "бакс": "USD",
	"€": "EUR", "eur": "EUR", "евро": "EUR",
	"₩": "KRW", "krw": "KRW", "вон": "KRW", "вона": "KRW", "вонов": "KRW",
	"£": "GBP", "gbp": "GBP", "фунт": "GBP", "фунтов": "GBP",
	"¥": "JPY", "jpy": "JPY", "иена": "JPY", "иен": "JPY",
	"cny": "CNY", "юань": "CNY", "юаней": "CNY", "元": "CNY",
	"chf": "CHF", "франк": "CHF",
	"cad": "CAD", "c$": "CAD",
	"aud": "AUD", "a$": "AUD",
	"₹": "INR", "inr": "INR", "рупия": "INR", "рупий": "INR",
	"₺": "TRY", "try": "TRY", "лира": "TRY", "лир": "TRY",
	"₸": "KZT", "kzt": "KZT", "тенге": "KZT",
	"₴": "UAH", "uah": "UAH", "гривна": "UAH", "гривен": "UAH",
	"byn": "BYN", "бел. руб": "BYN", "белорусский рубль": "BYN",
	"₾": "GEL", "gel": "GEL", "лари": "GEL",
	"aed": "AED", "дирхам": "AED",
	"sar": "SAR", "риял": "SAR", "риалов": "SAR",
	"฿": "THB", "thb": "THB", "бат": "THB",
	"sgd": "SGD", "s$": "SGD",
	"hkd": "HKD", "hk$": "HKD",
	"nzd": "NZD", "nz$": "NZD",
	"sek": "SEK", "крона": "SEK", "крон": "SEK",
	"nok": "NOK",
	"dkk": "DKK",
	"zł":  "PLN", "pln": "PLN", "злотый": "PLN", "злотых": "PLN",
	"czk": "CZK", "кч": "CZK",
	"brl": "BRL", "r$": "BRL",
	"mxn": "MXN", "мекс. песо": "MXN",
	"idr": "IDR", "рупия индонезии": "IDR",
	"myr": "MYR", "риннгит": "MYR", "ринггит": "MYR",
	"₱": "PHP", "php": "PHP", "песо": "PHP",
	"₫": "VND", "vnd": "VND", "донг": "VND",
	"pkr": "PKR", "пакистанская рупия": "PKR",
	"bdt": "BDT", "така": "BDT",
	"egp": "EGP", "египетский фунт": "EGP",
	"₦": "NGN", "ngn": "NGN", "найра": "NGN",
	"kes": "KES", "кенийский шиллинг": "KES",
	"ghs": "GHS", "седи": "GHS",
	"mad": "MAD", "марокканский дирхам": "MAD",
	"twd": "TWD", "nt$": "TWD",
	"ars": "ARS", "аргентинское песо": "ARS",
	"clp": "CLP", "чилийское песо": "CLP",
	"cop": "COP", "колумбийское песо": "COP",
	"pen": "PEN", "соль": "PEN",
	"uyu": "UYU", "уругвайское песо": "UYU",
	"ron": "RON", "лей": "RON",
	"huf": "HUF", "форинт": "HUF",
	"bgn": "BGN", "лев": "BGN", "лева": "BGN",
}

func optStr(s *string) string {
	if s == nil {
		return ""
	} else {
		return *s
	}
}

func joinNonEmpty(sep string, parts ...string) string {
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return strings.Join(out, sep)
}

func optDateStr(start, end *time.Time, layout string) string {
	format := func(t *time.Time) string {
		if t == nil {
			return "?"
		}
		return t.Format(layout)
	}
	if start == nil && end == nil {
		return ""
	}
	if end == nil {
		return format(start)
	}
	return format(start) + " - " + format(end)
}
