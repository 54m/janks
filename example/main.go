package main

import (
	"fmt"
	"github.com/54mch4n/janks"
)

func main() {
	if bank, err := janks.SearchBankByCode("0035", "001"); err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%+v\n", bank)
	}

	bank := janks.NewJapaneseBankSearch()
	if _, err := bank.SearchBankByName("ローソン銀行", "おすし支店"); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(bank.BankName)
		fmt.Println(bank.BranchCode)
		fmt.Println(bank.BranchName)
	}
}
