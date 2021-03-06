package micropay

import (
	"bytes"
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/util"
	"io"
)

type Payment struct {
	Ord            uint32
	Amount         int64
	SignatureShort []byte
}

type BatchPayment struct {
	Payer    address.Address
	Provider address.Address
	Payments []Payment
}

func NewPayment(ord uint32, amount int64, targetAddr address.Address, payerSigScheme signaturescheme.SignatureScheme) *Payment {
	data := paymentEssence(ord, amount, payerSigScheme.Address(), targetAddr)
	sig := payerSigScheme.Sign(data)
	shortSig := make([]byte, ed25519.SignatureSize)
	copy(shortSig, sig.Bytes()[1+ed25519.PublicKeySize:])
	return &Payment{
		Ord:            ord,
		Amount:         amount,
		SignatureShort: shortSig,
	}
}

func paymentEssence(ord uint32, amount int64, payerAddr, targetAddr address.Address) []byte {
	var buf bytes.Buffer
	buf.Write(util.Uint32To4Bytes(ord))
	buf.Write(util.Uint64To8Bytes(uint64(amount)))
	buf.Write(payerAddr[:])
	buf.Write(targetAddr[:])
	return buf.Bytes()
}

func NewPaymentFromBytes(data []byte) (*Payment, error) {
	ret := &Payment{}
	if err := ret.Read(bytes.NewReader(data)); err != nil {
		return nil, err
	}
	return ret, nil
}

func (p *Payment) Bytes() []byte {
	ret, err := util.Bytes(p)
	if err != nil {
		panic(err)
	}
	return ret
}

func (p *Payment) Write(w io.Writer) error {
	if err := util.WriteUint32(w, p.Ord); err != nil {
		return err
	}
	if err := util.WriteInt64(w, p.Amount); err != nil {
		return err
	}
	if err := util.WriteBytes16(w, p.SignatureShort); err != nil {
		return err
	}
	return nil
}

func (p *Payment) Read(r io.Reader) error {
	if err := util.ReadUint32(r, &p.Ord); err != nil {
		return err
	}
	if err := util.ReadInt64(r, &p.Amount); err != nil {
		return err
	}
	var err error
	if p.SignatureShort, err = util.ReadBytes16(r); err != nil {
		return err
	}
	if len(p.SignatureShort) != ed25519.SignatureSize {
		return fmt.Errorf("wrong public key bytes")
	}
	return nil
}
