#ifndef VO_LUV
#define VO_LUV

#include "uv.h"

typedef union uv_any_handle uv_any_handle;
typedef union uv_any_req uv_any_req;

struct HandleData {
    void *self;
    void *internal;
    void *user;
};

void _uv_timer_cb(uv_timer_t *timer){
    struct HandleData *data = (struct HandleData *)uv_handle_get_data((uv_handle_t *)timer);
    ((void (^)(void *))data->internal)(data->self);
};

void _uv_close_cb(uv_handle_t *handle) {
    struct HandleData *data = (struct HandleData *)uv_handle_get_data(handle);
    ((void (^)(void *))data->internal)(data->self);
};

void _uv_check_cb(uv_check_t *check) {
    struct HandleData *data = (struct HandleData *)uv_handle_get_data((uv_handle_t *)check);
    ((void (^)(void *))data->internal)(data->self);
};

#endif