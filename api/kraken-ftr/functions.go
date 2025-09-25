package krakenftr

import (
    "crypto/sha256"
    "crypto/sha512"
    "crypto/hmac"
    "strings"
    "time"
    "fmt"
    "encoding/json"
    "encoding/base64"
    "net/http"
    "io"
)
import (
    "github.com/FAH2S/illusory-exchange-of-scarlet-fortune/types"
)
import (
    "github.com/google/go-querystring/query"
)


const (
    Reset = "\033[0m"
    Orange = "\033[38;5;208m"
)
type Exchange struct {
    baseURL     string
    publicKey   string
    privateKey  string
}


// URL = baseURL + endpoint + (optional) pathParams + (optional) query
//{{{ DRY
func makeRequest(
    method string,
    url string,
    body io.Reader,     // nil for GET, real body for POST
    publicKey string,
    signature string,
    nonce string,
) (*http.Response, error) {
    fmt.Printf("\n%s %s: %s %s\n", Orange, method, url, Reset)
    // Create request
    req, err := http.NewRequest(method, url, body)
    if err != nil {
        return nil, fmt.Errorf("Failed to create request: %w", err)
    }
    // Standard headers
    req.Header.Add("Accept", "application/json")
    req.Header.Add("APIKey", publicKey)
    req.Header.Add("Authent", signature)
    req.Header.Add("Nonce", nonce)

    client := &http.Client{}
    return client.Do(req)
}
//}}} DRY


//{{{ Sign
// Query(without '?') + data/body are part of signature
func (exch *Exchange) signRequestFn(
    endpoint string,    // ex.: "/derivatives/api/v3/openpositions"
    data string,        // query + data
    nonce string,       // number(as str) that HAS to be larger than previous one
) (string, error) {
    // Message = HASH(data + nonce + path/url/endpoint)
    message := sha256.New()
    // endpoint w trim = `/api/v3/openpositions`
    message.Write([]byte(data + nonce + strings.TrimPrefix(endpoint, "/derivatives")))
    digest := message.Sum(nil)

    // Extract key
    key, err := base64.StdEncoding.DecodeString(exch.privateKey)
    if err != nil {
        return "", err
    }

    // HmachHash = HMAC(message + key)
    hmacHash := hmac.New(sha512.New, key)
    hmacHash.Write(digest)
    return base64.StdEncoding.EncodeToString(hmacHash.Sum(nil)), nil
}
//}}} Sign


//{{{ Get OHLC
func sincePeriod(days int) int64 {
    return time.Now().Add(-time.Duration(days) * 24 * time.Hour).Unix()
}


func (exch *Exchange) GetOHLC(
    tickType string,    // "spot", "mark", "trade"
    symbol string,      // "PF_BCHUSD", ...
    resolution string,  // "1m", "5m", "15m", "30m", "1h", "4h", "12h", "1d", "1w"
    sinceDays int,      // 128, 32, ...
    ) (*types.CandleResponseWithMeta, error) {
    pathParams := fmt.Sprintf("/%s/%s/%s", tickType, symbol, resolution)
    url := exch.baseURL + "/api/charts/v1" + pathParams
    query := ""
    if sinceDays > 0 {
        // Convert last X days to unix timestamp
        since := sincePeriod(sinceDays)
        query += fmt.Sprintf("?from=%d", since)
    }

    fmt.Printf("\n%s GET: %s %s\n", Orange, (url+query), Reset)
    resp, err := http.Get(url+query)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result types.CandleResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    candleResponseWithMeta := types.CandleResponseWithMeta{
        Meta: types.CandleMeta{
            TickType:   tickType,
            Symbol:     symbol,
            Resolution: resolution,
            SinceDays:  sinceDays,
        },
        Response: result,
    }

    return &candleResponseWithMeta, nil
}
//}}} FetchOHLC


//{{{ Get open positions
func (exch *Exchange) GetOpenPositions() (*types.OpenPositionResponse, error) {
    nonce := fmt.Sprintf("%d", time.Now().UnixMilli()) // ms timestamp
    endpoint :=  "/derivatives/api/v3/openpositions"
    url := exch.baseURL + endpoint

    signature, err := exch.signRequestFn(endpoint, "", nonce)

    resp, err := makeRequest("GET", url, nil, exch.publicKey, signature, nonce)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    var result types.OpenPositionResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }
    return &result, nil
}
//}}}


//{{{ Get active/open orders
func (exch *Exchange) GetOpenOrders() (*types.OpenOrdersResponse, error){
    nonce := fmt.Sprintf("%d", time.Now().UnixMilli()) // ms timestamp
    endpoint :=  "/derivatives/api/v3/openorders"
    url := exch.baseURL + endpoint

    signature, err := exch.signRequestFn(endpoint, "", nonce)

    resp, err := makeRequest("GET", url, nil, exch.publicKey, signature, nonce)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    var result types.OpenOrdersResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }
    return &result, nil
}

//}}} Get active/open orders


