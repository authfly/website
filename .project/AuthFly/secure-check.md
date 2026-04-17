## Вердикт

Как **узкий IdP для controlled environment** эта реализация выглядит **рабочей, но заметно менее безопасной и менее protocol-hardened**, чем **Keycloak** и **Zitadel**. По `go.mod` видно главное: тут почти нет готового security/protocol слоя, а SAML/OIDC логика написана вручную. Для IdP это важнее, чем «малый размер кода».

Если коротко: **для внутреннего стенда или небольшой закрытой схемы с фиксированными SP/client allowlist** я бы это назвал «осторожно приемлемо». Для **internet-facing production IdP** уровень доверия у меня был бы **существенно ниже, чем у Keycloak/Zitadel**.

## Основные риски

1. **Hanko JWT доверяется слишком широко.**  
В `internal/auth/jwt.go` проверяются подпись и `exp`, но я не вижу проверки `iss`, `aud`, `nbf`, `azp` или привязки токена к конкретному intended use. Для IdP это важный момент: если Hanko подписывает разные типы токенов одним JWKS, этот сервер рискует принять «чужой, но валидно подписанный» JWT.

```80:130:e:\_@Go\@SSO\internal\auth\jwt.go
func (v *JWTVerifier) VerifyToken(tokenStr string) (*HankoClaims, error) {
	parts := strings.Split(tokenStr, ".")
	// ...
	pubKey, err := v.getKey(header.Kid)
	// ...
	if err := rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hash[:], sigBytes); err != nil {
		return nil, fmt.Errorf("invalid signature: %w", err)
	}
	// ...
	if claims.Exp > 0 && time.Now().Unix() > claims.Exp {
		return nil, fmt.Errorf("token expired")
	}
	return &claims, nil
}
```

2. **Токен можно принять из query string.**  
`ExtractHankoToken()` принимает `?token=...`. Для bearer token это плохая практика: утечки через логи, history, referer, reverse proxy и чужие analytics. Для IdP это особенно неприятно, потому что таким токеном создаётся сессия.

```175:185:e:\_@Go\@SSO\internal\auth\session.go
func ExtractHankoToken(r *http.Request) string {
	if auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	if cookie, err := r.Cookie("hanko"); err == nil {
		return cookie.Value
	}
	if token := r.URL.Query().Get("token"); token != "" {
		return token
	}
	return ""
}
```

3. **SAML AuthnRequest почти не валидируется, а metadata прямо говорит `WantAuthnRequestsSigned="false"`.**  
Сейчас запрос просто распаковывается, парсится и матчится по `Issuer`. Нет проверки подписи AuthnRequest, `Destination`, `IssueInstant`, anti-replay, request ID cache. Это упрощает жизнь, но делает SAML-часть гораздо слабее, чем у Keycloak/Zitadel.

```30:63:e:\_@Go\@SSO\internal\saml\authnrequest.go
func ParseAuthnRequest(r *http.Request, cfg *config.Config) (*ParsedRequest, error) {
	raw := r.URL.Query().Get("SAMLRequest")
	// ...
	var req AuthnRequest
	if err := xml.Unmarshal(xmlBytes, &req); err != nil {
		return nil, fmt.Errorf("xml parse: %w", err)
	}
	sp := cfg.FindSP(req.Issuer.Value)
	if sp == nil {
		return nil, fmt.Errorf("unknown SP: %s", req.Issuer.Value)
	}
	return &ParsedRequest{
		AuthnRequest: req,
		SP:           sp,
		RelayState:   relayState,
	}, nil
}
```

```15:29:e:\_@Go\@SSO\internal\saml\metadata.go
<IDPSSODescriptor
    WantAuthnRequestsSigned="false"
    protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">
```

4. **OIDC реализован как минимальный confidential-client flow без PKCE.**  
Это допустимо только для строго серверных confidential clients. На уровне IdP это значит: по сравнению с Keycloak/Zitadel тут нет современного hardening для public/native/SPA сценариев.

