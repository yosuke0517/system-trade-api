package service

import "fmt"

// 今のtrendを受け取って、trend == 2 && newTrend == 5またはtrend == 1 && newTrend == 4になっていたらtrue
func LossCut(trend int) bool {
	fmt.Println("trend")
	fmt.Println(trend)
	newTrend := SimpleSmaAnalysis()
	if (trend == 2 && newTrend == 1) || (trend == 1 && newTrend == 2) {
		return true
	}
	return false
}
