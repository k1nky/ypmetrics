package handler

import "strconv"

func convertToInt64(s string) (v int64, err error) {
	v, err = strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return
}

func convertToFloat64(s string) (v float64, err error) {
	v, err = strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	return
}