5. **Кастомная реализация XMLDSig/C14N и JWT.**  
Самый большой архитектурный минус не в одной строке, а в подходе: `internal/saml/xmldsig.go` и `internal/oidc/jwt.go` пишут критичную протокольную логику вручную. В SAML это особенно опасно, потому что XML canonicalization и signature validation исторически очень «ломкие». У Keycloak/Zitadel этот слой давно обкатан.

6. **Встроенная admin-интеграция слабее enterprise IdP уровня.**  
`internal/admin/handlers.go` вызывает Hanko Admin API без отдельной auth-логики в этом сервисе и создаёт пользователя сразу с `IsVerified: true`. Это может быть нормой для закрытой админки за приватной сетью, но это не уровень встроенного admin security из Keycloak/Zitadel.

## Что здесь сделано неплохо

- В OIDC есть **строгое сравнение `redirect_uri`** по allowlist в `internal/config/config.go`.
- Authorization code **одноразовый** и с TTL в `internal/oidc/code_store.go`.
- SAML response уходит на **ACS из конфига SP**, а не на адрес из входящего запроса, это хороший защитный выбор.
- Cookie сессии подписаны, `HttpOnly`, `Secure`.
- JWT/JWKS и SAML signing в целом есть, то есть это не «совсем игрушка».

## Сравнение с Keycloak и Zitadel

Если сравнивать **именно как Identity Provider**, а не «в общем»:

- **Keycloak/Zitadel заметно безопаснее** по протокольной зрелости.
- Их преимущество не только в количестве фич, а в том, что у них уже есть:
  - строгая валидация claims и request context;
  - key rotation и lifecycle;
  - PKCE/prompt/max_age/logout/session management;
  - более зрелая SAML request/response processing;
  - более сильная admin/security модель;
  - более предсказуемое поведение в edge cases и under attack.

Эта реализация выглядит как **хороший компактный bridge IdP**, но **не как полноценная hardened IdP-платформа**. То есть против Keycloak/Zitadel я бы поставил её на уровень:

- **закрытый B2B/internal use**: можно довести до приемлемого состояния;
- **публичный продовый IdP**: сейчас нет, слишком много protocol-hardening сделано «вручную и по минимуму».

Если хотите, я могу следующим сообщением сделать уже совсем практичный список: **что именно нужно усилить в этом репозитории, чтобы он приблизился к “минимально безопасному production IdP” для SAML и OIDC**.

======================

Прочитал **каждый файл** в репозитории. Теперь могу дать исчерпывающий, построчный аудит, с подтверждением или опровержением каждого пункта из прошлого сообщения, плюс **новые находки**.

---

## Часть 1: Перепроверка предыдущего аудита

### Пункт 1 — «Hanko JWT доверяется слишком широко»

**ПОДТВЕРЖДАЮ.** Перечитал `jwt.go` полностью — после верификации подписи и `exp` никакие другие claims не проверяются:

```80:130:e:\_@Go\@SSO\internal\auth\jwt.go
func (v *JWTVerifier) VerifyToken(tokenStr string) (*HankoClaims, error) {
	// ...
	if err := rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hash[:], sigBytes); err != nil {
		return nil, fmt.Errorf("invalid signature: %w", err)
	}
	// ...
	if claims.Exp > 0 && time.Now().Unix() > claims.Exp {
		return nil, fmt.Errorf("token expired")
	}
	return &claims, nil
}
```

Не проверяются: `iss` (issuer), `aud` (audience), `nbf` (not before), а также **алгоритм из заголовка** (algorithm confusion attack — если Hanko когда-то начнёт выдавать токены с другим `alg`, этот код не откажет, а молча не совпадёт подпись, что хорошо; но если ключ RSA переиспользуется, отсутствие проверки `iss`/`aud` становится вектором).

### Пункт 2 — «Токен можно принять из query string»

**ПОДТВЕРЖДАЮ.**

```175:185:e:\_@Go\@SSO\internal\auth\session.go
func ExtractHankoToken(r *http.Request) string {
	// ...
	if token := r.URL.Query().Get("token"); token != "" {
		return token
	}
	return ""
}
```

Токен в URL попадает в: browser history, referer header, access-логи reverse proxy и CDN, browser extensions. Для IdP это создаёт канал утечки.

