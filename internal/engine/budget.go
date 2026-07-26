package engine

import (
	"fmt"
	"sync"
	"time"

	"github.com/sorafujitani/atlantis/internal/orchestration"
)

type BudgetSnapshot struct {
	Calls        int                 `json:"calls"`
	AdvisorCalls int                 `json:"advisor_calls"`
	Usage        orchestration.Usage `json:"usage"`
	Elapsed      time.Duration       `json:"elapsed"`
}

type Ledger struct {
	mu              sync.Mutex
	startedAt       time.Time
	maxCalls        int
	maxAdvisorCalls int
	calls           int
	advisorCalls    int
	usage           orchestration.Usage
}

func NewLedger(maxCalls, maxAdvisorCalls int) *Ledger {
	return &Ledger{startedAt: time.Now(), maxCalls: maxCalls, maxAdvisorCalls: maxAdvisorCalls}
}

func (l *Ledger) Reserve(advisor bool) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.calls >= l.maxCalls {
		return fmt.Errorf("call budget exhausted (%d)", l.maxCalls)
	}
	if advisor && l.advisorCalls >= l.maxAdvisorCalls {
		return fmt.Errorf("advisor call budget exhausted (%d)", l.maxAdvisorCalls)
	}
	l.calls++
	if advisor {
		l.advisorCalls++
	}
	return nil
}

func (l *Ledger) Settle(usage orchestration.Usage) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.usage.InputTokens += usage.InputTokens
	l.usage.OutputTokens += usage.OutputTokens
	l.usage.CostUSD += usage.CostUSD
}

func (l *Ledger) Snapshot() BudgetSnapshot {
	l.mu.Lock()
	defer l.mu.Unlock()
	return BudgetSnapshot{Calls: l.calls, AdvisorCalls: l.advisorCalls, Usage: l.usage, Elapsed: time.Since(l.startedAt)}
}
