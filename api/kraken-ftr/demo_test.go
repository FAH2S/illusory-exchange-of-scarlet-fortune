package krakenftr

import (
    "os"
    "testing"
    "time"
)
import (
    "github.com/FAH2S/illusory-exchange-of-scarlet-fortune/types"
)
// DISCLAIMER:
// These tests cannot guarantee 100% reproducibility or coverage,
// since they depend on external systems (live/demo APIs, network,
// time, and market data) that change constantly.
// Might also need eyeball outputs also some values like price of
// Send order might need to be updated to different values


var demo *Exchange
var sleepTime time.Duration

func TestMain(m *testing.M) {
    apiKeyPublic := os.Getenv("KRAKEN_API_KEY_PUBLIC")
    apiKeyPrivate := os.Getenv("KRAKEN_API_KEY_PRIVATE")
    if apiKeyPublic == "" || apiKeyPrivate == "" {
        panic("API keys not set ")
    }
    // Initialize once for all tests
    demo = &Exchange{
        baseURL: "https://demo-futures.kraken.com",
        publicKey: apiKeyPublic,
        privateKey: apiKeyPrivate,
    }
    sleepTime = 4 * time.Second

    // Run tests
    code := m.Run()
    os.Exit(code)
}


//{{{ Test GetOHLC
func TestGetOHLC(t *testing.T) {
    num_days := 13
    t.Logf("Sleep %d sec", sleepTime/1000000000)
    time.Sleep(sleepTime)

    candles, err := demo.GetOHLC("trade", "PF_BCHUSD", "1d", num_days)
    if err != nil {
        t.Fatalf("GetOHLC failed: %v", err)
    }

    // This only works if resolution is 1d
    num_fetched_candles := len(candles.Response.Candles)
    if num_fetched_candles != num_days {
        t.Errorf("Wrong number of candles\nExpected:\t%d\nGot:\t\t%d", num_days, num_fetched_candles)
    }
    // Check if OHLC data is present
    if candles.Response.Candles[0].Open == 0.0 {
        t.Errorf("Data might be wrong or its just 0.0\nCandle: %v", candles.Response.Candles[0])
    }

    t.Logf("Fetched %d candles\n\n", num_fetched_candles)
}
//}}} Test GetOHLC


//{{{ Test GetOpenPositions
func TestGetOpenPositions(t *testing.T) {
    t.Logf("Sleep %d sec", sleepTime/1000000000)
    time.Sleep(sleepTime)
    result, err := demo.GetOpenPositions()
    if err != nil {
        t.Fatalf("GetOpenPostions failed: %v", err)
    }
    if result.Result != "success" {
        t.Errorf("GetOpenPositions failed: result is %s", result.Result)
    }
    t.Logf("Result: %+v\n", result)
    t.Logf("Open positions: %+v\n\n", result.OpenPositions)
}


func TestGetTicker(t *testing.T) {
    t.Logf("Sleep %d sec", sleepTime/1000000000)
    time.Sleep(sleepTime)
    result, err := demo.GetTicker("PF_BCHUSD")
    if err != nil {
        t.Fatalf("GetTicker failed: %v", err)
    }
    if result.Result != "success" {
        t.Errorf("GetTicker failed: result is %s", result.Result)
    }
    t.Logf("Result: %+v\n", result)
    t.Logf("Open positions: %+v\n\n", result.Ticker)
}
//}}} Test GetOpenPositions


//{{{ Test GetActiveOrders
func TestGetActiveOrders(t *testing.T) {
    t.Logf("Sleep %d sec", sleepTime/1000000000)
    time.Sleep(sleepTime)
    result, err := demo.GetOpenOrders()
    if err != nil {
        t.Fatalf("GetOpenOrders failed: %v", err)
    }
    if result.Result != "success" {
        t.Errorf("GetOpenOrders failed: result is %s", result.Result)
    }
    t.Logf("Result: %+v\n", result)
    t.Logf("Open positions: %+v\n\n", result.OpenOrders)
}
//}}} Test GetActiveOrders


