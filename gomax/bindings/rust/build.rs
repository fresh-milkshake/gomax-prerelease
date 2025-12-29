use std::env;
use std::fs;
use std::path::PathBuf;
use std::process::Command;

fn main() {
    println!("cargo:rerun-if-changed=ffi/bridge.go");

    let manifest_dir =
        PathBuf::from(env::var("CARGO_MANIFEST_DIR").expect("CARGO_MANIFEST_DIR not set"));
    let ffi_dir = manifest_dir.join("ffi");
    let out_dir = PathBuf::from(env::var("OUT_DIR").expect("OUT_DIR not set"));

    let target_os = env::var("CARGO_CFG_TARGET_OS").unwrap_or_else(|_| "linux".into());
    let target_arch = env::var("CARGO_CFG_TARGET_ARCH").unwrap_or_else(|_| "amd64".into());

    let go_os = match target_os.as_str() {
        "macos" => "darwin",
        other => other,
    };

    let go_arch = match target_arch.as_str() {
        "x86_64" => "amd64",
        "aarch64" => "arm64",
        "arm" => "arm",
        "x86" => "386",
        other => other,
    };

    let lib_name = "gomax";
    let ext = match target_os.as_str() {
        "windows" => "dll",
        "macos" => "dylib",
        _ => "so",
    };
    let lib_path = out_dir.join(format!("lib{lib_name}.{ext}"));

    let status = Command::new("go")
        .env("CGO_ENABLED", "1")
        .env("GOOS", go_os)
        .env("GOARCH", go_arch)
        .current_dir(&ffi_dir)
        .args([
            "build",
            "-buildmode=c-shared",
            "-o",
            lib_path.to_string_lossy().as_ref(),
            ".",
        ])
        .status()
        .expect("failed to start go build");

    if !status.success() {
        panic!("go build failed with status {status}");
    }

    if target_os == "windows" {
        // Создаём import library, чтобы MSVC линкер видел DLL без ручных шагов.
        let def_path = out_dir.join("libgomax.def");
        let def_body = r#"
LIBRARY libgomax.dll
EXPORTS
    gomax_new_client
    gomax_start
    gomax_close
    gomax_send_message
    gomax_get_profile
    gomax_get_chats
    gomax_set_on_message
    gomax_free_string
"#;
        fs::write(&def_path, def_body.trim_start().replace("\r\n", "\n")).expect("write def failed");

        let machine = match target_arch.as_str() {
            "x86_64" => "x64",
            "aarch64" => "ARM64",
            "x86" => "x86",
            other => {
                println!("cargo:warning=unsupported target_arch for import lib: {other}");
                "x64"
            }
        };

        let status = Command::new("llvm-lib")
            .current_dir(&out_dir)
            .args([
                format!("/def:{}", def_path.file_name().unwrap().to_string_lossy()),
                "/out:gomax.lib".into(),
                format!("/machine:{machine}"),
            ])
            .status()
            .expect("failed to run llvm-lib (need LLVM toolchain in PATH)");
        if !status.success() {
            panic!("llvm-lib failed with status {status}");
        }
    }

    // Header is emitted next to the library with the same base name.
    println!("cargo:rustc-link-search=native={}", out_dir.display());
    println!("cargo:rustc-link-lib=dylib={}", lib_name);
}
