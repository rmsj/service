-- Version: 1.01
-- Description: Create table users
CREATE TABLE users
(
    user_id       UUID        NOT NULL,
    name          TEXT        NOT NULL,
    email         TEXT UNIQUE NOT NULL,
    mobile        TEXT NULL,
    profile_image TEXT NULL,
    roles         TEXT[]      NOT NULL,
    password_hash TEXT        NOT NULL,
    department    TEXT NULL,
    enabled       BOOLEAN     NOT NULL,
    refresh_token TEXT NULL,
    date_created  TIMESTAMP   NOT NULL,
    date_updated  TIMESTAMP   NOT NULL,

    PRIMARY KEY (user_id)
);

-- Version: 1.02
-- Description: Create table products
CREATE TABLE products
(
    product_id   UUID           NOT NULL,
    user_id      UUID           NOT NULL,
    name         TEXT           NOT NULL,
    cost         NUMERIC(10, 2) NOT NULL,
    quantity     INT            NOT NULL,
    date_created TIMESTAMP      NOT NULL,
    date_updated TIMESTAMP      NOT NULL,

    PRIMARY KEY (product_id),
    FOREIGN KEY (user_id) REFERENCES users (user_id) ON DELETE CASCADE
);

-- Version: 1.03
-- Description: Add products view.
CREATE
OR REPLACE VIEW view_products AS
SELECT p.product_id,
       p.user_id,
       p.name,
       p.cost,
       p.quantity,
       p.date_created,
       p.date_updated,
       u.name AS user_name
FROM products AS p
         JOIN
     users AS u ON u.user_id = p.user_id

-- Version: 1.04
-- Description: Create table homes
CREATE TABLE homes
(
    home_id      UUID      NOT NULL,
    type         TEXT      NOT NULL,
    user_id      UUID      NOT NULL,
    address_1    TEXT      NOT NULL,
    address_2    TEXT NULL,
    zip_code     TEXT      NOT NULL,
    city         TEXT      NOT NULL,
    state        TEXT      NOT NULL,
    country      TEXT      NOT NULL,
    date_created TIMESTAMP NOT NULL,
    date_updated TIMESTAMP NOT NULL,

    PRIMARY KEY (home_id),
    FOREIGN KEY (user_id) REFERENCES users (user_id) ON DELETE CASCADE
);

-- Version: 1.05
-- Description: Add password reset table
CREATE TABLE password_reset_tokens
(
    email     TEXT      NOT NULL,
    token     TEXT      NOT NULL,
    expiry_at TIMESTAMP NOT NULL
);

-- Version: 1.06
-- Description: Add index to table password_reset_tokens
CREATE INDEX email_index ON password_reset_tokens (email);