# system-trade-api
### システムトレードアプリのAPI
- 現在所持している現金やビットコインの情報を取得する：`GetBalance`
- ビットコインの情報（現在の価格等）を取得する：`GetTicker`

# SETUP
- アプリ起動
  - `docker-compose up`
  - 下図のようにデバッグ設定を追加
![スクリーンショット 2020-06-14 10 15 39](https://user-images.githubusercontent.com/39196956/84582665-f70df280-ae29-11ea-9531-4580cdef853f.jpg)
- godoc
  - コンテナに入る必要有り
  - `godoc -http=:6060`