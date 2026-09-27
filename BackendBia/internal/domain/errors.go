package domain

import "errors"

var (
	ErrNotFound           = errors.New("recurso no encontrado")
	ErrAnalysisInProgress = errors.New("ya hay un análisis en curso")
	ErrInvalidStatus      = errors.New("estado inválido")
	ErrInvalidCredentials = errors.New("credenciales inválidas")
)
