package data

import (
	"context"
	"errors"
	"sync"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Wallet struct {
	UUID    uuid.UUID `db:"uuid" json:"uuid"`
	Balance float64   `db:"balance" json:"balance"`
	Version int       `db:"version" json:"version"`
}

type WalletModel struct {
	DB *pgxpool.Pool
}

func (m WalletModel) Create() (*Wallet, error) {

	query := `
	INSERT INTO wallets DEFAULT VALUES
	RETURNING uuid,balance,version`

	var wallet Wallet

	err := m.DB.QueryRow(context.Background(), query).Scan(&wallet.UUID, &wallet.Balance, &wallet.Version)

	return &wallet, err
}

func (m WalletModel) Get(id uuid.UUID) (*Wallet, error) {

	query := `
	SELECT uuid, balance, version
	FROM wallets
	WHERE uuid = $1`

	var wallet Wallet

	err := m.DB.QueryRow(context.Background(), query, id).Scan(
		&wallet.UUID,
		&wallet.Balance,
		&wallet.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &wallet, nil
}

func (m WalletModel) Update(wallet *Wallet) error {
	query := `
	UPDATE wallets
	SET balance = $1, version = version + 1
	WHERE uuid = $2 AND version = $3
	RETURNING version`

	args := []any{
		wallet.Balance,
		wallet.UUID,
		wallet.Version,
	}
	_, err := m.DB.Exec(context.Background(), query, args...)

	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (m WalletModel) Delete(id uuid.UUID) error {
	query := `
	DELETE FROM wallets
	WHERE uuid = $1`

	tag, err := m.DB.Exec(context.Background(), query, id)

	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrEditConflict
	}
	return nil
}

type MockWalletModel struct {
	mu      sync.RWMutex
	wallets map[uuid.UUID]*Wallet
}

func (m *MockWalletModel) Create() (*Wallet, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	wallet := Wallet{
		UUID:    uuid.New(),
		Balance: 0,
		Version: 1,
	}
	m.wallets[wallet.UUID] = &wallet

	return &wallet, nil
}

func (m *MockWalletModel) Get(id uuid.UUID) (*Wallet, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	wallet, exists := m.wallets[id]
	if !exists {
		return nil, ErrRecordNotFound
	}
	return wallet, nil
}

func (m *MockWalletModel) Update(wallet *Wallet) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.wallets[wallet.UUID]; !exists {
		return ErrRecordNotFound
	}
	wallet.Version++
	m.wallets[wallet.UUID] = wallet
	return nil
}

func (m *MockWalletModel) Delete(id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.wallets[id]; !exists {
		return ErrRecordNotFound
	}
	delete(m.wallets, id)
	return nil
}
