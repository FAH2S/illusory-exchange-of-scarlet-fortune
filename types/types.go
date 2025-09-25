package types


import (
    "fmt"
)
// Kraken sometimes uses CamelCase and sometimes `_` not consistant


//{{{ Candle-s
type Candle struct {
    Time    int64   `json:"time"`
    Open    float64 `json:"open,string"`
    High    float64 `json:"high,string"`
    Low     float64 `json:"low,string"`
    Close   float64 `json:"close,string"`
    Volume  float64 `json:"volume,string"`
}
type CandleResponse struct {
    Candles     []Candle    `json:"candles"`
    MoreCandles bool        `json:"more_candles"`
}
type CandleMeta struct {
    TickType    string  `json:"tick_type"`
    Symbol      string  `json:"symbol"`
    Resolution  string  `json:"resolution"`
    SinceDays   int     `json:"since_days"`
}
type CandleResponseWithMeta struct {
    Meta        CandleMeta      `json:"meta"`
    Response    CandleResponse  `json:"response"`
}
 //}}} Candle-s


//{{{ Open position-s
// Some fields can be null/ommited thus `*` pointer
type OpenPosition struct {
    Side                string      `json:"side"`
    Symbol              string      `json:"symbol"`
    Price               float64     `json:"price"`
    FillTime            string      `json:"fillTime"`
    Size                float64     `json:"size"`
    // Can be null or ommited sometimes (ex.: single collateral won't have pnlCurrency and multi will)
    UnrealizedFunding   *float64    `json:"unrealizedFunding,omitempty"`
    PnlCurrency         *string     `json:"pnlCurrency,omitempty"`
    MaxFixedLeverage    *float64    `json:"maxFixedLeverage,omitempty"`
}
type OpenPositionResponse struct {
    Result          string          `json:"result"`
    ServerTime      string          `json:"serverTime"`
    // Optional, only on success
    OpenPositions   *[]OpenPosition `json:"openPositions,omitempty"`
    // Optional, only on failure
    Error           *string         `json:"error,omitempty"`
}
//}}} Open position-s


//{{{ History, past order-s
// TODO: might be conflicting with order as orders that are active or even executed
type Order struct {
    Time    string  `json:"time"`
    Price   float64 `json:"price"`
    Size    float64 `json:"size"`
    Side    string  `json:"side"`
    Type    string  `json:"type"`
    UID     string  `json:"uid"`
}

//}}} History, past order-s 


///{{{ Ticker
// There are many more parameters but so far we don't care about them
type Ticker struct {
    Symbol      string  `json:"symbol"`
    MarkPrice   float64 `json:"markPrice"`
    Change24h   float64 `json:"change24h"`
    Suspended   bool    `json:"suspended"`
    PostOnly    bool    `json:"postOnly"`
}
type TickerResponse struct {
    Result      string  `json:"result"`
    ServerTime  string  `json:"serverTime"`
    // Optional, only on success
    Ticker      Ticker  `json:"ticker,omitempty"`
    // Optional, only on failure
    Error       *string `json:"error,omitempty"`
}
//}}} Ticker


//{{{ Order
// CliOrdId must be unique so lets make it: '{symbol}-{orderType}-{side}@{limitPrice}'
//  ex.: `PF_BCHUSD-[lmt/post]:buy@550`                         for limit order
//  ex.: `PF_BCHUSD-stp:trig@550-sell@650-[mark,index,last]`    for trigger entry
type SendOrderRequest struct {
    Symbol          string      `json:"symbol"                  url:"symbol"`
    OrderType       string      `json:"orderType"               url:"orderType"`                // lmt, post, stp, mkt
    Side            string      `json:"side"                    url:"side"`                     // buy, sell
    Size            float64     `json:"size"                    url:"size"`
    LimitPrice      float64     `json:"limitPrice"              url:"limitPrice"`
    CliOrdId        string      `json:"cliOrdId"                url:"cliOrdId"`                 // must be unique
    StopPrice       *float64    `json:"stopPrice,omitempty"     url:"stopPrice,omitempty"`
    ReduceOnly      *bool       `json:"reduceOnly,omitempty"    url:"reduceOnly,omitempty"`
    TriggerSignal   *string     `json:"triggerSignal,omitempty" url:"triggerSignal,omitempty"`  // mark, index, last
}
func (sor *SendOrderRequest) CreateLimitOrderId() {
    sor.CliOrdId = fmt.Sprintf("%s-%s:%s@%.8f",
        sor.Symbol, sor.OrderType, sor.Side, sor.LimitPrice,
    )
}
func (sor *SendOrderRequest) CreateTriggerEntryOrderId() error{
    if sor.StopPrice == nil {
        return fmt.Errorf("StopPrice must be set, currently: `nil`")
    }
    if sor.TriggerSignal == nil {
        return fmt.Errorf("TriggerSignal must be set, currently: `nil`")
    }
    sor.CliOrdId = fmt.Sprintf("%s-%s:trig@%.6f-%s@%.6f-%s",
        sor.Symbol, sor.OrderType, *sor.StopPrice, sor.Side, sor.LimitPrice, *sor.TriggerSignal,
    )
    return nil
}
// Many more parameters but we don't care about them so far
type SendStatus struct {
    OrderId string  `json:"order_id"`
    Status  string  `json:"status"`     // placed succ, anything else failure
}
type SendOrderResponse struct {
    Result      string      `json:"result"`
    ServerTime  string      `json:"serverTime"`
    SendStatus  SendStatus  `json:"sendStatus"`
}


// OPEN/ACTIVE orders
type OpenOrder struct {
    OrderId         string      `json:"order_id"`
    Symbol          string      `json:"symbol"`
    Side            string      `json:"side"`
    OrderType       string      `json:"orderType"`
    LimitPrice      float64     `json:"limitPrice"`
    FilledSize      float64     `json:"filledSize"`
    UnfilledSize    float64     `json:"unfilledSize"`
    Status          string      `json:"status"`
    ReduceOnly      bool        `json:"reduceOnly"`
    ReceivedTime    string      `json:"receivedTime"`
    LastUpdateTime  string      `json:"lastUpdateTime"`
    // Optiona if present
    CliOrdId        *string      `json:"cliOrdId,omitempty"`
    // Optional only for `stp/take_profit` orderType
    StopPrice       *float64     `json:"stopPrice,omitempty"`
    TriggerSignal   *string      `json:"triggerSignal,omitempty"`
}
type OpenOrdersResponse struct {
    Result      string      `json:"result"`
    ServerTime  string      `json:"serverTime"`
    OpenOrders  []OpenOrder `json:"openOrders,omitempty"`
}


// Batch order response
type BatchStatus struct {
    Status  string  `json:"status"`
    OrderId string  `json:"order_id"`
}
type BatchOrderResponse struct {
    Result      string          `json:"result"`
    ServerTime  string          `json:"serverTime"`
    BatchStatus []BatchStatus   `json:"batchStatus"`
}
//}}} Order


//{{{ Fill
// Multiple different fillId's can reference same orderId, when partially exectuted
type Fill struct {
    FillId      string  `json:"fill_id"`
    Symbol      string  `json:"symbol"`
    Side        string  `json:"side"`
    OrderId     string  `json:"order_id"`
    Size        float64 `json:"size"`
    Price       float64 `json:"price"`
    FillTime    string  `json:"fillTime"`
    FillType    string  `json:"fillType"`
}
type FillsResponse struct {
    Result      string  `json:"result"`
    ServerTime  string  `json:"serverTime"`
    Fills       []Fill  `json:"fills"`
}
//}}} Fill




