package ur

import (
	"context"
)

type URController struct {
	*URCommon
	Positions PositionMap
}

func NewController(ctx context.Context, cfg URConfig) *URController {
	return &URController{
		URCommon: &URCommon{
			Ctx: ctx,
			cfg: cfg,
		},
		Positions: NewPositionMap(),
	}
}
