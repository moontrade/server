package model

import "sync"

var (
	poolEOS = sync.Pool{New: func() interface{} {
		return &EOS{}
	}}
	poolEOB = sync.Pool{New: func() interface{} {
		return &EOB{}
	}}
	poolSavepoint = sync.Pool{New: func() interface{} {
		return &Savepoint{}
	}}
	poolStarting = sync.Pool{New: func() interface{} {
		return &Starting{}
	}}
	poolProgress = sync.Pool{New: func() interface{} {
		return &Progress{}
	}}
	poolStarted = sync.Pool{New: func() interface{} {
		return &Started{}
	}}
	poolStopped = sync.Pool{New: func() interface{} {
		return &Stopped{}
	}}
)

func GetEOS() *EOS {
	return poolEOS.Get().(*EOS)
}
func PutEOS(msg *EOS) {
	poolEOS.Put(msg)
}

func GetEOB() *EOB {
	return poolEOB.Get().(*EOB)
}
func PutEOB(msg *EOB) {
	poolEOB.Put(msg)
}

func GetSavepoint() *Savepoint {
	return poolSavepoint.Get().(*Savepoint)
}
func PutSavepoint(msg *Savepoint) {
	poolSavepoint.Put(msg)
}

func GetStarting() *Starting {
	return poolStarting.Get().(*Starting)
}
func PutStarting(msg *Starting) {
	poolStarting.Put(msg)
}

func GetProgress() *Progress {
	return poolProgress.Get().(*Progress)
}
func PutProgress(msg *Progress) {
	poolProgress.Put(msg)
}

func GetStarted() *Started {
	return poolStarted.Get().(*Started)
}
func PutStarted(msg *Started) {
	poolStarted.Put(msg)
}

func GetStopped() *Stopped {
	return poolStopped.Get().(*Stopped)
}
func PutStopped(msg *Stopped) {
	poolStopped.Put(msg)
}