//{{{ Get order fills aka own histroy WIP/AKA NOT DONE/AKA NEED MORE WORK
func (exch *Exchange) GetOrderFills(
    lastFillTime int,   //timestamp in unix ms time, cursor for pagination, not filter
    ) (*types.FillsResponse, error) {
    nonce := fmt.Sprintf("%d", time.Now().UnixMilli()) // ms timestamp
    endpoint :=  "/derivatives/api/v3/fills"
    url := exch.baseURL + endpoint
    query := ""
    if lastFillTime > 0 {
        // Convert last X days to unix miliseconds
        query += fmt.Sprintf("lastFillTime=%d", lastFillTime) // well we don't know if its sec or ms and can't test it until there is 100+ orders
    }

    // Query is considerd data for signature but withoug `?`
    signature, err := exch.signRequestFn(endpoint, query, nonce)

    resp, err := makeRequest("GET", url+"?"+query, nil, exch.publicKey, signature, nonce)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    var result types.FillsResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }
    return &result, nil
}
// TODO: GetOrderFills recursive up until given timestamp/date
// Get order fills as some abstract FN that gets all from last X days

//}}} Get order fills aka own histroy


//{{{ Get ticker
func (exch *Exchange) GetTicker(symbol string) (*types.TickerResponse, error) {
    nonce := fmt.Sprintf("%d", time.Now().UnixMilli()) // ms timestamp
    endpoint :=  "/derivatives/api/v3/tickers"
    pathParams := fmt.Sprintf("/%s", symbol)
    url := exch.baseURL + endpoint + pathParams

    signature, err := exch.signRequestFn(endpoint, "", nonce)

    resp, err := makeRequest("GET", url, nil, exch.publicKey, signature, nonce)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    var result types.TickerResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }
    return &result, nil
}
//}}} Get ticker


//{{{ Send order
// body data is expected to be url encoded
// ex.: "symbol=PF_BCHUSD&orderType=post&side=buy&size=0.1&limitPrice=550&cliOrdId=test123"
// cliOrdId must be unique it saves it server side each order has it
func (exch *Exchange) SendOrder(orderReq types.SendOrderRequest) (*types.SendOrderResponse, error) {
    nonce := fmt.Sprintf("%d", time.Now().UnixMilli()) // ms timestamp
    endpoint :=  "/derivatives/api/v3/sendorder"
    url := exch.baseURL + endpoint

    // Encode the order struct as URL-encoded form data (dynamic, works with optional fields)
    v, err := query.Values(orderReq)    // uses `url:xxx`
    if err != nil {
        return nil, err
    }
    // string for signing
    bodyString := v.Encode()
    // io.Reader for HTTP POST
    bodyReader := strings.NewReader(bodyString)

    signature, err := exch.signRequestFn(endpoint, bodyString, nonce)

    resp, err := makeRequest("POST", url, bodyReader, exch.publicKey, signature, nonce)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    var result types.SendOrderResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }
    return &result, nil
}
//}}} Send order


//{{{ Batch send order(s)
// "json":"batchOrder:{ {order1},{order2}, ... }, => encoded as json string (then later url encoded)
//{{{ helper fn
func structToMapBSO(order types.SendOrderRequest, i int) map[string]any {
    // convert struct to JSON
    data, _ := json.Marshal(order)
    // convert JSON into map
    var m map[string]any
    json.Unmarshal(data, &m)
    // add extra fields
    m["order"] = "send"
    m["order_tag"] = fmt.Sprintf("%d", i)
    // return
    return m
}
//}}} helperr fn


func (exch *Exchange) BatchSendOrders(orderReqList []types.SendOrderRequest) (*types.BatchOrderResponse, error) {
    nonce := fmt.Sprintf("%d", time.Now().UnixMilli()) // ms timestamp
    endpoint :=  "/derivatives/api/v3/batchorder"
    url := exch.baseURL + endpoint

    var batchOrder []map[string]any
    for i, order := range orderReqList {
        orderMap := structToMapBSO(order, i)
        batchOrder = append(batchOrder, orderMap)
    }

    batchOrderJSON := map[string]any {
        "batchOrder": batchOrder,
        }
    data, _ := json.Marshal(batchOrderJSON)

    bodyString := fmt.Sprintf("json=%s", data)
    bodyReader := strings.NewReader(bodyString)

    signature, err := exch.signRequestFn(endpoint, bodyString, nonce)

    resp, err := makeRequest("POST", url, bodyReader, exch.publicKey, signature, nonce)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    var result types.BatchOrderResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }
    return &result, nil

    // TODO save it as struct
    //body, _ := ioutil.ReadAll(resp.Body)
    //fmt.Println("\n\nResponse:", string(body))


}
//}}} Batch send order(s)


//{{{ Batch cancel order(s)
func (exch *Exchange) BatchCancelOrders(orderIDs []string) (*types.BatchOrderResponse, error) {
    nonce := fmt.Sprintf("%d", time.Now().UnixMilli()) // ms timestamp
    endpoint :=  "/derivatives/api/v3/batchorder"
    url := exch.baseURL + endpoint

    var batchOrder []map[string]string
    for _, orderID := range(orderIDs) {
        tmp := map[string]string{}
        tmp["order"] = "cancel"
        tmp["order_id"] = orderID
        batchOrder = append(batchOrder, tmp)
    }

    batchOrderJSON := map[string]any {
        "batchOrder": batchOrder,
        }
    data, _ := json.Marshal(batchOrderJSON)

    bodyString := fmt.Sprintf("json=%s", data)
    bodyReader := strings.NewReader(bodyString)

    signature, err := exch.signRequestFn(endpoint, bodyString, nonce)

    resp, err := makeRequest("POST", url, bodyReader, exch.publicKey, signature, nonce)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    var result types.BatchOrderResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }
    return &result, nil
}


//}}} Batch cancel order(s)


