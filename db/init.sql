CREATE TABLE IF NOT EXISTS users (
    username        VARCHAR(32) UNIQUE NOT NULL,
    salt            VARCHAR(64),    -- hex string
    enc_pub_key     VARCHAR(256),   -- hex string
    enc_priv_key    VARCHAR(256),   -- hex string

    CHECK (salt ~           '^[0-9a-fA-F]+$'),
    CHECK (enc_pub_key ~    '^[0-9a-fA-F]+$'),
    CHECK (enc_priv_key ~   '^[0-9a-fA-F]+$')
);

CREATE TABLE IF NOT EXISTS order_fills (
    fill_id         VARCHAR(256) UNIQUE NOT NULL,
    symbol          VARCHAR(32) NOT NULL,
    side            VARCHAR(4) NOT NULL CHECK (side in ('buy', 'sell')),
    price           DECIMAL(16, 8) NOT NULL,
    coin_amount     DECIMAL(16, 8) NOT NULL,
    coin            VARCHAR(32) NOT NULL,
    currency_amount DECIMAL(12, 4) NOT NULL,
    currency        VARCHAR(32) NOT NULL,
    fill_type       VARCHAR(32) NOT NULL,
    date_time       TIMESTAMPTZ NOT NULL,
    owner           VARCHAR(32) NOT NULL,
    FOREIGN KEY (owner) REFERENCES users(username) ON DELETE CASCADE
);

