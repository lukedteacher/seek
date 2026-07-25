package dto

import (
	"fmt"
	"seek/internal/features/_shared/shareddto"
	"seek/internal/features/_shared/sharedmodels"
	"seek/internal/features/periods/models"
	sdto "seek/internal/features/students/dto"
	"strconv"
	"strings"
)

type PeriodView struct {
	ID        string             `json:"id"`
	Title     string             `json:"title"`
	StartTime string             `json:"start_time"`
	EndTime   string             `json:"end_time"`
	Duration  int                `json:"duration"`
	Days      string             `json:"days"`
	Students  []sdto.StudentView `json:"students"`
	Row       string             `json:"row"`
	Columns   []int              `json:"columns"`
}

func NewViewFromPeriod(p *models.Period) (PeriodView, error) {
	if p == nil {
		return PeriodView{}, nil
	}
	if p.StartTime == "" {
		return PeriodView{}, fmt.Errorf("start time not initialized in period")
	}
	if p.Duration == 0 {
		println("duration not initialized in period")
	}
	return PeriodView{
		ID:        p.ID,
		Title:     p.Title,
		StartTime: p.StartTime,
		EndTime:   add(p.StartTime, int(p.Duration)),
		Duration:  int(p.Duration),
		Days:      shareddto.BitmaskToInitial(p.Days),
		Columns:   sharedmodels.DaysBitmaskToColumnNumbers(p.Days),
		Row:       timeToRow(p.StartTime, 479),
	}, nil
}

func NewPeriodFromView(pv *PeriodView) models.Period {
	if pv == nil {
		return models.Period{}
	}
	return models.Period{
		ID:        pv.ID,
		Title:     pv.Title,
		StartTime: pv.StartTime,
		Duration:  int64(pv.Duration),
		Days:      0,
	}
}

func add(time string, duration int) string {
	totalMinutes := timeToMinutes(time) + duration
	hours := totalMinutes / 60
	minutes := totalMinutes % 60
	return fmt.Sprintf("%d:%02d", hours, minutes)
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
