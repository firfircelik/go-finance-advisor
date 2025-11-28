## Hedef
Aequitas’ın backend’ini üretim seviyesine taşımak: güvenli kimlik doğrulama, konfigürasyon, kalıcı veritabanı şemaları, servis katmanı, sağlayıcı entegrasyonları, zamanlanmış işler, gözlemlenebilirlik, test ve dokümantasyon.

## Mimari ve Katmanlar
- HTTP API: net/http + handler’lar (mevcut yapıyı koru, route’ları organize et)
- Service Katmanı: iş kuralları (overview, allocation, holdings, auth, rebalancing)
- Repository Katmanı: DB erişimi için arayüzler (users, holdings, performance, ticker, liabilities)
- Provider Katmanı: dış veri kaynakları (ör. borsa/kripto/banka API’leri)

## Konfigürasyon
- Ortam değişkenleri: `AEQ_DB_URL`, `AEQ_JWT_SECRET`, `AEQ_PROVIDER_KEYS_*`, `AEQ_PORT`
- `.env` yükleme opsiyonu, varsayılanlarla güvenli fallback

## Güvenlik ve Auth
- Parola: `bcrypt/argon2id` hash + salt, minimum karmaşık parola politikası
- Token: JWT (HS256), `exp/iat/iss/sub` alanları, refresh token desteği
- Revocation/Blacklist: refresh token tablosu ve tek oturum sonlandırma
- Rate Limiting: IP/route bazlı, login denemelerinde artan gecikme
- CORS: üretimde domain whitelist, `Authorization` ve `Content-Type` dışında kısıtlı header

## Veritabanı ve Migrasyonlar
- SQLite başlangıç, PostgreSQL’e kolay geçiş için abstraction
- Migration aracı: `goose` veya `golang-migrate` ile versiyonlu migrasyonlar
- Şema (PostgreSQL uyumlu):
  - `users(id, email UNIQUE, password_hash, name, created_at)`
  - `holdings(id, user_id, asset, price, quantity, value, change_pct, icon, icon_color, category, updated_at)`
  - `performance(id, user_id, x, y, created_at)`
  - `ticker(symbol PRIMARY KEY, price, change, updated_at)`
  - `liabilities(id, user_id, name, amount, created_at)`
  - `refresh_tokens(id, user_id, token, expires_at, revoked)`
  - Index’ler: `holdings(user_id)`, `performance(user_id)`, `liabilities(user_id)`

## Endpoint’ler (Üretim)
- Auth: `POST /api/auth/signup`, `POST /api/auth/login`, `POST /api/auth/refresh`, `POST /api/auth/logout`, `GET /api/me`
- Portfolio: `GET /api/holdings`, `POST /api/holdings`, `PUT /api/holdings/:id`, `DELETE /api/holdings/:id`
- Overview: `GET /api/overview?range=1d|1w|1m|ytd|1y`
- Allocation: `GET /api/allocation`
- Performance: `GET /api/performance?range=...`, `POST /api/performance/import`
- Ticker: `GET /api/ticker` (REST) + `GET /api/ws/ticker` (WebSocket)
- Liabilities: `GET/POST/PUT/DELETE /api/liabilities`
- Rebalancing: `POST /api/rebalance/suggest` (hedef dağılıma göre öneri)
- Health: `GET /api/health`

## Sağlayıcı Entegrasyonları
- Piyasa Verileri: bir fiyat sağlayıcısı (örn. Polygon.io/AlphaVantage) için adapter
- Kripto: borsa fiyatları (örn. Binance/Coinbase) adapter’i
- Banka: Plaid gibi read-only bağlantı (mock → gerçek anahtarlarla değiştirilebilir)
- Provider anahtarları: env üzerinden, güvenli kullanım (retry, rate limit)

## Zamanlanmış İşler (Scheduler)
- Cron: `ticker` ve `performance` güncelleme (örn. her 1-5 dk)
- Günlük kapanışta `overview` snapshot ve PnL hesapları
- Job’lar için lock (tek çalıştırıcı), hata loglama ve retry backoff

## Gözlemlenebilirlik ve Hata Yönetimi
- Log: structured (zap/logrus), request-id, kullanıcı-id bağlamı
- Metrics: Prometheus endpoint (istek süresi, hata oranı, job metrikleri)
- Tracing: OpenTelemetry ile handler/service/provider zinciri
- Error Response: tutarlı JSON hata gövdesi (code, message)

## Performans ve Ölçekleme
- DB bağlantı havuzu, `PRAGMA` ayarları (SQLite) ve PostgreSQL hazırlığı
- N+1 kaçınma, toplu sorgular ve index’ler
- WebSocket: ticker için düşük gecikme yayın

## Test ve Kalite
- Unit test: service ve repository
- Integration test: in-memory SQLite/Postgres test DB
- Contract test: endpoint şemaları (OpenAPI) ve örnekler
- Load test: basit k6 senaryoları (ticker/overview)

## API Dokümantasyonu
- OpenAPI/Swagger JSON: `/api/docs`
- Endpoint örnekleri, şema tanımları, auth akışı

## Dağıtım ve Çevreler
- Dev: SQLite + env
- Prod: PostgreSQL + gizli anahtarlar (env/Secret Manager)
- CI: go vet, lint, test, build

## Geçiş Planı (Artımlı Uygulama)
1. Güvenlik: bcrypt/JWT/refresh + CORS daraltma
2. Migrations: goose ile versiyonlu şema; repository refactor
3. Overview/Allocation hesaplarını gerçek zamanlı ve kullanıcı-bazlı
4. Provider adapter’ları (fiyatlar) ve cron job’ları
5. WebSocket ticker ve Prometheus/OTel entegrasyonu
6. OpenAPI ve testler (unit/integration)

## Doğrulama
- `curl` ile canlı endpoint testleri
- Unit/Integration test çalıştırma
- Features-live sayfasında canlı statü ve önizleme kontrolü

Onay verirsen bu planı parça parça uygulamaya başlayacağım; önce güvenlik ve migrasyonlar, ardından veri hesapları ve sağlayıcı entegrasyonları.