package executor

import (
	"github.com/zarbchain/zarb-go/errors"
	"github.com/zarbchain/zarb-go/sandbox"
	"github.com/zarbchain/zarb-go/tx"
	"github.com/zarbchain/zarb-go/tx/payload"
)

type BondExecutor struct {
	fee    int64
	strict bool
}

func NewBondExecutor(strict bool) *BondExecutor {
	return &BondExecutor{strict: strict}
}

func (e *BondExecutor) Execute(trx *tx.Tx, sb sandbox.Sandbox) error {
	pld := trx.Payload().(*payload.BondPayload)

	bonderAcc := sb.Account(pld.Bonder)
	if bonderAcc == nil {
		return errors.Errorf(errors.ErrInvalidTx, "unable to retrieve bonder account")
	}
	val := sb.Validator(pld.Validator.Address())
	if val == nil {
		val = sb.MakeNewValidator(pld.Validator)
	}
	if e.strict && sb.IsInCommittee(pld.Validator.Address()) {
		return errors.Errorf(errors.ErrInvalidTx, "validator is in committee right now")
	}
	if bonderAcc.Sequence()+1 != trx.Sequence() {
		return errors.Errorf(errors.ErrInvalidTx, "invalid sequence. Expected: %v, got: %v", bonderAcc.Sequence()+1, trx.Sequence())
	}
	if bonderAcc.Balance() < pld.Stake+trx.Fee() {
		return errors.Errorf(errors.ErrInvalidTx, "insufficient balance")
	}
	bonderAcc.IncSequence()
	bonderAcc.SubtractFromBalance(pld.Stake + trx.Fee())
	val.AddToStake(pld.Stake)

	sb.UpdateAccount(bonderAcc)
	sb.UpdateValidator(val)

	e.fee = trx.Fee()

	return nil
}

func (e *BondExecutor) Fee() int64 {
	return e.fee
}
