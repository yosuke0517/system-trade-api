
-- +migrate Up
CREATE TABLE IF NOT EXISTS `fx_btc_jpy1m` (
  `time` DATETIME PRIMARY KEY NOT NULL,
  `open` float,
  `close` float,
  `high` float,
  `low` float,
  `volume` float
);
-- +migrate Down
DROP TABLE IF EXISTS `fx_btc_jpy1m`;