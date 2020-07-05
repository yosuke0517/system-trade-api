/*
bitflyer is access to bitflyterAPI
*/
package api

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const baseURL = "https://api.bitflyer.com/v1/"

// TODO usecaces/dto/配下へファイルとして格納
type APIClient struct {
	key        string
	secret     string
	httpClient *http.Client
}

func New(key, secret string) *APIClient {
	apiClient := &APIClient{key, secret, &http.Client{}}
	return apiClient
}

// header returns the map[string]string
func (api APIClient) header(method, endpoint string, body []byte) map[string]string {
	timeStamp := strconv.FormatInt(time.Now().Unix(), 10)
	message := timeStamp + method + endpoint + string(body)

	mac := hmac.New(sha256.New, []byte(api.secret))
	mac.Write([]byte(message))
	sign := hex.EncodeToString(mac.Sum(nil))
	return map[string]string{
		"ACCESS-KEY":       api.key,
		"ACCESS-TIMESTAMP": timeStamp,
		"ACCESS-SIGN":      sign,
		"Content-Type":     "application/json",
	}
}

func (api *APIClient) doRequest(method, urlPath string, query map[string]string, data []byte, isAllRes bool) (body []byte, statusCode int, err error) {
	baseURL, err := url.Parse(baseURL)
	if err != nil {
		log.Fatal(err)
	}
	apiURL, err := url.Parse(urlPath)
	if err != nil {
		log.Fatal(err)
	}
	endPoint := baseURL.ResolveReference(apiURL).String()

	// リクエストを作る
	req, err := http.NewRequest(method, endPoint, bytes.NewBuffer(data))
	if err != nil {
		log.Fatal(err)
	}
	// クエリー
	q := req.URL.Query()
	for key, value := range query {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()

	for key, value := range api.header(method, req.URL.RequestURI(), data) {
		req.Header.Add(key, value)
	}
	// APIアクセス
	resp, err := api.httpClient.Do(req)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return body, resp.StatusCode, nil
}

/*
getBalanceのレスポンス
https://lightning.bitflyer.com/docs/playground#GETv1%2Fme%2Fgetbalance/javascript
*/
type Balance struct {
	CurrentCode string  `json:"current_code"`
	Amount      float64 `json:amount`
	Available   float64 `json:available`
}

/*
現在所持している現金やビットコインの情報を取得する
*/
func (api *APIClient) GetBalance() ([]Balance, error) {
	url := "me/getbalance"
	resp, _, err := api.doRequest("GET", url, map[string]string{}, nil, false)
	if err != nil {
		log.Printf("action=GetBalance err=%s", err.Error())
		return nil, err
	}
	var balance []Balance
	err = json.Unmarshal(resp, &balance)
	if err != nil {
		log.Printf("action=GetBalance err=%s", err.Error())
		return nil, err
	}
	return balance, nil
}

/*
/v1/tickerのレスポンス
*/
type Ticker struct {
	ProductCode     string  `json:"product_code"`
	Timestamp       string  `json:"timestamp"`
	TickID          int     `json:"tick_id"`
	BestBid         float64 `json:"best_bid"`
	BestAsk         float64 `json:"best_ask"`
	BestBidSize     float64 `json:"best_bid_size"`
	BestAskSize     float64 `json:"best_ask_size"`
	TotalBidDepth   float64 `json:"total_bid_depth"`
	TotalAskDepth   float64 `json:"total_ask_depth"`
	Ltp             float64 `json:"ltp"`
	Volume          float64 `json:"volume"`
	VolumeByProduct float64 `json:"volume_by_product"`
}

/*
中間値を求める
*/
func (t *Ticker) GetMidPrice() float64 {
	return (t.BestBid + t.BestAsk) / 2
}

/*
データベースが対応している日付型になおすメソッド
*/
func (t *Ticker) DateTime() time.Time {
	dateTime, err := time.Parse(time.RFC3339, t.Timestamp)
	if err != nil {
		log.Printf("action=DateTime, err=%s", err.Error())
	}
	return dateTime
}

/*
時間変換用メソッド
@param 時間単位：h,m,s
@return time.Time（12:12:00 → duration hを与えると12:00:00に変換される
*/
func (t *Ticker) TruncateDateTime(duration time.Duration) time.Time {
	return t.DateTime().Truncate(duration)
}

/*
ビットコインの情報を取得する
*/
func (api *APIClient) GetTicker(productCode string) (*Ticker, error) {
	url := "ticker"
	resp, _, err := api.doRequest("GET", url, map[string]string{"product_code": productCode}, nil, false)
	if err != nil {
		log.Printf("action=getTicker err=%s", err.Error())
		return nil, err
	}
	var ticker Ticker
	err = json.Unmarshal(resp, &ticker)
	if err != nil {
		log.Printf("action=getTicker err=%s", err.Error())
		return nil, err
	}
	return &ticker, nil
}

type JsonRPC2 struct {
	Version string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	Result  interface{} `json:"result,omitempty"`
	Id      *int        `json:"id,omitempty"`
}

type SubscribeParams struct {
	Channel string `json:"channel"`
}

// リアルタイムTicker情報取得
func (api *APIClient) GetRealTimeTicker(symbol string, ch chan<- Ticker) {
	u := url.URL{Scheme: "wss", Host: "ws.lightstream.bitflyer.com", Path: "/json-rpc"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	channel := fmt.Sprintf("lightning_ticker_%s", symbol)
	if err := c.WriteJSON(&JsonRPC2{Version: "2.0", Method: "subscribe", Params: &SubscribeParams{channel}}); err != nil {
		log.Fatal("subscribe:", err)
		return
	}

OUTER:
	for {
		message := new(JsonRPC2)
		if err := c.ReadJSON(message); err != nil {
			log.Println("read:", err)
			return
		}

		if message.Method == "channelMessage" {
			switch v := message.Params.(type) {
			case map[string]interface{}:
				for key, binary := range v {
					if key == "message" {
						marshaTic, err := json.Marshal(binary)
						if err != nil {
							continue OUTER
						}
						var ticker Ticker
						if err := json.Unmarshal(marshaTic, &ticker); err != nil {
							continue OUTER
						}
						ch <- ticker
					}
				}
			}
		}
	}
}

// GetTradingCommission responce
type TradingCommission struct {
	CommissionRate float64 `json:"commission_rate"`
}

// get GetTradingCommission 手数料を取得する
func (api *APIClient) GetTradingCommission(productCode string) (*TradingCommission, error) {
	url := "me/gettradingcommission"
	resp, _, err := api.doRequest("GET", url, map[string]string{"product_code": productCode}, nil, false)
	if err != nil {
		log.Printf("action=GetTradingCommission err=%s", err.Error())
		return nil, err
	}
	var tradingCommission TradingCommission
	err = json.Unmarshal(resp, &tradingCommission)
	if err != nil {
		log.Printf("action=GetTradingCommission err=%s", err.Error())
		return nil, err
	}
	return &tradingCommission, nil
}

// SendOrder 送るdata
type Order struct {
	ID                     int     `json:"id"`
	ChildOrderAcceptanceID string  `json:"child_order_acceptance_id"`
	ProductCode            string  `json:"product_code"`
	ChildOrderType         string  `json:"child_order_type"`
	Side                   string  `json:"side"`
	Price                  float64 `json:"price"`
	Size                   float64 `json:"size"`
	MinuteToExpires        int     `json:"minute_to_expire"`
	TimeInForce            string  `json:"time_in_force"`
	Status                 string  `json:"status"`
	ErrorMessage           string  `json:"error_message"`
	AveragePrice           float64 `json:"average_price"`
	ChildOrderState        string  `json:"child_order_state"`
	ExpireDate             string  `json:"expire_date"`
	ChildOrderDate         string  `json:"child_order_date"`
	OutstandingSize        float64 `json:"outstanding_size"`
	CancelSize             float64 `json:"cancel_size"`
	ExecutedSize           float64 `json:"executed_size"`
	TotalCommission        float64 `json:"total_commission"`
	Count                  int     `json:"count"`
	Before                 int     `json:"before"`
	After                  int     `json:"after"`
}

// SendOrder responce
type ResponseSendChildOrder struct {
	ChildOrderAcceptanceID string `json:"child_order_acceptance_id"`
}

// 注文を送る
func (api *APIClient) SendOrder(order *Order) (*ResponseSendChildOrder, error) {
	data, err := json.Marshal(order)
	if err != nil {
		return nil, err
	}
	url := "me/sendchildorder"
	resp, _, err := api.doRequest("POST", url, map[string]string{}, data, false)
	if err != nil {
		return nil, err
	}
	var response ResponseSendChildOrder
	err = json.Unmarshal(resp, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// 注文の詳細を取得する
func (api *APIClient) ListOrder(query map[string]string) ([]Order, error) {
	resp, _, err := api.doRequest("GET", "me/getchildorders", query, nil, false)
	if err != nil {
		return nil, err
	}
	var responseListOrder []Order
	err = json.Unmarshal(resp, &responseListOrder)
	if err != nil {
		return nil, err
	}
	return responseListOrder, nil
}

// キャンセルStruct
type CancelOrder struct {
	ProductCode            string `json:"product_code"`
	ChildOrderAcceptanceID string `json:"child_order_acceptance_id"`
}

// オーダーをキャンセルする
func (api *APIClient) CancelOrder(cancelOrder *CancelOrder) (int, error) {
	data, err := json.Marshal(cancelOrder)
	if err != nil {
		return 400, err
	}
	url := "me/cancelchildorder"
	_, statusCode, err := api.doRequest("POST", url, map[string]string{}, data, true)
	if err != nil {
		return 400, err
	}
	return statusCode, err
}
