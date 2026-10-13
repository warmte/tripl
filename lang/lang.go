package lang

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

type TripObjectCfg struct {
	One                string     `json:"one"`
	Many               string     `json:"many"`
	Archive            string     `json:"archive"`
	Actual             string     `json:"actual"`
	Empty              string     `json:"empty"`
	AddMainFields      string     `json:"add_main_fields"`
	AddOtherFields     string     `json:"add_other_fields"`
	AddOtherFieldsTips string     `json:"add_other_fields_tips"`
	MainFields         []string   `json:"main_fields"`
	OtherFields        []string   `json:"other_fields"`
	AddExample         ExampleCfg `json:"add_example"`
}

type TripObjectsCfg struct {
	Ticket  TripObjectCfg `json:"ticket"`
	Booking TripObjectCfg `json:"booking"`
	Event   TripObjectCfg `json:"event"`
	Trip    TripObjectCfg `json:"trip"`
}

type ButtonsCfg struct {
	Add      string  `json:"add"`
	Back     string  `json:"back"`
	Cancel   string  `json:"cancel"`
	Attach   string  `json:"attach"`
	Detach   string  `json:"detach"`
	Edit     string  `json:"edit"`
	Delete   string  `json:"delete"`
	Settings string  `json:"settings"`
	Skip     string  `json:"skip"`
	Stats    string  `json:"stats"`
	Lang     LangCfg `json:"lang"`
}

type ExampleCfg struct {
	Main  string `json:"main"`
	Other string `json:"other"`
}

type LangCfg struct {
	Ru   string `json:"ru"`
	En   string `json:"en"`
	Lang string `json:"lang"`
}

type ErrorsCfg struct {
	TooLongField               string `json:"too_long_field"`
	IncorrectDate              string `json:"incorrect_date"`
	IncorrectTime              string `json:"incorrect_time"`
	IncorrectDatetime          string `json:"incorrect_datetime"`
	IncorrectNumberOfFields    string `json:"incorrect_number_of_fields"`
	IncorrectInput             string `json:"incorrect_input"`
	DatetimeRestraint          string `json:"datetime_restraint"`
	ObjectsLimit               string `json:"objects_limit"`
	InternalError              string `json:"internal_error"`
	IncorrectFileFormat        string `json:"incorrect_file_format"`
	TooLargeFile               string `json:"too_large_file"`
	EmptyFile                  string `json:"empty_file"`
	DocumentFailed             string `json:"document_failed"`
	ErrDailyUploadLimitReached string `json:"daily_upload_limit_reached"`
}

type Cfg struct {
	Welcome           string         `json:"welcome"`
	ChooseLang        string         `json:"choose_lang"`
	Menu              string         `json:"menu"`
	Add               string         `json:"add"`
	Added             string         `json:"added"`
	AddAsFile         string         `json:"add_as_file"`
	Info              string         `json:"info"`
	Buttons           ButtonsCfg     `json:"buttons"`
	TripObjects       TripObjectsCfg `json:"trip_object"`
	Errors            ErrorsCfg      `json:"errors"`
	WhatEdit          string         `json:"what_edit"`
	OnEdit            string         `json:"on_edit"`
	OnEditReply       string         `json:"on_edit_reply"`
	Edited            string         `json:"edited"`
	OnDelete          string         `json:"on_delete"`
	Deleted           string         `json:"deleted"`
	SpentFor          string         `json:"spent_for"`
	SpentTotal        string         `json:"spent_total"`
	TomorrowPlans     string         `json:"tomorrow_plans"`
	DocumentInProcess string         `json:"document_in_process"`
	DocumentFinished  string         `json:"document_finished"`
}

//go:embed cfg/ru.json
var ruRaw []byte

//go:embed cfg/en.json
var enRaw []byte

var RuCfg = load(ruRaw)
var EnCfg = load(enRaw)

func load(data []byte) Cfg {
	var cfg Cfg

	if err := json.Unmarshal(data, &cfg); err != nil {
		panic(fmt.Sprintf("invalid config: %v", err))
	}

	return cfg
}

type Language string

const (
	LanguageRu Language = "ru"
	LanguageEn Language = "en"
)

func Get(language Language) *Cfg {
	switch language {
	case LanguageRu:
		return &RuCfg
	case LanguageEn:
		return &EnCfg
	default:
		return &RuCfg
	}
}
