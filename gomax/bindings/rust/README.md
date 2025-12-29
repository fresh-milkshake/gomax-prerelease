# gomax-rs

Rust-обёртка для Go [`gomax`](https://github.com/fresh-milkshake/gomax`), сборка через `build.rs` и `go build -buildmode=c-shared`.

**Требуется:** Go 1.22+ (`CGO_ENABLED=1` выставляется автоматом).

**Сборка:**  
```bash
cargo build
```
`build.rs` соберёт и подключит динамическую `libgomax.*`.

**Быстрый пример:**
```rust
use gomax_rs::Client;
let cfg = r#"{"Phone": "+79991234567", "WorkDir": "cache"}"#;
let mut c = Client::new(cfg)?;
c.on_message(|j| println!("msg: {j}"))?;
c.start(None)?;
let s = c.send_message(12345, "Привет!", true)?;
println!("{s:#?}");
println!("{:#?}", c.profile()?);
```
> Если нет токена, авторизация через stdin (или добавьте `Token` в конфиг).

**Возможности:**
- создание/закрытие клиента
- старт с таймаутом
- обработчик OnMessage (json)
- отправка сообщений, профиль, чаты (json)