//{{{ Test BatchOrder
func TestBatchOrder(t *testing.T) {
    // PREPARATION
    orderIDs := []string{}
    postOrderReq := types.SendOrderRequest {
        OrderType:      "post",
        Symbol:         "PF_BCHUSD",
        Side:           "buy",
        Size:           0.25,
        LimitPrice:     524.0,
    }
    postOrderReq.CreateLimitOrderId()

    stopPrice := 524.0
    triggerSignal := "last"
    stpOrderReq := types.SendOrderRequest {
        OrderType:      "stp",
        Symbol:         "PF_BCHUSD",
        Side:           "sell",
        Size:           0.25,
        LimitPrice:     574.0,
        StopPrice:      &stopPrice,
        TriggerSignal:  &triggerSignal,
    }
    if err := stpOrderReq.CreateTriggerEntryOrderId(); err != nil {
        t.Fatalf("CreateTriggerEntryOrderId failed: %v", err)
    }

    batchOrder := []types.SendOrderRequest{}
    batchOrder = append(batchOrder, postOrderReq)
    batchOrder = append(batchOrder, stpOrderReq)

    // SEND BATCH ORDERS
    t.Logf("Sleep %d sec", sleepTime/1000000000)
    time.Sleep(sleepTime)
    result, err := demo.BatchSendOrders(batchOrder)
    if err != nil {
        t.Fatalf("BatchSendOrders failed: %v", err)
    }

    t.Logf("Result: %+v\n", result)
    t.Logf("Open positions: %+v\n\n", result.BatchStatus)
    for _, order := range(result.BatchStatus){
        orderIDs = append(orderIDs, order.OrderId)
    }
    t.Logf("Sent order ID's: %v", orderIDs)

    // CANCEL BATCH ORDERS
    t.Logf("Sleep %d sec", sleepTime/1000000000)
    time.Sleep(sleepTime)
    result, err = demo.BatchCancelOrders(orderIDs)
    if err != nil {
        t.Fatalf("BatchCancelOrders failed: %v", err)
    }

    t.Logf("Result: %+v\n", result)
    t.Logf("Open positions: %+v\n\n", result.BatchStatus)
    t.Logf("Canceled order ID's: %v", orderIDs)
}
//}}} Test BatchOrder


//{{{ Test GetOrderFills
func TestGetOrderFills(t *testing.T) {
    t.Logf("Sleep %d sec", sleepTime/1000000000)
    time.Sleep(sleepTime)
    result, err := demo.GetOrderFills(0) // 0 for last 100 fills
    if err != nil {
        t.Fatalf("GetOrderFills failed: %v", err)
    }
    t.Logf("Result: %+v\n", result)
    t.Logf("Open positions: %+v\n\n", result.Fills)
}
//}}} Test GetOrderFills




//{{{ some examples
/*
trigger entry example:
    stopPrice := 550.0
    triggerSignal := "last"
    lmtOrderReq := types.SendOrderRequest {
        OrderType:      "stp",
        Symbol:         "PF_BCHUSD",
        Side:           "sell",
        Size:           0.1,
        LimitPrice:     650,
        StopPrice:      &stopPrice,
        TriggerSignal:  &triggerSignal,
    }
    if err := lmtOrderReq.CreateTriggerEntryOrderId(); err != nil {
        t.Fatalf("CreateTriggerEntryOrderId failed: %v", err)
    }
    result, err := demo.SendOrder(lmtOrderReq)


limit/post example:
    lmtOrderReq := types.SendOrderRequest {
        OrderType:  "post",
        Symbol:     "PF_BCHUSD",
        Side:       "buy",
        Size:       0.1,
        LimitPrice: 550,
    }
    lmtOrderReq.CreateLimitOrderId()
    result, err := demo.SendOrder(lmtOrderReq)
*/
//}}}


