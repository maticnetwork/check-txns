package main

import (
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

func main() {
	input, err := os.Open("/home/ferran/Downloads/transactions.rlp")
	if err != nil {
		panic(err)
	}
	defer input.Close()

	// Inject all transactions from the journal into the pool
	stream := rlp.NewStream(input, 0)

	var handler Handler
	handler = &aggregateSender{
		sender: map[common.Address]uint64{},
	}

	var (
		failure error
	)
	for {
		// Parse the next transaction and terminate on error
		tx := new(types.Transaction)
		if err = stream.Decode(tx); err != nil {
			if err != io.EOF {
				failure = err
			}
			break
		}

		handler.Handle(tx)
	}

	handler.Finish()
	fmt.Println(failure)
}

type Handler interface {
	Handle(tx *types.Transaction)
	Finish()
}

type aggregateSender struct {
	sender           map[common.Address]uint64
	contractCreation uint64
}

func (a *aggregateSender) Handle(tx *types.Transaction) {
	if tx.To() == nil {
		a.contractCreation++
	} else {
		to := *tx.To()
		_, ok := a.sender[to]
		if !ok {
			a.sender[to] = 1
		} else {
			a.sender[to]++
		}
	}
}

func (a *aggregateSender) Finish() {
	bySenderMap := sortMapSender(a.sender)

	fmt.Printf("Contract creation: %d\n", a.contractCreation)
	fmt.Printf("Unique accounts: %d\n", len(a.sender))

	fmt.Printf("Sort by target contract")
	for i := 0; i < 20; i++ {
		fmt.Printf("Target %s: %d\n", bySenderMap[i].sender, bySenderMap[i].count)
	}

	total := 0
	for _, i := range bySenderMap {
		total += int(i.count)
	}
	fmt.Println(total)
}

type totalCount struct {
	total uint64
}

func (t *totalCount) Handle(tx *types.Transaction) {
	t.total++
}

func (t *totalCount) Finish() {
	fmt.Printf("Total: %d\n", t.total)
}

type item struct {
	sender common.Address
	count  uint64
}

func sortMapSender(m map[common.Address]uint64) items {
	items := items{}
	for k, v := range m {
		items = append(items, item{
			sender: k,
			count:  v,
		})
	}
	sort.Sort(items)
	return items
}

type items []item

func (a items) Len() int           { return len(a) }
func (a items) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a items) Less(i, j int) bool { return a[i].count > a[j].count }
