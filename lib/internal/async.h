#ifndef VO_INTERNAL_ASYNC
#define VO_INTERNAL_ASYNC

#define VO_ASYNC_START_CONTEXT()   a_ctx_t *a_ctx = (a_ctx_t*)(check->data); switch(a_ctx->state){ case 0:;
#define VO_AWAIT(promise, state1)  ({ ctx->prom[state1-1] = promise; a_ctx->state++; case state1: if(a_ctx->prom[state1-1]->state == PENDING) return; a_ctx->prom[state1-1]->val; })
#define VO_RETURN_ASYNC(value)     RESOLVE_PROM(ctx->return_promise, value); uv_idle_stop(check); free(check); free(a_ctx); return;
#define VO_ASYNC_VAR(var)          a_ctx->v_##var
#define VO_ASYNC_END_CONTEXT()     a_ctx->return_promise->state = COMPLETED; uv_idle_stop(check); free(check); free(a_ctx); return; }

#endif