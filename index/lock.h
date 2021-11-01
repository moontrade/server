#ifndef ART_LOCK_H
#define ART_LOCK_H

#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif


typedef struct moontrade_rwlock_t moontrade_rwlock_t;

uint64_t moontrade_rwlock_size();
void* moontrade_rwlock_new();
void moontrade_rwlock_lock(void* lock);
void moontrade_rwlock_unlock(void* lock);
void moontrade_rwlock_lock_shared(void* lock);
void moontrade_rwlock_unlock_shared(void* lock);
void moontrade_rwlock_destroy(void* lock);

#ifdef __cplusplus
}
#endif

#endif //ART_LOCK_H
