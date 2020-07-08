package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/shopspring/decimal"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

var (
	descrCRDRe  = regexp.MustCompile(` [^ ]+?\\.+?\\.+?\\.+?\\(.+?)  `)
	descrHOLDRe = regexp.MustCompile(`^(?:[0-9]+ )?[^ ]+ +?([^>]+)`)
)

type Date struct {
	time.Time
}

func (date *Date) MarshalCSV() (string, error) {
	return date.Time.Format("01/02/2006"), nil
}

func (date *Date) UnmarshalCSV(csv string) (err error) {
	date.Time, err = time.Parse("02.01.06", csv)
	return err
}

type Amount struct {
	d decimal.Decimal
}

func (a Amount) String() string {
	return a.d.StringFixed(2)
}

func (a *Amount) UnmarshalCSV(s string) (err error) {
	a.d, err = decimal.NewFromString(strings.ReplaceAll(s, ",", "."))
	return err
}

type AlfaOperation struct {
	AccType     string `csv:"Тип счёта"`
	AccNumber   string `csv:"Номер счета"`
	Currency    string `csv:"Валюта"`
	Date        Date   `csv:"Дата операции"`
	Reference   string `csv:"Референс проводки"`
	Description string `csv:"Описание операции"`
	Inflow      Amount `csv:"Приход"`
	Outflow     Amount `csv:"Расход"`
}

type YNABOperation struct {
	Date    Date
	Payee   string
	Memo    string
	Outflow Amount
	Inflow  Amount
}

func main() {
	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
		r := csv.NewReader(
			transform.NewReader(in, charmap.Windows1251.NewDecoder()),
		)
		r.Comma = ';'
		r.LazyQuotes = true
		return r
	})

	inOperations := []*AlfaOperation{}

	if err := gocsv.UnmarshalFile(os.Stdin, &inOperations); err != nil {
		panic(err)
	}

	outOperations := []*YNABOperation{}
	for _, inOp := range inOperations {
		outOp := &YNABOperation{
			Date:    inOp.Date,
			Inflow:  inOp.Inflow,
			Outflow: inOp.Outflow,
		}

		if inOp.Reference == "HOLD" {
			outOp.Payee = getFirstMatch(inOp.Description, descrHOLDRe)
		} else if strings.HasPrefix(inOp.Reference, "CRD_") {
			outOp.Payee = getFirstMatch(inOp.Description, descrCRDRe)
		} else if strings.HasPrefix(inOp.Reference, "MOPJ") {
			outOp.Payee = "Плата за оповещения об операциях"
			outOp.Memo = inOp.Description
		} else if strings.HasPrefix(inOp.Reference, "C") {
			outOp.Payee = "Перевод"
			outOp.Memo = inOp.Description
		} else if strings.HasPrefix(inOp.Reference, "B") {
			outOp.Payee = "Перевод"
			outOp.Memo = inOp.Description
		}

		if len(outOp.Payee) == 0 {
			outOp.Payee = inOp.Description
		}

		outOperations = append(outOperations, outOp)
	}

	csvContent, err := gocsv.MarshalString(&outOperations)
	if err != nil {
		panic(err)
	}
	fmt.Print(csvContent)
}

func getFirstMatch(in string, re *regexp.Regexp) string {
	if sm := re.FindStringSubmatch(in); len(sm) > 1 {
		return sm[1]
	}
	return ""
}
