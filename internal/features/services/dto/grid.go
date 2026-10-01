// internal/features/services/dto/grid.go
package dto

import (
	"seek/internal/features/_shared/sharedmodels"
	"seek/internal/features/services/models"
	studentDTO "seek/internal/features/students/dto"
	studentModels "seek/internal/features/students/models"
)

type ServiceGrid struct {
	Columns       []sharedmodels.ServiceType `json:"columns"`
	Rows          []ServiceGridRow           `json:"-"`
	StudentFilter studentDTO.Filter          `json:"student_filter"`
}

type ServiceGridRow struct {
	Student studentModels.Student
	Cells   [][]models.Service
}

func NewServiceGrid(
	students []studentModels.Student,
	servicesByStudent map[string]map[sharedmodels.ServiceType][]models.Service,
	columns []sharedmodels.ServiceType,
) ServiceGrid {
	rows := make([]ServiceGridRow, 0, len(students))
	for _, s := range students {
		cells := make([][]models.Service, len(columns))
		if byType, ok := servicesByStudent[s.ID]; ok {
			for i, ct := range columns {
				cells[i] = byType[ct]
			}
		}
		rows = append(rows, ServiceGridRow{Student: s, Cells: cells})
	}
	return ServiceGrid{Columns: columns, Rows: rows}
}
