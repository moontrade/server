#include "lock.h"
#include "RWSpinLock.h"

#include <shared_mutex>

uint64_t moontrade_rwlock_size() {
    return (uint64_t)sizeof(RWTicketSpinLock64);
}
void* moontrade_rwlock_new() {
    return new RWTicketSpinLock64();
}
void moontrade_rwlock_lock(void* lock) {
    ((RWTicketSpinLock64*)(lock))->lock();
}
void moontrade_rwlock_unlock(void* lock) {
    ((RWTicketSpinLock64*)(lock))->unlock();
}
void moontrade_rwlock_lock_shared(void* lock) {
    ((RWTicketSpinLock64*)(lock))->lock_shared();
//    ((RWTicketSpinLock64*)(lock))->lock();
}
void moontrade_rwlock_unlock_shared(void* lock) {
    ((RWTicketSpinLock64*)(lock))->unlock_shared();
//    ((RWTicketSpinLock64*)(lock))->unlock();
}
void moontrade_rwlock_destroy(void* lock) {
    delete (RWTicketSpinLock64*)(lock);
}
