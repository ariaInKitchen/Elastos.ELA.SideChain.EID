package auxpow

import (
	"errors"
	"io"

	"github.com/elastos/Elastos.ELA/auxpow"
	"github.com/elastos/Elastos.ELA/common"
	elatx "github.com/elastos/Elastos.ELA/core/transaction"
	elacommon "github.com/elastos/Elastos.ELA/core/types/common"
	ela "github.com/elastos/Elastos.ELA/core/types/interfaces"
	"github.com/elastos/Elastos.ELA/core/types/payload"
)

type SideAuxPow struct {
	SideAuxMerkleBranch []common.Uint256
	SideAuxMerkleIndex  int
	SideAuxBlockTx      ela.Transaction
	MainBlockHeader     elacommon.Header
}

func NewSideAuxPow(sideAuxMerkleBranch []common.Uint256,
	sideAuxMerkleIndex int,
	sideAuxBlockTx ela.Transaction,
	mainBlockHeader elacommon.Header) *SideAuxPow {

	return &SideAuxPow{
		SideAuxMerkleBranch: sideAuxMerkleBranch,
		SideAuxMerkleIndex:  sideAuxMerkleIndex,
		SideAuxBlockTx:      sideAuxBlockTx,
		MainBlockHeader:     mainBlockHeader,
	}
}

func (sap *SideAuxPow) Serialize(w io.Writer) error {
	err := sap.SideAuxBlockTx.Serialize(w)
	if err != nil {
		return err
	}

	err = common.WriteUint32(w, uint32(len(sap.SideAuxMerkleBranch)))
	if err != nil {
		return err
	}

	for _, branch := range sap.SideAuxMerkleBranch {
		err = branch.Serialize(w)
		if err != nil {
			return err
		}
	}

	err = common.WriteUint32(w, uint32(sap.SideAuxMerkleIndex))
	if err != nil {
		return err
	}

	return sap.MainBlockHeader.Serialize(w)
}

func (sap *SideAuxPow) Deserialize(r io.Reader) error {
	tx, err := elatx.GetTransactionByBytes(r)
	if err != nil {
		return err
	}

	err = tx.Deserialize(r)
	if err != nil {
		return err
	}

	sap.SideAuxBlockTx = tx

	count, err := common.ReadUint32(r)
	if err != nil {
		return err
	}

	sap.SideAuxMerkleBranch = make([]common.Uint256, 0)
	for i := uint32(0); i < count; i++ {
		var branch common.Uint256
		err = branch.Deserialize(r)
		if err != nil {
			return err
		}
		sap.SideAuxMerkleBranch = append(sap.SideAuxMerkleBranch, branch)
	}

	index, err := common.ReadUint32(r)
	if err != nil {
		return err
	}
	sap.SideAuxMerkleIndex = int(index)

	return sap.MainBlockHeader.Deserialize(r)
}

func (sap *SideAuxPow) SideAuxPowCheck(hashAuxBlock common.Uint256) error {
	mainBlockHeader := sap.MainBlockHeader
	mainBlockHeaderHash := mainBlockHeader.Hash()
	if !mainBlockHeader.AuxPow.Check(&mainBlockHeaderHash, auxpow.AuxPowChainID) {
		return errors.New("mainBlockHeader AuxPow check is failed")
	}

	sideAuxPowMerkleRoot := auxpow.GetMerkleRoot(sap.SideAuxBlockTx.Hash(), sap.SideAuxMerkleBranch, sap.SideAuxMerkleIndex)
	if sideAuxPowMerkleRoot != sap.MainBlockHeader.MerkleRoot {
		return errors.New("sideAuxPowMerkleRoot check is failed")
	}

	payloadData := sap.SideAuxBlockTx.Payload().Data(payload.SideChainPowVersion)
	payloadHashData := payloadData[0:32]
	payloadHash, err := common.Uint256FromBytes(payloadHashData)
	if err != nil {
		return errors.New("payloadHash to Uint256 is failed")
	}
	if *payloadHash != hashAuxBlock {
		return errors.New("payloadHash is not equal to hashAuxBlock")
	}

	return nil
}
