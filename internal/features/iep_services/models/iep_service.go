package models

import (
	"seek/internal/features/_shared/sharedmodels"
	"seek/pkg/templui/components/icon"

	"github.com/a-h/templ"
)

type IEPService struct {
	ID              string                   `display:"ID"`
	StudentID       string                   `display:"student ID" csv:"MARSS ID"`
	ServiceType     sharedmodels.ServiceType `display:"type" csv:"Service"`
	IndirectMinutes int                      `display:"indirect (min)" csv:"Indirect minutes"`
	DirectMinutes   int                      `display:"direct (min)" csv:"Direct minutes"`
	FrequencyCount  int                      `display:"freq count" csv:"Frequency count"`
	FrequencyType   string                   `display:"freq type" csv:"Frequency"`
	Location        string                   `display:"location"`
	StartDate       string                   `display:"start date" csv:"Start date"`
	EndDate         string                   `display:"end date" csv:"End date"`
	Provider        string                   `display:"provider" csv:"Provider"`
	CreatedAt       string
	UpdatedAt       string
	ArchivedAt      string
}

func NewIEPService() *IEPService {
	return &IEPService{}
}

func (s IEPService) Icon() templ.Component {
	return icon.Icon(s.ServiceType.IconName())(icon.Props{Size: "16"})
}
