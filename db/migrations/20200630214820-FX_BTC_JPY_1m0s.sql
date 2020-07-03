
-- +migrate Up
CREATE TABLE IF NOT EXISTS `FX_BTC_JPY_1m0s` (
  `time` TIMESTAMP PRIMARY KEY NOT NULL,
  `open` float,
  `close` float,
  `high` float,
  `low` float,
  `volume` float
);
-- +migrate Down
DROP TABLE IF EXISTS `FX_BTC_JPY_1m0s`;