package chain

import (
	"errors"
	"fmt"
	"github.com/SongOf/edge-storage-core/core"
	"testing"
)

func TestFunctionChain(t *testing.T) {
	chain := FunctionChain{}

	err := chain.AddUnitList(
		Unit{
			ForwardFunc: func(*core.Context) error {
				fmt.Println("unit1 forward")
				return nil
			},
			RollbackFunc: func(*core.Context) error {
				fmt.Println("unit1 rollback")
				return nil
			},
		},
		Unit{
			ForwardFunc: func(*core.Context) error {
				fmt.Println("unit2 forward")
				return nil
			},
			RollbackFunc: func(*core.Context) error {
				fmt.Println("unit2 rollback")
				return nil
			},
		},
		Unit{
			ForwardFunc: func(*core.Context) error {
				fmt.Println("unit3 forward")
				return errors.New("unit3 return error")
			},
			RollbackFunc: func(*core.Context) error {
				fmt.Println("unit3 rollback")
				return nil
			},
		},
	).Run(&core.Context{})

	if err != nil {
		fmt.Println(err)
	}
}

func TestNewChain(t *testing.T) {
	chain := NewChain(
		NewUnit(func(*core.Context) error {
			fmt.Println("unit1 forward")
			return nil
		}, func(*core.Context) error {
			fmt.Println("unit1 rollback")
			return nil
		}),
		NewUnit(func(*core.Context) error {
			fmt.Println("unit2 forward")
			return nil
		}, func(*core.Context) error {
			fmt.Println("unit2 rollback")
			return nil
		}),
		NewUnit(func(*core.Context) error {
			fmt.Println("unit3 forward")
			return nil
		}, nil),
	)

	err := chain.Run(&core.Context{})
	if err != nil {
		fmt.Println(err)
		t.Error(err)
	}
}
