package fails

import (
	"log"
	"sync"

	"github.com/morikuni/failure"
)

const (
	// ErrCritical はクリティカルなエラー。少しでも大幅減点・失格になるエラー
	ErrCritical failure.StringCode = "error critical"
	// ErrApplication はアプリケーションの挙動でおかしいエラー。Verify時は1つでも失格。Validation時は一定数以上で失格
	ErrApplication failure.StringCode = "error application"
	// ErrTimeout はタイムアウトエラー。基本は大目に見る。
	ErrTimeout failure.StringCode = "error timeout"
	// ErrTemporary は一時的なエラー。基本は大目に見る。
	ErrTemporary failure.StringCode = "error temporary"
)

var (
	ErrorsForCheck *Errors
	ErrorsForFinal *Errors
)

func init() {
	ErrorsForCheck = NewErrors()
	ErrorsForFinal = NewErrors()
}

type Errors struct {
	Msgs []string

	critical    int
	application int
	trivial     int

	mu sync.Mutex
}

func NewErrors() *Errors {
	msgs := make([]string, 0, 100)
	return &Errors{
		Msgs: msgs,
	}
}

func (e *Errors) GetMsgs() (msgs []string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	return e.Msgs[:]
}

func (e *Errors) Get() (msgs []string, critical, application, trivial int) {
	e.mu.Lock()
	defer e.mu.Unlock()

	return e.Msgs[:], e.critical, e.application, e.trivial
}

func (e *Errors) Add(err error) {
	if err == nil {
		return
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	log.Printf("%+v", err)

	if msg, ok := failure.MessageOf(err); ok {
		switch code, _ := failure.CodeOf(err); code {
		case ErrCritical:
			msg += " (critical error)"
			e.critical++
		case ErrApplication:
			e.application++
		case ErrTimeout:
			msg += "（タイムアウトしました）"
			e.trivial++
		case ErrTemporary:
			msg += "（一時的なエラー）"
			e.trivial++
		}

		e.Msgs = append(e.Msgs, msg)
	} else {
		// 想定外のエラーなのでcritical扱いにしておく
		e.critical++
		e.Msgs = append(e.Msgs, "運営に連絡してください")
	}
}
