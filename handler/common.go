package handler

import (
	"context"
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	. "github.com/warmte/tripl/common"
	"github.com/warmte/tripl/lang"
	"github.com/warmte/tripl/service"
)

type CommonHandler[T TripObjectView] struct {
	label       Label
	service     service.Service[T]
	tripService service.Service[Trip]
}

func extractDate(action CallbackAction) *time.Time {
	input := string(action)
	if !strings.HasPrefix(input, ForDatePref) {
		return nil
	}
	date, err := time.Parse(DateLayout, strings.TrimPrefix(input, ForDatePref))
	if err != nil {
		return nil
	}
	return &date
}

func (h CommonHandler[T]) SelectAction(
	parsedCallback Callback,
	ctx context.Context,
	callback *tgbotapi.CallbackQuery,
	cfg *lang.Cfg,
) ([]tgbotapi.Chattable, error) {
	unsupportedCallbackError := fmt.Errorf(
		"unsupported callback: %s:%s:%d",
		parsedCallback.Label,
		parsedCallback.Action,
		parsedCallback.V1,
	)

	possibleDate := extractDate(parsedCallback.Action)
	if possibleDate != nil {
		r, err := h.HandleListForDate(ctx, callback, int(parsedCallback.V1), *possibleDate, cfg)
		return []tgbotapi.Chattable{r}, err
	}

	switch parsedCallback.Action {
	case ActualList, ArchiveList:
		r, err := h.HandleList(ctx, callback, parsedCallback, TripObjectCfgByLabel(cfg, parsedCallback.Label))
		return []tgbotapi.Chattable{r}, err
	case GetSingle:
		r, err := h.HandleGetSingle(ctx, callback, parsedCallback.V1, cfg)
		return []tgbotapi.Chattable{r}, err
	case EditSingle:
		r, err := h.HandleEditSingle(ctx, callback, parsedCallback.V1, cfg)
		return []tgbotapi.Chattable{r}, err
	case DeleteSingle:
		r, err := h.HandleDeleteSingle(ctx, callback, parsedCallback.V1, cfg)
		return []tgbotapi.Chattable{r}, err
	case DeleteSingleConfirmed:
		return h.HandleDeleteSingleConfirmed(ctx, callback, parsedCallback.V1, cfg)
	case DetachSingle:
		r, err := h.HandleDetachSingle(ctx, callback, parsedCallback.V1, cfg)
		return []tgbotapi.Chattable{r}, err

	default:
		return nil, unsupportedCallbackError
	}
}

func (h CommonHandler[T]) SelectExtendedAction(
	parsedCallback ExtendedCallback,
	ctx context.Context,
	callback *tgbotapi.CallbackQuery,
	cfg *lang.Cfg,
) ([]tgbotapi.Chattable, error) {
	unsupportedCallbackError := fmt.Errorf(
		"unsupported callback: %s:%s:%d%d",
		parsedCallback.Label,
		parsedCallback.Action,
		parsedCallback.V1,
		parsedCallback.V2,
	)

	switch parsedCallback.Action {
	case AttachSingleActual, AttachSingleArchived:
		r, err := h.HandleAttachSingle(ctx, callback, parsedCallback.Action, parsedCallback.V1, int(parsedCallback.V2), cfg)
		return []tgbotapi.Chattable{r}, err
	case AttachTrip:
		r, err := h.HandleAttachTrip(ctx, callback, parsedCallback.V1, parsedCallback.V2, cfg)
		return []tgbotapi.Chattable{r}, err
	case EditField:
		r, err := h.HandleEditField(ctx, callback, parsedCallback.V1, int(parsedCallback.V2), cfg)
		return []tgbotapi.Chattable{r}, err
	default:
		return nil, unsupportedCallbackError
	}
}

func button(text string, callback string) tgbotapi.InlineKeyboardButton {
	return tgbotapi.NewInlineKeyboardButtonData(text, callback)
}
