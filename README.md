# system-trade-api
### システムトレードアプリのAPI
- 現在所持している現金やビットコインの情報を取得する：`GetBalance`
- ビットコインの情報（現在の価格等）を取得する：`GetTicker`

# SETUP
- アプリ起動
  - `docker-compose up`
  - 下図のようにデバッグ設定を追加
- godoc
  - コンテナに入る必要有り
  - `godoc -http=:6060`