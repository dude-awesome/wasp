package committee

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/sctransaction"
)

type StateTransactionMsg struct {
	*sctransaction.Transaction
	Confirmed bool
}

type BalancesMsg struct {
	Balances map[valuetransaction.ID][]*balance.Balance
}

type RequestMsg struct {
	*sctransaction.Transaction
	Index uint16
}

func (reqMsg *RequestMsg) RequestId() *sctransaction.RequestId {
	ret := sctransaction.NewRequestId(reqMsg.Transaction.ID(), reqMsg.Index)
	return &ret
}

func (reqMsg *RequestMsg) RequestBlock() *sctransaction.RequestBlock {
	return reqMsg.Requests()[reqMsg.Index]
}

func (reqMsg *RequestMsg) Timelock() uint32 {
	return reqMsg.RequestBlock().Timelock()
}
