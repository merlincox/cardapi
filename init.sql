CREATE TABLE IF NOT EXISTS customers (
  id       INT          NOT NULL AUTO_INCREMENT,
  fullname VARCHAR(256) NOT NULL,
  PRIMARY KEY (id)
)
  ENGINE = INNODB;

ALTER TABLE customers
  AUTO_INCREMENT = 1001;

CREATE TABLE IF NOT EXISTS vendors (
  id       INT          NOT NULL AUTO_INCREMENT,
  fullname VARCHAR(256) NOT NULL,
  PRIMARY KEY (id)
)
  ENGINE = INNODB;

ALTER TABLE vendors
  AUTO_INCREMENT = 1001;

CREATE TABLE IF NOT EXISTS cards (
  id          INT NOT NULL AUTO_INCREMENT,
  customer_id INT NOT NULL,
  balance     INT NOT NULL DEFAULT 0,
  available   INT NOT NULL DEFAULT 0,
  ts TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  INDEX card_customer_idx (customer_id),
  FOREIGN KEY (customer_id)
  REFERENCES customers (id)
    ON DELETE RESTRICT
    ON UPDATE RESTRICT
)
  ENGINE = INNODB;

ALTER TABLE cards
  AUTO_INCREMENT = 100001;

CREATE TABLE IF NOT EXISTS movements (
  id            INT          NOT NULL AUTO_INCREMENT,
  card_id       INT,
  movement_type          VARCHAR(256) NOT NULL,
  description   VARCHAR(256) NOT NULL,
  amount        INT          NOT NULL,
  ts TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  INDEX movement_card_idx (card_id),
  FOREIGN KEY (card_id)
  REFERENCES cards (id)
    ON DELETE RESTRICT
    ON UPDATE RESTRICT
)
  ENGINE = INNODB;

ALTER TABLE movements
  AUTO_INCREMENT = 1001;

CREATE TABLE IF NOT EXISTS authorisations (
  id                 INT          NOT NULL AUTO_INCREMENT,
  card_id            INT,
  vendor_id          INT,
  description        VARCHAR(256) NOT NULL,
  amount             INT          NOT NULL,
  captured           INT          NOT NULL DEFAULT 0,
  ts TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  INDEX authorisation_card_idx (card_id),
  INDEX authorisation_vendor_idx (vendor_id),
  FOREIGN KEY (card_id)
  REFERENCES cards (id)
    ON DELETE RESTRICT
    ON UPDATE RESTRICT,
  FOREIGN KEY (vendor_id)
  REFERENCES vendors (id)
    ON DELETE RESTRICT
    ON UPDATE RESTRICT
)
  ENGINE = INNODB;

ALTER TABLE authorisations
  AUTO_INCREMENT = 1001;

INSERT INTO customers (fullname)
VALUES ('John Smith'),('Jane Doe');

INSERT INTO vendors (fullname)
VALUES('Coffee Shop 1'),('Supermarket 1'),('Pub 1');

INSERT INTO cards (customer_id, balance, available)
VALUES(1001, 100000, 100000);