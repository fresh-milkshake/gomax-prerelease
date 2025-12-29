//! Rust биндинги для Go библиотеки `github.com/fresh-milkshake/gomax`.
//! Оборачивает минимальный набор функций через C ABI, собранный в build.rs.

use serde_json::Value;
use std::ffi::{CStr, CString};
use std::os::raw::{c_char, c_int, c_longlong, c_void};
use std::sync::Mutex;
use std::time::Duration;

#[derive(Debug, Clone)]
pub struct GomaxError(pub String);

pub type Result<T> = std::result::Result<T, GomaxError>;

impl std::fmt::Display for GomaxError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        self.0.fmt(f)
    }
}

impl std::error::Error for GomaxError {}

// FFI declarations generated in ffi/bridge.go (header выводится go build'ом).
extern "C" {
    fn gomax_new_client(config_json: *const c_char, err_out: *mut *mut c_char) -> usize;
    fn gomax_start(handle: usize, timeout_ms: c_int, err_out: *mut *mut c_char) -> c_int;
    fn gomax_close(handle: usize);
    fn gomax_send_message(
        handle: usize,
        chat_id: c_longlong,
        text: *const c_char,
        notify: c_int,
        err_out: *mut *mut c_char,
    ) -> *mut c_char;
    fn gomax_get_profile(handle: usize, err_out: *mut *mut c_char) -> *mut c_char;
    fn gomax_get_chats(handle: usize, err_out: *mut *mut c_char) -> *mut c_char;
    fn gomax_set_on_message(
        handle: usize,
        cb: Option<extern "C" fn(*const c_char, *mut c_void)>,
        user_data: *mut c_void,
    );
    fn gomax_free_string(s: *mut c_char);
}

type OnMessageCb = Mutex<Box<dyn FnMut(String) + Send>>;

extern "C" fn on_message_trampoline(message_json: *const c_char, user_data: *mut c_void) {
    if user_data.is_null() || message_json.is_null() {
        return;
    }
    // Safety: pointers are provided by Go side; validated for null above.
    let message = unsafe { CStr::from_ptr(message_json) }.to_string_lossy().into_owned();
    let cb_ptr = user_data as *mut OnMessageCb;
    if let Ok(mut guard) = unsafe { (*cb_ptr).lock() } {
        (*guard)(message);
    }
}

pub struct Client {
    handle: usize,
    on_message_ud: Option<*mut OnMessageCb>,
}

impl Client {
    pub fn new(config_json: &str) -> Result<Self> {
        let c_json = CString::new(config_json).map_err(|e| GomaxError(e.to_string()))?;
        let mut err: *mut c_char = std::ptr::null_mut();
        let handle = unsafe { gomax_new_client(c_json.as_ptr(), &mut err) };
        unsafe {
            check_error(err)?;
        }
        if handle == 0 {
            return Err(GomaxError("gomax_new_client returned null handle".into()));
        }
        Ok(Self {
            handle,
            on_message_ud: None,
        })
    }

    pub fn start(&self, timeout: Option<Duration>) -> Result<()> {
        let mut err: *mut c_char = std::ptr::null_mut();
        let ms: i32 = timeout
            .map(|d| {
                d.as_millis()
                    .try_into()
                    .unwrap_or(i32::MAX as i128)
                    .min(i32::MAX as i128) as i32
            })
            .unwrap_or(0);
        let ok = unsafe { gomax_start(self.handle, ms as c_int, &mut err) };
        unsafe {
            check_error(err)?;
        }
        if ok == 0 {
            return Err(GomaxError("gomax_start returned failure".into()));
        }
        Ok(())
    }

    pub fn send_message(&self, chat_id: i64, text: &str, notify: bool) -> Result<Value> {
        let c_text = CString::new(text).map_err(|e| GomaxError(e.to_string()))?;
        let mut err: *mut c_char = std::ptr::null_mut();
        let ptr = unsafe {
            gomax_send_message(
                self.handle,
                chat_id as c_longlong,
                c_text.as_ptr(),
                if notify { 1 } else { 0 },
                &mut err,
            )
        };
        unsafe {
            check_error(err)?;
        }
        let json = unsafe { take_string(ptr)? };
        serde_json::from_str(&json).map_err(|e| GomaxError(e.to_string()))
    }

    pub fn profile(&self) -> Result<Option<Value>> {
        let mut err: *mut c_char = std::ptr::null_mut();
        let ptr = unsafe { gomax_get_profile(self.handle, &mut err) };
        unsafe {
            check_error(err)?;
        }
        if ptr.is_null() {
            return Ok(None);
        }
        let json = unsafe { take_string(ptr)? };
        serde_json::from_str(&json)
            .map(Some)
            .map_err(|e| GomaxError(e.to_string()))
    }

    pub fn chats(&self) -> Result<Value> {
        let mut err: *mut c_char = std::ptr::null_mut();
        let ptr = unsafe { gomax_get_chats(self.handle, &mut err) };
        unsafe {
            check_error(err)?;
        }
        let json = unsafe { take_string(ptr)? };
        serde_json::from_str(&json).map_err(|e| GomaxError(e.to_string()))
    }

    pub fn on_message<F>(&mut self, callback: F) -> Result<()>
    where
        F: FnMut(String) + Send + 'static,
    {
        let boxed: Box<OnMessageCb> = Box::new(Mutex::new(Box::new(callback)));
        let ptr = Box::into_raw(boxed);
        unsafe {
            gomax_set_on_message(self.handle, Some(on_message_trampoline), ptr as *mut c_void);
        }
        self.on_message_ud = Some(ptr);
        Ok(())
    }
}

impl Drop for Client {
    fn drop(&mut self) {
        unsafe { gomax_close(self.handle) };
        if let Some(ptr) = self.on_message_ud.take() {
            unsafe {
                drop(Box::from_raw(ptr));
            }
        }
    }
}

unsafe fn check_error(err: *mut c_char) -> Result<()> {
    if err.is_null() {
        return Ok(());
    }
    let msg = CStr::from_ptr(err).to_string_lossy().into_owned();
    gomax_free_string(err);
    Err(GomaxError(msg))
}

unsafe fn take_string(ptr: *mut c_char) -> Result<String> {
    if ptr.is_null() {
        return Err(GomaxError("unexpected NULL string from Go".into()));
    }
    let s = CStr::from_ptr(ptr).to_string_lossy().into_owned();
    gomax_free_string(ptr);
    Ok(s)
}
