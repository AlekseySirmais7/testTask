DROP DATABASE myService;

CREATE DATABASE myService
    WITH
    OWNER = postgres
    CONNECTION LIMIT = -1
    ;

CREATE COLLATION posix (LOCALE = 'POSIX');
CREATE EXTENSION citext;

CREATE unlogged TABLE Products (
    offer_id int NOT NULL CHECK (offer_id > -1),
    seller_id int NOT NULL CHECK (seller_id > -1),
    name text NOT NULL,
    price int NOT NULL CHECK (price > -1), --- better use money type, check xlsx format for money
    quantity int NOT NULL,
    available boolean NOT NULL,
    UNIQUE (offer_id, seller_id)
);

CREATE INDEX Products_offer_id ON Products (offer_id);
CREATE INDEX Products_seller_id ON Products (seller_id);

CREATE  INDEX Select_products ON Products (seller_id, offer_id, LOWER(name));

CREATE INDEX Products_name ON Products (LOWER(name));

CREATE INDEX Gin_name
ON Products
USING gin (to_tsvector('english', "name"));
