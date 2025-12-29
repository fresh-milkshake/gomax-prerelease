# gomax-rs

Минимальные Rust биндинги к Go библиотеке [`github.com/fresh-milkshake/gomax`](https://github.com/fresh-milkshake/gomax). Сборка выполняется через `go build -buildmode=c-shared` (см. `build.rs`), затем Rust-обёртка подключает сгенерированный `libgomax.{so|dylib|dll}`.

## Требования

- Go 1.22+ установлен и доступен в `PATH`
- CGO включён (по умолчанию `build.rs` выставляет `CGO_ENABLED=1`)

## Сборка

```bash
cargo build
```

`build.rs` соберёт `libgomax.*` и заголовок в `$OUT_DIR`, добавит его в `rustc-link-search` и залинкует как динамическую библиотеку.

## Пример использования

```rust
use gomax_rs::Client;
use serde_json::json;
use std::time::Duration;

fn main() -> anyhow::Result<()> {
    // Конфигурация совпадает со структурой gomax.ClientConfig (JSON поля -> экспортируемые имена).
    let cfg = json!({
        "Phone": "+79991234567",
        "WorkDir": "cache",
        // При наличии готового токена можно передать его так:
        // "Token": "eyJhbGciOiJI..."
    })
    .to_string();

    let mut client = Client::new(&cfg)?;

    client.on_message(|msg_json| {
        println!("msg: {msg_json}");
    })?;

    // Таймаут передаётся в миллисекундах; None = без таймаута
    client.start(Some(Duration::from_secs(60)))?;

    // Отправка сообщения
    let sent = client.send_message(12345, "Привет из Rust!", true)?;
    println!("sent = {sent:#?}");

    // Профиль/чаты приходят как serde_json::Value
    let me = client.profile()?;
    println!("me = {me:#?}");
    let chats = client.chats()?;
    println!("chats total = {}", chats.as_array().map_or(0, |a| a.len()));

    Ok(())
}
```

> **Важно:** если токена нет, `gomax` запустит интерактивную авторизацию и запросит код из SMS через stdin. Для полностью безголовой авторизации передайте действующий `Token` в конфиге или расширьте FFI, добавив колбэк для `ClientConfig.CodeProvider`.

## Экспортируемые возможности

- Создание/закрытие клиента
- Старт клиента с таймаутом
- Обработчик `OnMessage` (колбэк получает JSON `types.Message`)
- `SendMessage`, получение профиля и списка чатов (JSON)

Если нужен больший охват API, добавьте новые функции в `ffi/bridge.go` и соответствующие обёртки в `src/lib.rs`. Основной паттерн: вернуть JSON строку + `gomax_free_string` для освобождения.
