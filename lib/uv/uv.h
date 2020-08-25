void callback(uv_handle_t *handle) {
    
}

void uv_close_b(uv_handle_t *handle, void (^cb)(void *)) {
    uv_close(handle, &callback)
}