### Пункт 3 — «SAML AuthnRequest почти не валидируется»

**ПОДТВЕРЖДАЮ, но с важной оговоркой в плюс.** Не проверяются: подпись AuthnRequest, `Destination`, `IssueInstant`, нет anti-replay cache. **НО** при построении SAML Response `ACSUrl` берётся из **конфига SP** (`req.SP.ACSUrl`), а не из поля `AssertionConsumerServiceURL` из входящего XML. Это **хорошее решение**, которое закрывает атаку с подменой ACS URL:

```39:40:e:\_@Go\@SSO\internal\saml\response.go
	resp.CreateAttr("Destination", req.SP.ACSUrl)
	// ...
	subjectConfData.CreateAttr("Recipient", req.SP.ACSUrl)
```

А SP lookup идёт по `Issuer`, который матчится по whitelisted config:

```54:57:e:\_@Go\@SSO\internal\saml\authnrequest.go
	sp := cfg.FindSP(req.Issuer.Value)
	if sp == nil {
		return nil, fmt.Errorf("unknown SP: %s", req.Issuer.Value)
	}
```

Так что **уязвимости с подменой ACS нет**, и это важная коррекция к моему прошлому анализу. Проблема скорее в отсутствии anti-replay и validation of Destination, что даёт возможность relay-атак.

### Пункт 4 — «OIDC реализован без PKCE»

**ПОДТВЕРЖДАЮ.** В `HandleAuthorize` нет обработки `code_challenge` / `code_challenge_method`. Discovery тоже не объявляет поддержку PKCE:

```41:46:e:\_@Go\@SSO\internal\oidc\handlers.go
		"response_types_supported": []string{"code"},
		// ... no code_challenge_methods_supported
		"grant_types_supported":  []string{"authorization_code"},
		"token_endpoint_auth_methods_supported": []string{"client_secret_post", "client_secret_basic"},
```

### Пункт 5 — «Кастомная C14N и JWT»

**ПОДТВЕРЖДАЮ, но с нюансом.** Кастомная C14N (`xmldsig.go`, 332 строки) реализует Exclusive XML Canonicalization с корректной сортировкой namespace и атрибутов. Код не тривиальный и **выглядит рабочим** по результатам тестов, но **не покрыт edge case тестами** — единственный тест `TestSignAssertion` проверяет лишь наличие элементов, не валидирует подпись обратно.

### Пункт 6 — «Admin интеграция слабее enterprise»

**ПОДТВЕРЖДАЮ и дополняю** (см. новые находки ниже).

---

## Часть 2: Новые находки

Вот что я обнаружил при полном аудите, чего **не было** в прошлом сообщении.

### CRITICAL — HTTP-сервер без таймаутов

```88:89:e:\_@Go\@SSO\cmd\server\main.go
	if err := http.ListenAndServe(cfg.ListenAddr, mux); err != nil {
		log.Fatalf("Server error: %v", err)
	}
```

`http.ListenAndServe` создаёт сервер **без ReadTimeout, WriteTimeout, IdleTimeout**. Это делает сервер уязвимым к **Slowloris** и аналогичным DoS-атакам: атакующий открывает сотни соединений и держит их, исчерпывая горутины.

### CRITICAL — session_key не валидируется при старте

```8:8:e:\_@Go\@SSO\config.yaml
session_key: "CHANGE-ME-TO-A-RANDOM-32-BYTE-KEY"
```

Нигде в коде нет проверки, что `session_key` был изменён с дефолтного значения. Если кто-то задеплоит без `.env`, все HMAC-подписи будут с известным ключом — **полный обход аутентификации**.

### HIGH — Dockerfile: контейнер запускается от root

```20:30:e:\_@Go\@SSO\Dockerfile
FROM alpine:3.21
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /idp .
# ... no USER directive
CMD ["./idp"]
```

Нет `USER` / `adduser`. Процесс работает от root внутри контейнера.

### HIGH — Права на private key файл

