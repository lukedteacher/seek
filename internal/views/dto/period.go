package dto

import (
	"seek/internal/domain/models"
	"strconv"
	"strings"
)

type PeriodView struct {
	ID        string             `json:"period.id"`
	Title     string             `json:"period.title"`
	StartTime string             `json:"period.start_time"`
	EndTime   string             `json:"period.end_time"`
	Duration  int                `json:"period.duration"`
	Days      models.DaysSignals `json:"period.days"`
	Row       string             `json:"period.row"`
}

func NewViewFromPeriod(p *models.Period) PeriodView {
	if p == nil {
		return PeriodView{}
	}
	startTime := "9:00"
	row := timeToRow(startTime, 479)
	if p.StartTime != "" {
		row = timeToRow(p.StartTime, 479)
	}
	endTime := "9:30"
	if p.StartTime != "" {
		endTime = add(p.StartTime, int(p.Duration))
	}
	days := models.DaysBitmaskToDaysSignals(p.Days)
	return PeriodView{
		ID:        p.ID,
		Title:     p.Title,
		StartTime: p.StartTime,
		EndTime:   endTime,
		Duration:  int(p.Duration),
		Days:      days,
		Row:       row,
	}
}

func NewPeriodFromView(pv *PeriodView) models.Period {
	if pv == nil {
		return models.Period{}
	}
	daysBitmask := models.DaysSignalsToDaysBitmask(pv.Days)
	return models.Period{
		ID:        pv.ID,
		Title:     pv.Title,
		StartTime: pv.StartTime,
		Duration:  int64(pv.Duration),
		Days:      daysBitmask,
	}
}

func add(time string, duration int) string {
	totalMinutes := timeToMinutes(time) + duration
	hours := totalMinutes / 60
	minutes := totalMinutes % 60
	return strconv.Itoa(hours) + ":" + strconv.Itoa(minutes)
}

func timeToMinutes(time string) int {
	parts := strings.Split(time, ":")
	hours, _ := strconv.Atoi(parts[0])
	minutes, _ := strconv.Atoi(parts[1])
	return hours*60 + minutes
}

func timeToRow(time string, offset int) string {
	totalMinutes := timeToMinutes(time)
	return strconv.Itoa(totalMinutes - offset)
}