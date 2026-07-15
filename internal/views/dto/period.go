package dto

import (
	"fmt"
	"seek/internal/domain/models"
	"strconv"
	"strings"
)

type PeriodView struct {
	ID        string        `json:"id"`
	Title     string        `json:"title"`
	StartTime string        `json:"start_time"`
	EndTime   string        `json:"end_time"`
	Duration  int           `json:"duration"`
	Days      string        `json:"days"`
	Students  []StudentView `json:"students"`
	Row       string        `json:"row"`
	Columns   []int         `json:"columns"`
}

type PeriodFormView struct {
	ID         string             `json:"id"`
	Title      string             `json:"title"`
	StartTime  string             `json:"start_time"`
	EndTime    string             `json:"end_time"`
	Duration   int                `json:"duration"`
	Days       models.DaysSignals `json:"days"`
	StudentIDs []string           `json:"student_ids"`
}

func NewViewFromPeriod(p *models.Period) (PeriodView, error) {
	if p == nil {
		return PeriodView{}, nil
	}
	if p.StartTime == "" {
		return PeriodView{}, fmt.Errorf("start time not initialized in period")
	}
	if p.Duration == 0 {
		return PeriodView{}, fmt.Errorf("duration not initialized in period")
	}
	return PeriodView{
		ID:        p.ID,
		Title:     p.Title,
		StartTime: p.StartTime,
		EndTime:   add(p.StartTime, int(p.Duration)),
		Duration:  int(p.Duration),
		Days:      BitmaskToInitial(p.Days),
		Columns:   models.DaysBitmaskToColumnNumbers(p.Days),
		Row:       timeToRow(p.StartTime, 479),
	}, nil
}

func NewPeriodFromView(pv *PeriodView) models.Period {
	if pv == nil {
		return models.Period{}
	}
	return models.Period{
		ID:         pv.ID,
		Title:      pv.Title,
		StartTime:  pv.StartTime,
		Duration:   int64(pv.Duration),
		Days:       0,
	}
}

// fails if period isn't created with default start time and duration values
func NewFormViewFromPeriod(p *models.Period) (PeriodFormView, error) {
	if p == nil {
		return PeriodFormView{}, nil
	}
	if p.StartTime == "" {
		return PeriodFormView{}, fmt.Errorf("start time not initialized in period")
	}
	if p.Duration == 0 {
		return PeriodFormView{}, fmt.Errorf("duration not initialized in period")
	}
	days := models.DaysBitmaskToDaysSignals(p.Days)
	return PeriodFormView{
		ID:        p.ID,
		Title:     p.Title,
		StartTime: p.StartTime,
		EndTime:   add(p.StartTime, int(p.Duration)),
		Duration:  int(p.Duration),
		Days:      days,
	}, nil
}

func NewPeriodFromFormView(v *PeriodFormView) models.Period {
	if v == nil {
		return models.Period{}
	}
	return models.Period{
		ID:         v.ID,
		Title:      v.Title,
		StartTime:  v.StartTime,
		Duration:   int64(v.Duration),
		Days:       models.DaysSignalsToDaysBitmask(v.Days),
		StudentIDs: v.StudentIDs,
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
