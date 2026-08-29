#[tauri::command]
fn backend_version() -> &'static str {
    env!("CARGO_PKG_VERSION")
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .invoke_handler(tauri::generate_handler![backend_version])
        .run(tauri::generate_context!())
        .expect("running the sddviz application");
}