```89:96:e:\_@Go\@SSO\internal\auth\crypto.go
	keyFile, err := os.Create(keyPath)
	// ...
	if err := pem.Encode(keyFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}); err != nil {
```

`os.Create` создаёт файл с правами 0666 (до umask). Приватный ключ IdP должен создаваться с правами **0600**.

### HIGH — CodeStore: нет garbage collection

```49:58:e:\_@Go\@SSO\internal\oidc\code_store.go
func (s *CodeStore) Cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for k, v := range s.codes {
		if now.After(v.ExpiresAt) {
			delete(s.codes, k)
		}
	}
}
```

Метод `Cleanup()` существует, но **нигде не вызывается** — ни в `main.go`, ни в goroutine. Истёкшие коды авторизации накапливаются в памяти бесконечно. Для целенаправленного DoS достаточно генерировать OIDC-потоки без завершения.

### HIGH — Нет CSRF-защиты на admin формах

```17:17:e:\_@Go\@SSO\internal\web\views\admin_console.templ
			<form method="POST" action="/console/users" class="flex gap-3 items-end">
```

POST `/console/users` не защищён CSRF-токеном. Атакующий может разместить скрытую форму на своём сайте, и если админ зайдёт на эту страницу с активной IdP-сессией, будет создан произвольный пользователь.

### HIGH — verifyAccessToken не проверяет issuer

```216:248:e:\_@Go\@SSO\internal\oidc\handlers.go
func (h *Handlers) verifyAccessToken(tokenStr string) (*AccessTokenClaims, error) {
	// ...parses JWT manually...
	if claims.Exp > 0 && time.Now().Unix() > claims.Exp {
		return nil, http.ErrNoCookie
	}
	// ...verifies RSA signature...
	return &claims, nil
}
```

Нет проверки `claims.Iss` — если этот же RSA-ключ используется в другом контексте (или если access_token от другого IdP с тем же ключом), он будет принят.

### HIGH — id_token не содержит at_hash

```47:58:e:\_@Go\@SSO\internal\oidc\jwt.go
func GenerateIDToken(kp *auth.IdPKeyPair, issuer, audience, sub, email, nonce string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := idTokenClaims{
		Iss:   issuer,
		Sub:   sub,
		Aud:   audience,
		Exp:   now.Add(ttl).Unix(),
		Iat:   now.Unix(),
		Email: email,
		Nonce: nonce,
	}
```

По OIDC Core Section 3.2.2.9, когда `id_token` выдаётся вместе с `access_token` из token endpoint, он **MUST** содержать `at_hash`. Без этого клиент не может верифицировать привязку `access_token` к `id_token`.

### MEDIUM — SameSite=None на IdP-сессии

```49:57:e:\_@Go\@SSO\internal\auth\session.go
	http.SetCookie(w, &http.Cookie{
		Name:     IdpSessionCookie,
		Value:    val,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		MaxAge:   int(SessionDuration.Seconds()),
	})
```

`SameSite=None` необходим для cross-origin SSO (SAML POST binding, OIDC redirect), но ослабляет защиту от CSRF на всех эндпоинтах IdP. В сочетании с отсутствием CSRF-токенов на admin формах — это усиливает предыдущую находку.

### MEDIUM — Нет security headers

Нигде в коде нет middleware для заголовков: `Content-Security-Policy`, `X-Frame-Options`, `X-Content-Type-Options`, `Strict-Transport-Security`, `Referrer-Policy`. Для IdP это важнее, чем для обычного сайта: login-страница может быть встроена через iframe (clickjacking) или стать целью XSS через injection в error messages.

### MEDIUM — RSA 2048-bit ключ

```65:65:e:\_@Go\@SSO\internal\auth\crypto.go
	key, err := rsa.GenerateKey(rand.Reader, 2048)
```

RSA 2048 формально ещё допустим, но для нового IdP рекомендуется **3072 или 4096 bit**. NIST рекомендует 3072+ для использования после 2030 года.

### MEDIUM — Сертификат на 10 лет без ротации

```75:75:e:\_@Go\@SSO\internal\auth\crypto.go
		NotAfter:     time.Now().Add(10 * 365 * 24 * time.Hour),
```

