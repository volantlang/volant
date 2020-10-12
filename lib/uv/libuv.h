#ifndef VO_LUV
#define VO_LUV

#include "uv.h"

typedef struct sockaddr_in sockaddr_in;

typedef union uv_any_handle uv_any_handle;
typedef union uv_any_req uv_any_req;

struct Data {
    void *self;
    void *internal;
    void *user;
};

struct StreamData {
    void (^read_cb)(void *, ssize_t, void *);
    void (^write_cb)(void *, int);
    void (^connect_cb)(void *, int);
    void (^shutdown_cb)(void *, int);
    void (^connection_cb)(void *, int);
};

void vo_uv_alloc_cb(uv_handle_t *hanlde, size_t suggested_size, uv_buf_t *buf){
    buf->base = (char *)malloc(suggested_size);
    buf->len = suggested_size;
};

void vo_uv_close_cb(uv_handle_t *handle) {
    struct Data *data = (struct Data *)uv_handle_get_data(handle);
    ((void (^)(void *))data->internal)(data->self);
};

void vo_uv_timer_cb(uv_timer_t *timer){
    struct Data *data = (struct Data *)uv_handle_get_data((uv_handle_t *)timer);
    ((void (^)(void *))data->internal)(data->self);
};

void vo_uv_check_cb(uv_check_t *check) {
    struct Data *data = (struct Data *)uv_handle_get_data((uv_handle_t *)check);
    ((void (^)(void *))data->internal)(data->self);
};

void vo_uv_prepare_cb(uv_prepare_t *prepare) {
    struct Data *data = (struct Data *)uv_handle_get_data((uv_handle_t *)prepare);
    ((void (^)(void *))data->internal)(data->self);
};

void vo_uv_idle_cb(uv_idle_t *idle) {
    struct Data *data = (struct Data *)uv_handle_get_data((uv_handle_t *)idle);
    ((void (^)(void *))data->internal)(data->self);
};

void vo_uv_async_cb(uv_async_t *async) {
    struct Data *data = (struct Data *)uv_handle_get_data((uv_handle_t *)async);
    ((void (^)(void *))data->internal)(data->self);
};

void vo_uv_process_exit_cb(uv_process_t *process) {
    struct Data *data = (struct Data *)uv_handle_get_data((uv_handle_t *)process);
    ((void (^)(void *))data->internal)(data->self);
};

void vo_uv_signal_cb(uv_signal_t *handle, int signal) {
    struct Data *data = (struct Data *)uv_handle_get_data((uv_handle_t *)handle);
    ((void (^)(void *, int))data->internal)(data->self, signal);
};

void vo_uv_read_cb(uv_stream_t *stream, ssize_t nread, const uv_buf_t* buf){
    struct Data *data = (struct Data *)uv_handle_get_data((uv_handle_t *)stream);
    ((void (^)(void *, ssize_t, const uv_buf_t *))data->internal)(data->self, nread, buf);
};

void vo_uv_write_cb(uv_write_t *req, int status){
    struct Data *data = (struct Data *)uv_req_get_data((uv_req_t *)req);
    ((void (^)(void *, int))data->internal)(data->self, status);
};

void vo_uv_shutdown_cb(uv_shutdown_t *req, int status){
    struct Data *data = (struct Data *)uv_req_get_data((uv_req_t *)req);
    ((void (^)(void *, int))data->internal)(data->self, status);
};

void vo_uv_connect_cb(uv_connect_t *req, int status){
    struct Data *data = (struct Data *)uv_req_get_data((uv_req_t *)req);
    ((void (^)(void *, int))data->internal)(data->self, status);
};

void vo_uv_connection_cb(uv_stream_t *stream, int status){
    struct Data *data = (struct Data *)uv_handle_get_data((uv_handle_t *)stream);
    ((void (^)(void *, int))data->internal)(data->self, status);
};

#endif