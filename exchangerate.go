package tiga

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/thinkeridea/go-extend/exmath"
)

var onceGet sync.Once

type ExchangerateWorker struct {
	Base            string `json:"base"`
	Date            string `json:"date"`
	TimeLastUpdated int    `json:"time_last_updated"`
	Rates           map[string]float64
}

func getExchangerateToday() *ExchangerateWorker {
	url := "https://api.exchangerate-api.com/v4/latest/USD"

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
func (e *ExchangerateWorker) DoExchange(currency string, money float64) (dollor float64, err error) {
	if rate, ok := e.Rates[currency]; ok {
		return exmath.Round(money/rate, 4), nil
	}
	return 0.0, errors.New("Exchange rate conversion for this currency is not supported at the moment.")
}
