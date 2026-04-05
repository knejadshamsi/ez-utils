mod workspace;

use workspace::{
    close_workspace, create_workspace, get_workspace_state, import_source, open_workspace,
    remove_source, rename_source, save_ui_state, save_workspace, WorkspaceManager,
};

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .manage(WorkspaceManager::default())
        .plugin(tauri_plugin_dialog::init())
        .plugin(tauri_plugin_shell::init())
        .invoke_handler(tauri::generate_handler![
            create_workspace,
            open_workspace,
            save_workspace,
            close_workspace,
            get_workspace_state,
            import_source,
            rename_source,
            remove_source,
            save_ui_state
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
