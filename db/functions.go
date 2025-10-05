package dbfns


import (
    "fmt"
    "errors"
    "database/sql"
    "os"
)
import (
    "github.com/FAH2S/illusory-exchange-of-scarlet-fortune/types"
)
import (
    "github.com/lib/pq"
)


func GetConnFromEnv() (*sql.DB, error) {
    // Extract values from env
    host := os.Getenv("DB_HOST")
    port := os.Getenv("DB_PORT")
    user := os.Getenv("DB_USER")
    password := os.Getenv("DB_PWD")
    name := os.Getenv("DB_NAME")
    if host == "" || port == "" || user == "" || password == "" || name == "" {
        return nil, fmt.Errorf("Missing one or more required DB environment variables")
    }

    // Create conn string
    connStr := fmt.Sprintf(
        "postgres://%s:%s@%s:%s/%s?sslmode=disable",
        user, password, host, port, name,
    )

    // Open connnections
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        return nil, fmt.Errorf("Failed to open connection: %v", err)
    }

    // Test connection
    if err = db.Ping(); err != nil {
        db.Close()
        return nil, fmt.Errorf("Failed to ping DB: %v", err)
    }

    return db, nil
}


// Potentially in future might need update enc_keys
func CreateUser(db *sql.DB, u types.User) error {
    if u.Username == "" {
        return fmt.Errorf("username is required")
    }
    query := `
        INSERT INTO users (username, salt, enc_pub_key, enc_priv_key)
        VALUES ($1, $2, $3, $4);
    `
    _, err := db.Exec(query, u.Username, u.Salt, u.EncPubKey, u.EncPrivKey)
    return err
}


func ReadUser(db *sql.DB, username string) (*types.User, error) {
    query := `SELECT username, salt, enc_pub_key, enc_priv_key FROM users WHERE username=$1;`
    row := db.QueryRow(query, username)

    var u types.User
    err := row.Scan(
        &u.Username,
        &u.Salt,
        &u.EncPubKey,
        &u.EncPrivKey,
    )
    if err != nil {
        return nil, err
    }
    return &u, nil
}


func CreateOrderFill(db *sql.DB, of types.OrderFill, ignoreFlag bool) error {
    query := `INSERT INTO order_fills(
        fill_id, symbol, side, price,
        coin_amount, coin, currency_amount, currency,
        fill_type, date_time, owner)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);`
    _, err := db.Exec(query,
        of.FillId, of.Symbol, of.Side, of.Price,
        of.CoinAmount, of.Coin, of.CurrencyAmount, of.Currency,
        of.FillType, of.DateTime, of.Owner)
    if err == nil {
        return nil
    }

    // To ignore 23505, easier to just bulk spam orders in DB
    if ignoreFlag {
        if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
            return nil
        }
    }

    return err
}


func ReadAvgPrice(db *sql.DB, owner, coin, side, currency string, dayRange int) (float64, float64, error) {
    // VWAP = SUM(price * volume)/SUM(volume)
    query := `
        SELECT SUM(price * currency_amount) / NULLIF(SUM(currency_amount), 0) AS avg_entry,
            NULLIF(SUM(currency_amount), 0) as volume
        FROM order_fills
        WHERE date_time >= NOW() - ($1 * INTERVAL '1 day')
            AND coin = $2 AND owner = $3 AND side = $4 AND currency = $5;
    `
    var avgEntry sql.NullFloat64
    var volume sql.NullFloat64
    err := db.QueryRow(query, dayRange, coin, owner, side, currency).Scan(&avgEntry, &volume)
    if err != nil {
        return 0, 0, err
    }
    if !avgEntry.Valid || !volume.Valid {
        return 0, 0, errors.New("No matching orders found")
    }
    return avgEntry.Float64, volume.Float64, nil
}