10-летний self-signed сертификат — нет механизма ротации ключей. Если ключ скомпрометирован, единственный вариант — ручная замена и обновление у всех SP/клиентов.

### MEDIUM — Logout не поддерживает SAML SLO

Logout (`HandleLogout` в `sso.go`) очищает IdP-сессию и Hanko-сессию, но **не отправляет LogoutRequest** подключённым SAML SP. Пользователь остаётся «залогинен» на сайте SP до истечения локальной сессии.

### LOW — Нет rate limiting

Ни на одном эндпоинте нет ограничения частоты запросов. `/token`, `/authorize`, `/sso` доступны для перебора.

### LOW — Нет graceful shutdown

```88:89:e:\_@Go\@SSO\cmd\server\main.go
	if err := http.ListenAndServe(cfg.ListenAddr, mux); err != nil {
```

Нет обработки `SIGTERM` / `context.WithCancel` / `server.Shutdown()`. При остановке контейнера текущие запросы обрываются.

### LOW — Нет аудит-логирования

Действия «пользователь вошёл», «код выдан», «токен обменян», «пользователь создан в admin» — логируются только в `log.Printf` без структуры и без возможности отправки в SIEM.

---

## Часть 3: Что реально хорошо

Чтобы картина была честной, вот что сделано **правильно** и что я могу подтвердить после полного аудита:

| Аспект | Оценка |
|--------|--------|
| ACS URL берётся из конфига SP, а не из AuthnRequest | Правильно, закрывает подмену ACS |
| redirect_uri — строгое exact match по allowlist | Правильно |
| Authorization code одноразовый + TTL 5 мин | Правильно |
| `secureCompare` через `hmac.Equal` | Правильно, timing-safe |
| Cookie подписаны HMAC-SHA256, `HttpOnly`, `Secure` | Правильно |
| SAML Response содержит все обязательные элементы (AudienceRestriction, SubjectConfirmation, NotOnOrAfter) | Правильно |
| XMLDSig: SHA-256 digest + RSA-SHA256 signature, enveloped-signature transform | Правильно |
| `Cache-Control: no-store` на `/token` | Правильно |
| Frontend SDK подключается из бандла, а не CDN | Правильно |
| Go `html/template` в `postform.html` авто-экранирует XSS | Правильно |
| templ-шаблоны авто-экранируют пользовательские данные | Правильно |
| `IsAllowedLogoutReturnURL` валидирует return_to по origin allowlist | Правильно |
| Docker: multi-stage build, alpine, pinned version | Правильно |
| Docker: listen `127.0.0.1:5800`, не exposed наружу | Правильно |

---

## Часть 4: Поэтапный план hardening

Упорядочено по приоритету: сначала то, что может привести к обходу аутентификации или DoS, потом — protocol compliance и enterprise-hardening.

### Этап 1 — Критические исправления (неделя 1)

| # | Что | Файл | Суть |
|---|-----|------|------|
| 1.1 | Проверка `iss` и `aud` в Hanko JWT | `internal/auth/jwt.go` | Добавить параметры `expectedIssuer`, `expectedAudience` в `JWTVerifier`, проверять их после парсинга claims |
| 1.2 | Валидация `alg` в JWT header | `internal/auth/jwt.go` | Отклонять любой `alg` кроме `RS256` |
| 1.3 | Убрать `?token=` из `ExtractHankoToken` | `internal/auth/session.go` | Оставить только `Authorization: Bearer` и `hanko` cookie |
| 1.4 | Runtime-проверка session_key | `cmd/server/main.go` | `log.Fatal` если `session_key` содержит `CHANGE-ME` или короче 32 байт |
| 1.5 | HTTP server timeouts | `cmd/server/main.go` | Заменить `http.ListenAndServe` на `&http.Server{ReadTimeout: 10s, WriteTimeout: 30s, IdleTimeout: 120s}` |
| 1.6 | Права на private key | `internal/auth/crypto.go` | Заменить `os.Create` на `os.OpenFile(path, O_CREATE|O_WRONLY, 0600)` |
| 1.7 | Dockerfile: non-root user | `Dockerfile` | Добавить `RUN adduser -D -H appuser` и `USER appuser` |

