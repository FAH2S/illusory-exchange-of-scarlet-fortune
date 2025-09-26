package dbfns
import (
    "testing"
    "reflect"
    "strings"
    "math"
    "time"
)
import (
    "github.com/FAH2S/illusory-exchange-of-scarlet-fortune/types"
)


//{{{ Create user
func TestCreateUser(t *testing.T) {
    // helper FN
    strPtr := func(s string) *string { return &s }

    // TODO: change expectErr to expectErrSubStr for more precise
    tests := []struct {
        name            string
        user            types.User
        expectErr       bool
        expectErrStr    string
    }{
        {
            name:           "SuccUsernameAndSalt",
            user:           types.User{
                Username:       "test_user",
                Salt:       strPtr("0987654321abcdef"),
            },
            expectErr:      false,
        }, {
            name:           "FailDuplicateUsername",
            user:           types.User{
                Username:   "test_user",
            },
            expectErr:      true,
            expectErrStr:   "duplicate key value violates unique constraint",
        }, {
            name:           "FailSaltViolatesConstraint",
            user:           types.User{
                Username:   "test_user_2",
                Salt:       strPtr("fyhjdv0987654321abcdef"),
            },
            expectErr:      true,
            expectErrStr:   "violates check constraint",
        }, {
            name:           "FailEmptyUsername",
            user:           types.User{
                Username:   "",
            },
            expectErr:      true,
            expectErrStr:   "username is required",
        },
    }
    // Iterate
    for _, tc := range tests{
        t.Run(tc.name, func(t *testing.T) {
            err := CreateUser(DB, tc.user)
            // When error is not nil BUT we don't expect error
            if (err != nil) != tc.expectErr {
                t.Errorf("Error that is not expected occured: %v", err)
            }
            if tc.expectErr {
                if !strings.Contains(err.Error(), tc.expectErrStr){
                    t.Errorf("Wrong error\nExpected:\t%q\nGot:\t\t%q", tc.expectErrStr, err)
                }
            }
        })
    }
}
//}}} Create user


//{{{ Read user
func TestReadUser(t *testing.T) {
    user := types.User { Username: "test_user_to_be_read" }
    err := CreateUser(DB, user)
    if err != nil {
        t.Fatalf("Could not create user: %v", err)
    }
    tests := []struct {
        name            string
        username        string
        expectErr       bool
        expectUsr       types.User
    }{
        {
            name:           "Succ",
            username:       "test_user_to_be_read",
            expectErr:      false,
            expectUsr:      types.User{
                Username:"test_user_to_be_read",
            },
        }, {
            name:           "UserNotFound",
            username:       "dose_not_exist",
            expectErr:      true,
        },
    }
    // Iterate
    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            readUser, err := ReadUser(DB, tc.username)
            if (err != nil) != tc.expectErr {
                t.Errorf("Error that is not expected occured: %v", err)
            }
            if err == nil {
                if !reflect.DeepEqual(tc.expectUsr, *readUser) {
                    t.Errorf("Users not the same\nExpected:\t%v\nGot:\t\t%v", tc.expectUsr, *readUser)
                }
            }
        })
    }
}
//}}} Read user


//{{{ Create OrderFill
func TestCreateOrderFill(t *testing.T){ // init
    user := types.User{ Username: "test_user_for_order_fill" }
    orderFill := types.OrderFill{
        FillId:         "b786f148-6621-4fe8-bda6-e24b9291de87",
        Symbol:         "PF_XRPUSD",
        Side:           "buy",
        Price:          2.89278,
        CoinAmount:     10.0,
        Coin:           "XRP",
        CurrencyAmount: 28.9278,
        Currency:       "dollar",
        FillType:       "taker",
        DateTime:       "2025-09-23T16:55:10.557Z",
        Owner:          "test_user",
    }
    err := CreateUser(DB, user)
    if err != nil {
        t.Fatalf("Failed to create user: %v", err)
    }

    tests := []struct {
        name            string
        orderFill       types.OrderFill
        ignoreFlag      bool
        expectErr       bool
        expectErrStr    string
    }{
        {
            name:           "Succ",
            orderFill:      orderFill,
            ignoreFlag:     false,
            expectErr:      false,
            expectErrStr:   "",
        }, {
            name:           "FailOrderAlreadyExist",
            orderFill:      orderFill,
            ignoreFlag:     false,
            expectErr:      true,
            expectErrStr:   "duplicate key value violates unique constraint",
        }, {
            name:           "SuccIgnoreFlag",
            orderFill:      orderFill,
            ignoreFlag:     true,
            expectErr:      false,
            expectErrStr:   "duplicate key value violates unique constraint",
        },
    }

    // Iterate
    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            err := CreateOrderFill(DB, tc.orderFill, tc.ignoreFlag)
            if (err != nil) != tc.expectErr {
                t.Errorf("Error that is not expected occured: %v", err)
            }
            if tc.expectErr {
                if err == nil {
                     t.Errorf("Expected error %q, but got nil", tc.expectErrStr)
                 } else if !strings.Contains(err.Error(), tc.expectErrStr){
                    t.Errorf("Wrong error\nExpected:\t%q\nGot:\t\t%q", tc.expectErrStr, err)
                }
            }
        })
    }
}
//}}} Create OrderFill


