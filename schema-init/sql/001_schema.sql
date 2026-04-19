CREATE TABLE customers (
  customer_id VARCHAR(64) PRIMARY KEY,
  created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB;

CREATE TABLE balances (
  customer_id    VARCHAR(64) PRIMARY KEY,
  active_points  INT NOT NULL DEFAULT 0,
  pending_points INT NOT NULL DEFAULT 0,
  updated_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
               ON UPDATE CURRENT_TIMESTAMP,
  CONSTRAINT fk_balances_customer
    FOREIGN KEY (customer_id) REFERENCES customers(customer_id)
) ENGINE=InnoDB;

CREATE TABLE points_ledger (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  customer_id VARCHAR(64) NOT NULL,
  order_id    VARCHAR(64) NULL,
  reference   VARCHAR(64) NULL,
  return_id   VARCHAR(64) NULL,

  kind   ENUM('accrue','confirm','revoke','redeem','restore') NOT NULL,
  status ENUM('pending','active') NOT NULL DEFAULT 'pending',
  points INT NOT NULL,

  home24_merch_cents INT NULL,
  mirakl_merch_cents INT NULL,
  order_total_cents  INT NULL,
  shipping_cents     INT NULL,
  currency           CHAR(3) NULL,

  activated_at TIMESTAMP NULL,
  occurred_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  created_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

  idempotency_key VARCHAR(128) NULL,

  CONSTRAINT fk_ledger_customer
    FOREIGN KEY (customer_id) REFERENCES customers(customer_id),

  -- Uniqueness only when the column is actually provided
  UNIQUE KEY uq_ledger_order (customer_id, kind, order_id),
  UNIQUE KEY uq_ledger_ref   (customer_id, kind, reference),
  UNIQUE KEY uq_ledger_ret   (customer_id, kind, return_id),
  UNIQUE KEY uq_ledger_idem  (idempotency_key)
) ENGINE=InnoDB;


