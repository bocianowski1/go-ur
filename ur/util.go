package ur

import (
	"fmt"
	"math"
)

func DegToRad(deg float64) float64 {
	return deg * math.Pi / 180.0
}

func RadToDeg(rad float64) float64 {
	return rad * 180.0 / math.Pi
}

func floatArrayToString(arr []float64) string {
	str := "["
	for i, f := range arr {
		str += fmt.Sprintf("%f", f)
		if i < len(arr)-1 {
			str += ", "
		}
	}
	str += "]"
	return str
}
