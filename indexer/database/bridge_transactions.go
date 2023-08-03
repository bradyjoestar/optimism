package database

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

/**
 * Types
 */

type Transaction struct {
	FromAddress common.Address `gorm:"serializer:json"`
	ToAddress   common.Address `gorm:"serializer:json"`
	Amount      U256
	Data        hexutil.Bytes `gorm:"serializer:json"`
	Timestamp   uint64
}

type L1TransactionDeposit struct {
	SourceHash        common.Hash `gorm:"serializer:json;primaryKey"`
	L2TransactionHash common.Hash `gorm:"serializer:json"`

	InitiatedL1EventGUID uuid.UUID

	Version    U256
	OpaqueData hexutil.Bytes `gorm:"serializer:json"`

	Tx       Transaction `gorm:"embedded"`
	GasLimit U256
}

type L2TransactionWithdrawal struct {
	WithdrawalHash common.Hash `gorm:"serializer:json;primaryKey"`

	InitiatedL2EventGUID uuid.UUID

	ProvenL1EventGUID    *uuid.UUID
	FinalizedL1EventGUID *uuid.UUID

	Nonce U256

	Tx       Transaction `gorm:"embedded"`
	GasLimit U256
}

type BridgeTransactionsView interface {
	L1TransactionDeposit(common.Hash) (*L1TransactionDeposit, error)
	L2TransactionWithdrawal(common.Hash) (*L2TransactionWithdrawal, error)
}

type BridgeTransactionsDB interface {
	BridgeTransactionsView

	StoreL1TransactionDeposits([]*L1TransactionDeposit) error

	StoreL2TransactionWithdrawals([]*L2TransactionWithdrawal) error
	MarkL2TransactionWithdrawalProvenEvent(common.Hash, uuid.UUID) error
	MarkL2TransactionWithdrawalFinalizedEvent(common.Hash, uuid.UUID) error
}

/**
 * Implementation
 */

type bridgeTransactionsDB struct {
	gorm *gorm.DB
}

func newBridgeTransactionsDB(db *gorm.DB) BridgeTransactionsDB {
	return &bridgeTransactionsDB{gorm: db}
}

/**
 * Transactions deposited from L1
 */

func (db *bridgeTransactionsDB) StoreL1TransactionDeposits(deposits []*L1TransactionDeposit) error {
	result := db.gorm.Create(&deposits)
	return result.Error
}

func (db *bridgeTransactionsDB) L1TransactionDeposit(sourceHash common.Hash) (*L1TransactionDeposit, error) {
	var deposit L1TransactionDeposit
	result := db.gorm.Where(&L1TransactionDeposit{SourceHash: sourceHash}).Take(&deposit)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}

	return &deposit, nil
}

/**
 * Transactions withdrawn from L2
 */

func (db *bridgeTransactionsDB) StoreL2TransactionWithdrawals(withdrawals []*L2TransactionWithdrawal) error {
	result := db.gorm.Create(&withdrawals)
	return result.Error
}

func (db *bridgeTransactionsDB) L2TransactionWithdrawal(withdrawalHash common.Hash) (*L2TransactionWithdrawal, error) {
	var withdrawal L2TransactionWithdrawal
	result := db.gorm.Where(&L2TransactionWithdrawal{WithdrawalHash: withdrawalHash}).Take(&withdrawal)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}

	return &withdrawal, nil
}

// MarkL2TransactionWithdrawalProvenEvent links a withdrawn transaction with associated Prove action on L1.
func (db *bridgeTransactionsDB) MarkL2TransactionWithdrawalProvenEvent(withdrawalHash common.Hash, provenL1EventGuid uuid.UUID) error {
	withdrawal, err := db.L2TransactionWithdrawal(withdrawalHash)
	if err != nil {
		return err
	} else if withdrawal == nil {
		return fmt.Errorf("transaction withdrawal hash %s not found", withdrawalHash)
	}

	withdrawal.ProvenL1EventGUID = &provenL1EventGuid
	result := db.gorm.Save(&withdrawal)
	return result.Error
}

// MarkL2TransactionWithdrawalProvenEvent links a withdrawn transaction in its finalized state
func (db *bridgeTransactionsDB) MarkL2TransactionWithdrawalFinalizedEvent(withdrawalHash common.Hash, finalizedL1EventGuid uuid.UUID) error {
	withdrawal, err := db.L2TransactionWithdrawal(withdrawalHash)
	if err != nil {
		return err
	} else if withdrawal == nil {
		return fmt.Errorf("transaction withdrawal hash %s not found", withdrawalHash)
	} else if withdrawal.ProvenL1EventGUID == nil {
		return fmt.Errorf("cannot mark unproven withdrawal hash %s as finalized", withdrawal.WithdrawalHash)
	}

	withdrawal.FinalizedL1EventGUID = &finalizedL1EventGuid
	result := db.gorm.Save(&withdrawal)
	return result.Error
}