package data

import (
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type Models struct {
	Wallets interface {
		Create() (*Wallet, error)
		Get(id uuid.UUID) (*Wallet, error)
		Update(wallet *Wallet) error
		Delete(id uuid.UUID) error
	}
}

func NewModels(db *pgxpool.Pool) Models {
	return Models{
		Wallets: WalletModel{DB: db},
	}
}

func NewMockModels() Models {
	return Models{
		Wallets: &MockWalletModel{
			wallets: make(map[uuid.UUID]*Wallet),
		},
	}
}