//{{{ Read AvgPrice
// DISCLAMER can't test last X ranges becuse over time it will give false positive
//  when time passing makes typed range obsolete
// TODO: when initializing dateTime for orderfills make it dynamic based on current date if that is important
// helper fn, for initializing dates (for range)
func pastDate(months int) string {
    return time.Now().UTC().AddDate(0, months, 0).Format(time.RFC3339Nano)
}
func TestReadAvgPrice(t *testing.T) {
    //{{{ init orderFill-s
    user := types.User{ Username: "test_user_for_avg_entry" }
    err := CreateUser(DB, user)
    if err != nil {
        t.Fatalf("Failed to create user: %v", err)
    }
    orderFill := types.OrderFill{
        FillId:         "b786f148-6621-4fe8-bda6-e24b9291de87",
        Symbol:         "PF_BCHUSD",
        Side:           "buy",
        Price:          550.23,
        CoinAmount:     1,
        Coin:           "BCH",
        CurrencyAmount: 550.23,
        Currency:       "dollar",
        FillType:       "taker",
        DateTime:       pastDate(-1),
        Owner:          "test_user_for_avg_entry",
    }
    var ofList []types.OrderFill
    of1 := orderFill
    of1.FillId = "b186f148-1621-1fe8-bda6-e14b9291de17"
    of1.DateTime = pastDate(-3)
    of1.Side = "sell"
    of1.Price = 620.0
    of1.CoinAmount = 0.7
    of1.CurrencyAmount = 434.0
    ofList = append(ofList, of1)

    of2 := orderFill
    of2.DateTime = pastDate(-6)
    of2.FillId = "b286f228-2622-2fe8-bda6-e22b9292de27"
    of2.Side = "sell"
    of2.Price = 715.0
    of2.CoinAmount = 0.3
    of2.CurrencyAmount = 214.5
    ofList = append(ofList, of2)

    of3 := orderFill
    of3.DateTime = pastDate(-9)
    of3.FillId = "b386f348-3623-3fe8-bda6-e34b9293de37"
    of3.Side = "buy"
    of3.Price = 220.12
    of3.CoinAmount = 1.1
    of3.CurrencyAmount = 242.132
    ofList = append(ofList, of3)

    of4 := orderFill
    of4.DateTime = pastDate(-12)
    of4.FillId = "b486f448-4624-4fe8-bda6-e44b9294de47"
    of4.Side = "buy"
    of4.Price = 154.23
    of4.CoinAmount = 0.23
    of4.CurrencyAmount = 35.4729
    ofList = append(ofList, of4)

    for _, of := range(ofList){
        err := CreateOrderFill(DB, of, false)
        if err != nil {
            t.Fatalf("Failed to insert orderFill: %v", err)
        }
    }
    //}}} init orderFill-s
    tests := []struct {
        name            string
        owner           string
        coin            string
        side            string
        currency        string
        dayRange        int
        expectAvgEntry  float64
        expectVolume    float64
    }{
        {
            name:           "SuccBuy3600",
            owner:          "test_user_for_avg_entry",
            coin:           "BCH",
            side:           "buy",
            currency:       "dollar",
            dayRange:       3600,
            expectAvgEntry: 211.7004,
            expectVolume:   277.6049,
        }, {
            name:           "SuccBuy290",
            owner:          "test_user_for_avg_entry",
            coin:           "BCH",
            side:           "buy",
            currency:       "dollar",
            dayRange:       290,
            expectAvgEntry: 220.12,
            expectVolume:   242.132,
        }, {
            name:           "FailNoCoin",
            owner:          "test_user_for_avg_entry",
            coin:           "NULL",
            side:           "buy",
            currency:       "dollar",
            dayRange:       360,
            expectAvgEntry: 0,
            expectVolume:   0,
        }, {
            name:           "SuccSell3600",
            owner:          "test_user_for_avg_entry",
            coin:           "BCH",
            side:           "sell",
            currency:       "dollar",
            dayRange:       3600,
            expectAvgEntry: 651.4225,
            expectVolume:   648.5,
        }, {
            name:           "SuccSell110",
            owner:          "test_user_for_avg_entry",
            coin:           "BCH",
            side:           "sell",
            currency:       "dollar",
            dayRange:       110,
            expectAvgEntry: 620.0,
            expectVolume:   434.0,
        },
    }
    // Iterate
    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            avgEntry, volume, err := ReadAvgPrice(DB, tc.owner, tc.coin, tc.side, tc.currency, tc.dayRange)
            t.Log(err)
            if math.Round(avgEntry *1000) != math.Round(tc.expectAvgEntry *1000) {
                t.Errorf("avgEntry not the same\nExpected:\t%.4f\nGot:\t\t%.4f", tc.expectAvgEntry, avgEntry)
            }
            if math.Round(volume *1000) != math.Round(tc.expectVolume *1000) {
                t.Errorf("Volume not the same\nExpected:\t%.4f\nGot:\t\t%.4f", tc.expectVolume, volume)
            }
        })
    }
}



//}}} Read AvgPrice


