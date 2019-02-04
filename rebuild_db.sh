#!/usr/bin/env bash

cd "$( dirname "$0" )"

for cmd in "mysql"; do

    if [[ -z "$(which ${cmd})" ]]; then
        echo "${cmd} is required to run this script."  >&2
        exit 1
    fi

done

# The mysql.sh script contain DB connection details and should be git-ignored
mysql_script=mysql.sh

if [[ ! -f ${mysql_script} ]]; then
   echo "${mysql_script} is required to run this script."  >&2
   echo "See ${mysql_script}.example for an example."  >&2

   exit 1
fi

source ${mysql_script}

read -p "About to rebuild the ${mysql_db} database on ${mysql_host}, dropping and recreating all API tables. Confirm? :" -n 1 -r reply
echo
case "$reply" in
    y|Y ) echo "Proceeding with rebuild...";;
    * )   echo "Cancelling rebuild" >&2
          exit 1
          ;;
esac

mysql -h "${mysql_host}" -u "${mysql_user}" "-p${mysql_passwd}" "${mysql_db}" <<!!!

DROP TABLE IF EXISTS auth_movements;
DROP TABLE IF EXISTS movements;
DROP TABLE IF EXISTS authorisations;
DROP TABLE IF EXISTS cards;
DROP TABLE IF EXISTS customers;
DROP TABLE IF EXISTS vendors;

CREATE TABLE IF NOT EXISTS customers (
  id       INT          NOT NULL AUTO_INCREMENT,
  fullname VARCHAR(256) NOT NULL,
  PRIMARY KEY (id)
)
  ENGINE = INNODB;

ALTER TABLE customers
  AUTO_INCREMENT = 1001;

CREATE TABLE IF NOT EXISTS vendors (
  id          INT          NOT NULL AUTO_INCREMENT,
  vendor_name VARCHAR(256) NOT NULL,
  balance     INT NOT NULL DEFAULT 0,
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
  refunded           INT          NOT NULL DEFAULT 0,
  reversed           INT          NOT NULL DEFAULT 0,
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

CREATE TABLE IF NOT EXISTS auth_movements (
  id               INT          NOT NULL AUTO_INCREMENT,
  authorisation_id INT,
  movement_type    VARCHAR(256) NOT NULL,
  description      VARCHAR(256) NOT NULL,
  amount           INT          NOT NULL,
  ts               TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  INDEX movement_auth_idx (authorisation_id),
  FOREIGN KEY (authorisation_id)
  REFERENCES authorisations (id)
    ON DELETE RESTRICT
    ON UPDATE RESTRICT
)
  ENGINE = INNODB;

ALTER TABLE auth_movements
  AUTO_INCREMENT = 1001;

INSERT INTO customers (fullname)
VALUES ('John Smith'),('Jane Doe');

INSERT INTO vendors (vendor_name)
VALUES('Coffee Shop'),('Supermarket'),('Pub');

INSERT INTO cards (customer_id, balance, available)
VALUES(1001, 100000, 100000);

SHOW TABLES;

!!!