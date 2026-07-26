package dto

import (
	"fmt"
	"seek/internal/features/_shared/shareddto"
	"seek/internal/features/_shared/sharedmodels"
	"seek/internal/features/periods/models"
	sdto "seek/internal/features/students/dto"
	"strconv"
	"time"
)

type PeriodView struct {
	ID        string             `json:"id"`
	Title     string             `json:"title"`
	StartTime time.Time          `json:"start_time"`
	EndTime   time.Time          `json:"end_time"`
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
	if p.StartTime.IsZero() {
		return PeriodView{}, fmt.Errorf("start time not initialized in period")
	}
	if p.Duration == 0 {
		println("duration not initialized in period")
	}
	return PeriodView{
		ID:        p.ID,
		Title:     p.Title,
		StartTime: p.StartTime,
		EndTime:   p.StartTime.Add(time.Duration(p.Duration) * time.Minute),
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

func timeToRow(t time.Time, offset int) string {
	totalMinutes := t.Hour()*60 + t.Minute()
	return strconv.Itoa(totalMinutes - offset)
}
