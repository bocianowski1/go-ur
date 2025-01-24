package ur

type URPosition [6]float64

type IPositionMap interface {
	ToMap() map[string]URPosition
}

type PositionMap struct {
	Home       URPosition
	OverDevice URPosition
	PickDevice URPosition
	OverLetter URPosition
	PickLetter URPosition
}

var DefaultPositionMap = PositionMap{
	Home: URPosition{
		DegToRad(-78.73),
		DegToRad(-53.49),
		DegToRad(-109.61),
		DegToRad(-106.93),
		DegToRad(90.0),
		DegToRad(-4.04),
	},
	OverDevice: URPosition{
		DegToRad(-42.30),
		DegToRad(-111.12),
		DegToRad(-107.10),
		DegToRad(-51.81),
		DegToRad(89.98),
		DegToRad(-2.13),
	},
	PickDevice: URPosition{
		DegToRad(-42.30),
		DegToRad(-126.70),
		DegToRad(-112.97),
		DegToRad(-30.36),
		DegToRad(89.97),
		DegToRad(-2.13),
	},
	OverLetter: URPosition{
		DegToRad(-94.96),
		DegToRad(-115.20),
		DegToRad(-88.95),
		DegToRad(-65.87),
		DegToRad(90.0),
		DegToRad(-2.39),
	},
	PickLetter: URPosition{
		DegToRad(-94.95),
		DegToRad(-126.01),
		DegToRad(-104.50),
		DegToRad(-39.50),
		DegToRad(90.0),
		DegToRad(-11.16),
	},
}

func NewPositionMap() PositionMap {
	return DefaultPositionMap
}

func (p PositionMap) ToMap() map[string]URPosition {
	return map[string]URPosition{
		"home":        p.Home,
		"over_device": p.OverDevice,
		"pick_device": p.PickDevice,
		"over_letter": p.OverLetter,
		"pick_letter": p.PickLetter,
	}
}
