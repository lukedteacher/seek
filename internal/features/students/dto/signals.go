package dto

import "seek/internal/features/_shared/shareddto"

type TableSignals struct {
	Table struct {
		Sort   shareddto.TableSort `json:"sort"`
		Filter Filter              `json:"filter"`
	} `json:"table"`
}
