package janks

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/xerrors"
)

var defaultJanks *Bank

// Bank bank information
type Bank struct {
	BankName       string
	BankNameHalf   string
	BankCode       string
	BranchName     string
	BranchNameHalf string
	BranchCode     string
}

func init() {
	defaultJanks = NewJapaneseBankSearch()
}

// NewJapaneseBankSearch constructor
func NewJapaneseBankSearch() *Bank {
	return new(Bank)
}

// SearchBankByCode search for a bank by bank code
func SearchBankByCode(bc, sc string) (*Bank, error) {
	return defaultJanks.SearchBankByCode(bc, sc)
}

// SearchBankByName search for a bank by bank name and branch name
func SearchBankByName(bc, sc string) (*Bank, error) {
	return defaultJanks.SearchBankByName(bc, sc)
}

// SearchBankByCode search for a bank by bank code
func (src *Bank) SearchBankByCode(bc, sc string) (*Bank, error) {
	form := url.Values{}
	form.Add("inbc", bc)
	form.Add("insc", sc)

	body := strings.NewReader(form.Encode())

	req, err := http.NewRequest("POST", "https://zengin.ajtw.net/search.php", body)
	if err != nil {
		return nil, setError(err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, setError(err)
	}
	defer resp.Body.Close()

	byteArray, _ := ioutil.ReadAll(resp.Body)
	stringReader := strings.NewReader(string(byteArray))
	doc, err := goquery.NewDocumentFromReader(stringReader)
	if err != nil {
		return nil, setError(err)
	}

	doc.Find("table > tbody > tr").Each(func(n int, s *goquery.Selection) {
		txt := strings.Split(s.Text(), "\n")
		switch n {
		case 1:
			src.BankName = txt[1]
			src.BankNameHalf = txt[2]
			src.BankCode = txt[3]
		case 4:
			src.BranchName = txt[1]
			src.BranchNameHalf = txt[2]
			src.BranchCode = txt[3]
		}
	})

	return src.responseCheck()
}

// SearchBankByName search for a bank by bank name and branch name
func (src *Bank) SearchBankByName(bc, sc string) (_ *Bank, err error) {
	form := url.Values{}
	form.Add("wd", bc)

	body := strings.NewReader(form.Encode())

	req, err := http.NewRequest("POST", "https://zengin.ajtw.net/ginkoukw.php", body)
	if err != nil {
		return nil, setError(err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, setError(err)
	}
	defer resp.Body.Close()

	byteArray, _ := ioutil.ReadAll(resp.Body)
	stringReader := strings.NewReader(string(byteArray))
	doc, err := goquery.NewDocumentFromReader(stringReader)
	if err != nil {
		return nil, setError(err)
	}

	doc.Find("table > tbody > tr > td.g1").Each(func(n int, s *goquery.Selection) {
		text := s.Text()
		if text == " " {
			err = setError(xerrors.New("SearchBankByNameError: Not found. [bank]"))
			return
		} else {
			switch n {
			case 0:
				src.BankName = text
			case 1:
				src.BankNameHalf = text
			}
		}
	})
	if err != nil {
		return nil, setError(err)
	}
	src.BankCode = doc.Find("table > tbody > tr > td.g2").Text()

	form = url.Values{}
	form.Add("wd", sc)
	if query, ok := doc.Find("table > tbody > tr > td.g3 > form > input[type=hidden]").Attr("value"); !ok {
		return nil, setError(xerrors.New("Not found."))
	} else {
		form.Add("pz", query)
	}

	body = strings.NewReader(form.Encode())
	req, err = http.NewRequest("POST", "https://zengin.ajtw.net/shitenmeisaikw.php", body)
	if err != nil {
		return nil, setError(err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client = &http.Client{Timeout: 10 * time.Second}
	resp, err = client.Do(req)
	if err != nil {
		return nil, setError(err)
	}
	defer resp.Body.Close()

	byteArray, _ = ioutil.ReadAll(resp.Body)
	stringReader = strings.NewReader(string(byteArray))
	doc, err = goquery.NewDocumentFromReader(stringReader)
	if err != nil {
		return nil, setError(err)
	}

	doc.Find("table > tbody > tr > td.g1").Each(func(n int, s *goquery.Selection) {
		text := s.Text()
		if text == " " {
			err = setError(xerrors.New("Not found. [branch]"))
		} else {
			switch n {
			case 0:
				src.BranchName = text
			case 1:
				src.BranchNameHalf = text
			}
		}
	})
	if err != nil {
		return nil, setError(err)
	}

	src.BranchCode = doc.Find("table > tbody > tr > td.g2").Text()

	return src.responseCheck()
}

func (src *Bank) responseCheck() (*Bank, error) {
	ind := reflect.Indirect(reflect.ValueOf(src))
	for i := 0; i < ind.NumField(); i++ {
		switch ind.Field(i).String() {
		case "":
			return nil, setError(xerrors.New("Empty character."), 2)
		case "該当するデータはありません":
			return nil, setError(xerrors.New("No found data."), 2)
		}
	}
	return src, nil
}

func setError(err error, nest ...interface{}) error {
	skip := 1
	if len(nest) != 0 {
		skip = nest[0].(int)
	}
	pc, _, _, _ := runtime.Caller(skip)
	sp := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	fn := sp[len(sp)-1]
	return xerrors.Errorf("%sError: %s", fn, err.Error())
}
