package models

import (
	"strconv"
	"strings"
)

var periodTypes = map[string]PeriodType{
	"reading": {Color: "#d55e00", ColorVar: "--cb-red", Icon: "book-open-text", Short: "reading"},
	"writing": {Color: "#e69f00", ColorVar: "--cb-red", Icon: "notebook-pen", Short: "writing"},
	"OT":      {Color: "#f0e442", ColorVar: "--cb-red", Icon: "pencil-ruler", Short: "ot"},
	"ex fun":  {Color: "#009e73", ColorVar: "--cb-red", Icon: "brain-cog", Short: "ef"},
	"vision":  {Color: "#56b4e9", ColorVar: "--cb-red", Icon: "eye", Short: "vision"},
	"math":    {Color: "#0072b2", ColorVar: "--cb-red", Icon: "calculator", Short: "math"},
	"SEL":     {Color: "#8430c9", ColorVar: "--cb-red", Icon: "heart-handshake", Short: "sel"},
	"speech":  {Color: "#cc79a7", ColorVar: "--cb-red", Icon: "message-circle-more", Short: "speech"},
}

type PeriodType struct {
	Color    string
	ColorVar string
	Icon     string
	Short    string
	Long     string
}

type Period struct {
	Id        string
	Title     string
	StartTime string
	Duration  int64
	Days      int64
	CreatedAt string
	UpdatedAt string
	DeletedAt string
}

type PeriodSignals struct {
	ID        string      `json:"id"`
	Title     string      `json:"title"`
	StartTime string      `json:"start_time"`
	Duration  int64       `json:"duration"`
	Days      DaysSignals `json:"days"`
}

func NewPeriod() *Period {
	return &Period{}
}

func (p *Period) DurationStr() string {
	return strconv.FormatInt(p.Duration, 64)
}

func (p *Period) TimeToRow(offset int) string {
	totalMinutes := timeToMinutes(p.StartTime)
	return strconv.Itoa(totalMinutes - offset)
}

func (p *Period) Color() string {
	return periodTypes[p.Title].Color
}

func (p *Period) Icon() string {
	return periodTypes[p.Title].Icon
}

func timeToMinutes(time string) int {
	parts := strings.Split(time, ":")
	hours, _ := strconv.Atoi(parts[0])
	minutes, _ := strconv.Atoi(parts[1])
	return hours*60 + minutes
}

func difference(StartTime, EndTime string) int {
	StartTimeMinutes := timeToMinutes(StartTime)
	EndTimeMinutes := timeToMinutes(EndTime)
	return EndTimeMinutes - StartTimeMinutes
}

func add(time string, duration int) string {
	totalMinutes := timeToMinutes(time) + duration
	hours := totalMinutes / 60
	minutes := totalMinutes % 60
	return strconv.Itoa(hours) + ":" + strconv.Itoa(minutes)
}