### Этап 2 — Protocol hardening (неделя 2-3)

| # | Что | Файл | Суть |
|---|-----|------|------|
| 2.1 | PKCE support | `internal/oidc/handlers.go`, `code_store.go` | Хранить `code_challenge` + `code_challenge_method` в `AuthCode`, проверять `code_verifier` в `/token`; объявить в Discovery |
| 2.2 | `at_hash` в id_token | `internal/oidc/jwt.go` | Вычислять SHA-256 от access_token, первые 128 бит → base64url → `at_hash` claim |
| 2.3 | Проверка `iss` в verifyAccessToken | `internal/oidc/handlers.go` | Сравнивать `claims.Iss` с `h.cfg.BaseURL` |
| 2.4 | SAML AuthnRequest validation | `internal/saml/authnrequest.go` | Проверять `IssueInstant` (отклонять старше 5 мин), проверять `Destination` если присутствует |
| 2.5 | SAML metadata: `WantAuthnRequestsSigned="true"` | `internal/saml/metadata.go` | Если SP шлёт подписанный AuthnRequest — проверять подпись (опционально, по конфигу SP) |
| 2.6 | CodeStore cleanup goroutine | `cmd/server/main.go` | Запустить `go func() { for { time.Sleep(1m); codeStore.Cleanup() } }()` |
| 2.7 | scope validation | `internal/oidc/handlers.go` | Проверять, что запрошенный scope содержит `openid`; отклонять неизвестные scopes |

### Этап 3 — Web security (неделя 3-4)

| # | Что | Файл | Суть |
|---|-----|------|------|
| 3.1 | Security headers middleware | новый `internal/middleware/security.go` | CSP, X-Frame-Options: DENY, X-Content-Type-Options: nosniff, Referrer-Policy: strict-origin, HSTS |
| 3.2 | CSRF-токены на admin формах | `internal/admin/` | Генерировать CSRF-токен в cookie, проверять в POST |
| 3.3 | Rate limiting | middleware или reverse proxy | Ограничение на `/token`, `/authorize`, `/sso`, `/console/users` |
| 3.4 | Request body size limits | middleware | `http.MaxBytesReader` на POST-эндпоинтах |
| 3.5 | Graceful shutdown | `cmd/server/main.go` | `signal.NotifyContext` + `server.Shutdown(ctx)` |

### Этап 4 — Enterprise features (неделя 4-6)

| # | Что | Файл | Суть |
|---|-----|------|------|
| 4.1 | Key rotation | `internal/auth/crypto.go`, `oidc/jwt.go` | Поддержка нескольких ключей в JWKS, выбор по `kid`, `previous_key` для валидации |
| 4.2 | RSA 4096 | `internal/auth/crypto.go` | При генерации новых ключей использовать 4096-bit |
| 4.3 | SAML Single Logout (SLO) | `internal/saml/`, `internal/handlers/sso.go` | Отправлять LogoutRequest всем активным SP при выходе |
| 4.4 | Structured audit logging | все handlers | JSON-структурированные логи: who, when, what, from where, result |
| 4.5 | Серверный session store | `internal/auth/session.go` | Опционально: хранить сессии в Redis/SQLite для revocation и forced logout |
| 4.6 | OIDC: refresh_token, prompt, max_age | `internal/oidc/` | Расширить flow до production-grade OIDC |
| 4.7 | Admin: CSRF + email validation | `internal/admin/handlers.go` | Не ставить `IsVerified: true` автоматически, или делать это конфигурируемым |
| 4.8 | Health/readiness endpoints | `cmd/server/main.go` | `GET /healthz`, `GET /readyz` для оркестратора |

---

**Итого по уровням:** 7 critical, 7 protocol-hardening, 5 web-security, 8 enterprise. Этап 1 закрывает реальные уязвимости и должен быть сделан до любого production-развёртывания. Этапы 2-3 приближают к уровню «безопасный production IdP». Этап 4 — до уровня, сопоставимого с enterprise-решениями.