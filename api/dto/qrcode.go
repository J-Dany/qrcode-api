package dto

import (
	"image/color"
)

type QrcodeCreateDto struct {
	Label   string `json:"label" form:"label"`
	Data    string `json:"data" form:"data" validate:"required"`
	Options *struct {
		Transparent     *bool       `json:"transparent" form:"options[transparent]"`
		Size            *uint8      `json:"size" form:"options[size]"`
		Borders         []int       `json:"borders" form:"options[borders]"`
		ForegroundColor *color.RGBA `json:"foregroundColor" form:"options[foregroundColor]"`
		BackgroundColor *color.RGBA `json:"backgroundColor" form:"options[backgroundColor]"`
	} `json:"options" form:"options"`
}

type InserQrcodeDto struct {
	Label string           `json:"label"`
	Bytes []byte           `json:"bytes"`
	Data  *QrcodeCreateDto `json:"data"`
}
