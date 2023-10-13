package exchangerate

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sync"
)

var onceGet sync.Once

type ExchangerateWorker struct {
	Base            string `json:"base"`
	Date            string `json:"date"`
	TimeLastUpdated int    `json:"time_last_updated"`
	Rates           map[string]float64
}

func getExchangerateToday() *ExchangerateWorker {
	url := "https://api.exchangerate-api.com/v4/latest/USD" // 你的API endpoint

	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	// 初始化一个People结构体来存放我们的结果
	var worker ExchangerateWorker

	// 使用json包的Unmarshal函数解析json数据
	err = json.Unmarshal(body, &worker)
	if err != nil {
		panic(err)
	}
	return &worker
}
func NewExchangerateWorker() *ExchangerateWorker {
	worker := &ExchangerateWorker{}
	onceGet.Do(func() {
		worker = getExchangerateToday()
	})
	return worker
}
