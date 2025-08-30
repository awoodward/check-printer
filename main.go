package main

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	_ "github.com/mattn/go-sqlite3"
	"github.com/phpdave11/gofpdf"
	"github.com/rivo/tview"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type Account struct {
	ID            int
	AccountName   string
	NameOnAccount string
	Address1      string
	Address2      string
	AccountNumber int
	RTN           int
	LastCheckNum  int
	BankName      string
}

type CheckLog struct {
	ID          int
	AccountID   int
	CheckNumber int
	Payee       string
	Amount      float64
	Memo        string
	Date        time.Time
	Filename    string
}

type App struct {
	db   *sql.DB
	app  *tview.Application
	main *tview.Flex
}

func initDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./checks.db")
	if err != nil {
		return nil, err
	}

	// Create accounts table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS accounts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			account_name TEXT NOT NULL,
			name_on_account TEXT NOT NULL,
			address1 TEXT NOT NULL,
			address2 TEXT NOT NULL,
			account_number INTEGER NOT NULL,
			rtn INTEGER NOT NULL,
			last_check_num INTEGER NOT NULL DEFAULT 1000,
			bank_name TEXT NOT NULL
		)
	`)
	if err != nil {
		return nil, err
	}

	// Create check_logs table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS check_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			account_id INTEGER NOT NULL,
			check_number INTEGER NOT NULL,
			payee TEXT NOT NULL,
			amount REAL NOT NULL,
			memo TEXT,
			date DATETIME NOT NULL,
			filename TEXT NOT NULL,
			FOREIGN KEY (account_id) REFERENCES accounts (id)
		)
	`)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func pow(i int, p int) int {
	return int(math.Pow(1000, float64(p)))
}

func spell(n int) string {
	to19 := []string{"One", "Two", "Three", "Four", "Five", "Six", "Seven", "Eight", "Nine", "Ten", "Eleven", "Twelve",
		"Thirteen", "Fourteen", "Fifteen", "Sixteen", "Seventeen", "Eighteen", "Nineteen"}

	tens := []string{"Twenty", "Thirty", "Forty", "Fifty", "Sixty", "Seventy", "Eighty", "Ninety"}
	if n == 0 {
		return ""
	}
	if n < 20 {
		return to19[n-1]
	}
	if n < 100 {
		return tens[n/10-2] + " " + spell(n%10)
	}
	if n < 1000 {
		return to19[n/100-1] + " Hundred " + spell(n%100)
	}

	for idx, w := range []string{"Thousand", "Million", "Billion"} {
		p := idx + 1
		if n < pow(1000, (p+1)) {
			return spell(n/pow(1000, p)) + " " + w + " " + spell(n%pow(1000, p))
		}
	}

	return "error"
}

func writeCheck(account Account, checkNum int, payee, memo string, amount float64, filename string) error {
	pdf := gofpdf.New("P", "mm", "Letter", ".")
	pdf.AddPage()

	// Account name and bank name
	pdf.SetFont("Arial", "B", 12)
	pdf.MoveTo(30, 12)
	pdf.Cell(0, 0, account.NameOnAccount)
	pdf.MoveTo(110, 12)
	pdf.Cell(0, 0, account.BankName)
	pdf.SetFont("Arial", "", 10)
	pdf.MoveTo(30, 16)
	pdf.Cell(0, 0, account.Address1)
	pdf.MoveTo(30, 20)
	pdf.Cell(0, 0, account.Address2)

	// Check number
	pdf.SetFont("Courier", "B", 12)
	chkstr := fmt.Sprintf("%v", checkNum)
	strW := pdf.GetStringWidth(chkstr)
	pdf.MoveTo(203-strW, 12)
	pdf.Cell(0, 0, chkstr)

	// Check date
	pdf.SetFont("Arial", "", 12)
	chkstr = time.Now().Format("January 2, 2006")
	strW = pdf.GetStringWidth(chkstr)
	pdf.MoveTo(203-strW, 21)
	pdf.Cell(0, 0, chkstr)

	// Payee
	pdf.SetFont("Arial", "", 9)
	pdf.MoveTo(10, 30)
	pdf.Cell(0, 0, "PAY TO THE ORDER OF")

	pdf.SetFont("Arial", "", 13)
	pdf.MoveTo(15, 37)
	pdf.Cell(0, 0, payee)

	// memo
	if len(memo) > 0 {
		pdf.SetFont("Arial", "", 11)
		pdf.MoveTo(15, 70)
		pdf.Cell(0, 0, memo)
	}

	// Check amount in words
	pdf.SetFont("Arial", "", 13)
	pdf.MoveTo(30, 52)
	chkstr = fmt.Sprintf("%v and %v/100", spell(int(amount)), math.Round((amount-float64(int(amount)))*100))
	pdf.Cell(0, 0, chkstr)

	pdf.SetFont("Arial", "", 9)
	chkstr = "CHECK AMOUNT"
	strW = pdf.GetStringWidth(chkstr)
	pdf.MoveTo(203-strW, 30)
	pdf.Cell(0, 0, chkstr)

	// Check amount in numbers
	pdf.AddUTF8Font("security", "", "security.ttf")
	pdf.SetFont("security", "", 20)
	p := message.NewPrinter(language.English)
	chkstr = p.Sprintf("$%.2f", amount)
	strW = pdf.GetStringWidth(chkstr)
	pdf.MoveTo(203-strW, 37)
	pdf.Cell(0, 0, chkstr)

	// Signature line
	pdf.SetFont("Arial", "", 6)
	pdf.MoveTo(155, 57)
	pdf.Cell(0, 0, "VOID AFTER 180 DAYS")

	pdf.Line(155, 74, 203, 74)
	pdf.MoveTo(160, 76)
	pdf.Cell(0, 0, "AUTHORIZED SIGNATURE")

	// Print MICR
	pdf.AddUTF8Font("micr", "", "micr.ttf")
	pdf.SetFont("micr", "", 10)
	chkstr = fmt.Sprintf("o%vo", checkNum)
	strW = pdf.GetStringWidth(chkstr)
	pdf.MoveTo(69-strW, 83)
	pdf.Cell(0, 0, chkstr)

	chkstr = fmt.Sprintf("t%09dt", account.RTN)
	pdf.MoveTo(72, 83)
	pdf.Cell(0, 0, chkstr)

	chkstr = fmt.Sprintf("%vo", account.AccountNumber)
	strW = pdf.GetStringWidth(chkstr)
	pdf.MoveTo(150-strW, 83)
	pdf.Cell(0, 0, chkstr)

	return pdf.OutputFileAndClose(filename)
}

func (a *App) getAccounts() ([]Account, error) {
	rows, err := a.db.Query("SELECT id, account_name, name_on_account, address1, address2, account_number, rtn, last_check_num, bank_name FROM accounts")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []Account
	for rows.Next() {
		var acc Account
		err := rows.Scan(&acc.ID, &acc.AccountName, &acc.NameOnAccount, &acc.Address1, &acc.Address2, &acc.AccountNumber, &acc.RTN, &acc.LastCheckNum, &acc.BankName)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, acc)
	}
	return accounts, nil
}

func (a *App) saveAccount(acc Account) error {
	_, err := a.db.Exec(`
		INSERT INTO accounts (account_name, name_on_account, address1, address2, account_number, rtn, last_check_num, bank_name)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, acc.AccountName, acc.NameOnAccount, acc.Address1, acc.Address2, acc.AccountNumber, acc.RTN, acc.LastCheckNum, acc.BankName)
	return err
}

func (a *App) updateLastCheckNum(accountID, checkNum int) error {
	_, err := a.db.Exec("UPDATE accounts SET last_check_num = ? WHERE id = ?", checkNum, accountID)
	return err
}

func (a *App) logCheck(accountID, checkNum int, payee, memo string, amount float64, filename string) error {
	_, err := a.db.Exec(`
		INSERT INTO check_logs (account_id, check_number, payee, amount, memo, date, filename)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, accountID, checkNum, payee, amount, memo, time.Now(), filename)
	return err
}

func (a *App) showAddAccountForm() {
	form := tview.NewForm()
	form.SetBorder(true).SetTitle("Add New Account")

	var account Account

	form.AddInputField("Account Name", "", 30, nil, func(text string) {
		account.AccountName = text
	})
	form.AddInputField("Name on Account", "", 30, nil, func(text string) {
		account.NameOnAccount = text
	})
	form.AddInputField("Bank Name", "", 30, nil, func(text string) {
		account.BankName = text
	})
	form.AddInputField("Address 1", "", 30, nil, func(text string) {
		account.Address1 = text
	})
	form.AddInputField("Address 2", "", 30, nil, func(text string) {
		account.Address2 = text
	})
	form.AddInputField("Account Number", "", 20, func(textToCheck string, lastChar rune) bool {
		return lastChar >= '0' && lastChar <= '9'
	}, func(text string) {
		if num, err := strconv.Atoi(text); err == nil {
			account.AccountNumber = num
		}
	})
	form.AddInputField("Routing Number", "", 20, func(textToCheck string, lastChar rune) bool {
		return lastChar >= '0' && lastChar <= '9'
	}, func(text string) {
		if num, err := strconv.Atoi(text); err == nil {
			account.RTN = num
		}
	})
	form.AddInputField("Starting Check Number", "1001", 10, func(textToCheck string, lastChar rune) bool {
		return lastChar >= '0' && lastChar <= '9'
	}, func(text string) {
		if num, err := strconv.Atoi(text); err == nil {
			account.LastCheckNum = num - 1 // Store as last used, not next
		}
	})

	form.AddButton("Save", func() {
		if account.AccountName == "" || account.NameOnAccount == "" || account.BankName == "" {
			a.showError("Account Name, Name on Account, and Bank Name are required")
			return
		}
		if account.AccountNumber == 0 || account.RTN == 0 {
			a.showError("Account Number and Routing Number are required")
			return
		}

		err := a.saveAccount(account)
		if err != nil {
			a.showError(fmt.Sprintf("Error saving account: %v", err))
			return
		}
		a.showMainMenu()
	})

	form.AddButton("Cancel", func() {
		a.showMainMenu()
	})

	// Set up proper navigation with Tab key
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			a.showMainMenu()
			return nil
		case tcell.KeyTab:
			// Let the form handle Tab navigation
			return event
		case tcell.KeyBacktab: // Shift+Tab
			// Let the form handle reverse Tab navigation
			return event
		}
		return event
	})

	a.main.Clear()
	a.main.AddItem(form, 0, 1, true)
	a.app.SetFocus(form)
}

func (a *App) showWriteCheckForm() {
	accounts, err := a.getAccounts()
	if err != nil {
		a.showError(fmt.Sprintf("Error loading accounts: %v", err))
		return
	}

	if len(accounts) == 0 {
		a.showError("No accounts found. Please add an account first.")
		return
	}

	form := tview.NewForm()
	form.SetBorder(true).SetTitle("Write New Check")

	var selectedAccount Account
	var payee, memo string
	var amount float64

	// Account dropdown
	accountOptions := make([]string, len(accounts))
	for i, acc := range accounts {
		accountOptions[i] = fmt.Sprintf("%s (%s)", acc.AccountName, acc.NameOnAccount)
	}

	form.AddDropDown("Account", accountOptions, 0, func(option string, optionIndex int) {
		selectedAccount = accounts[optionIndex]
	})

	if len(accounts) > 0 {
		selectedAccount = accounts[0]
	}

	form.AddInputField("Payee", "", 40, nil, func(text string) {
		payee = text
	})

	form.AddInputField("Amount", "", 15, func(textToCheck string, lastChar rune) bool {
		return (lastChar >= '0' && lastChar <= '9') || lastChar == '.'
	}, func(text string) {
		if amt, err := strconv.ParseFloat(text, 64); err == nil {
			amount = amt
		}
	})

	form.AddInputField("Memo", "", 40, nil, func(text string) {
		memo = text
	})

	form.AddButton("Print Check", func() {
		if payee == "" {
			a.showError("Payee is required")
			return
		}
		if amount <= 0 {
			a.showError("Amount must be greater than 0")
			return
		}

		// Get next check number
		nextCheckNum := selectedAccount.LastCheckNum + 1

		// Generate filename
		filename := fmt.Sprintf("check_%s_%d.pdf",
			strings.ReplaceAll(selectedAccount.AccountName, " ", "_"),
			nextCheckNum)

		// Write the check
		err := writeCheck(selectedAccount, nextCheckNum, payee, memo, amount, filename)
		if err != nil {
			a.showError(fmt.Sprintf("Error writing check: %v", err))
			return
		}

		// Update last check number
		err = a.updateLastCheckNum(selectedAccount.ID, nextCheckNum)
		if err != nil {
			a.showError(fmt.Sprintf("Error updating check number: %v", err))
			return
		}

		// Log the check
		err = a.logCheck(selectedAccount.ID, nextCheckNum, payee, memo, amount, filename)
		if err != nil {
			a.showError(fmt.Sprintf("Error logging check: %v", err))
			return
		}

		a.showSuccess(fmt.Sprintf("Check #%d written successfully!\nSaved as: %s", nextCheckNum, filename))
	})

	form.AddButton("Cancel", func() {
		a.showMainMenu()
	})

	// Set up proper navigation with Tab key
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			a.showMainMenu()
			return nil
		case tcell.KeyTab:
			// Let the form handle Tab navigation
			return event
		case tcell.KeyBacktab: // Shift+Tab
			// Let the form handle reverse Tab navigation
			return event
		}
		return event
	})

	a.main.Clear()
	a.main.AddItem(form, 0, 1, true)
	a.app.SetFocus(form)
}

func (a *App) showListAccounts() {
	table := tview.NewTable()
	table.SetBorder(true).SetTitle("Account List")

	// Headers
	headers := []string{"Account Name", "Name on Account", "Bank Name", "Account #", "Routing #", "Last Check #"}
	for i, header := range headers {
		table.SetCell(0, i, tview.NewTableCell(header).
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignCenter))
	}

	// Get accounts
	accounts, err := a.getAccounts()
	if err != nil {
		a.showError(fmt.Sprintf("Error loading accounts: %v", err))
		return
	}

	// Populate table
	for row, account := range accounts {
		table.SetCell(row+1, 0, tview.NewTableCell(account.AccountName))
		table.SetCell(row+1, 1, tview.NewTableCell(account.NameOnAccount))
		table.SetCell(row+1, 2, tview.NewTableCell(account.BankName))

		// Show only last 4 digits of account number for security
		accountDisplay := fmt.Sprintf("****%04d", account.AccountNumber%10000)
		table.SetCell(row+1, 3, tview.NewTableCell(accountDisplay))

		table.SetCell(row+1, 4, tview.NewTableCell(fmt.Sprintf("%d", account.RTN)))
		table.SetCell(row+1, 5, tview.NewTableCell(fmt.Sprintf("%d", account.LastCheckNum)))
	}

	// Add navigation info
	info := tview.NewTextView()
	info.SetText("Press 'b' to go back to main menu")
	info.SetTextAlign(tview.AlignCenter)

	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	flex.AddItem(table, 0, 1, true)
	flex.AddItem(info, 1, 0, false)

	// Handle key events on the flex container
	flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'b' || event.Rune() == 'B' || event.Key() == tcell.KeyEsc {
			a.showMainMenu()
			return nil
		}
		return event
	})

	a.main.Clear()
	a.main.AddItem(flex, 0, 1, true)
	a.app.SetFocus(flex)
}

func (a *App) showCheckHistory() {
	table := tview.NewTable()
	table.SetBorder(true).SetTitle("Check History")

	// Headers
	headers := []string{"Date", "Account", "Check #", "Payee", "Amount", "Memo", "Filename"}
	for i, header := range headers {
		table.SetCell(0, i, tview.NewTableCell(header).
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignCenter))
	}

	// Get check history with account names
	rows, err := a.db.Query(`
		SELECT cl.date, a.account_name, cl.check_number, cl.payee, cl.amount, cl.memo, cl.filename
		FROM check_logs cl
		JOIN accounts a ON cl.account_id = a.id
		ORDER BY cl.date DESC
	`)
	if err != nil {
		a.showError(fmt.Sprintf("Error loading check history: %v", err))
		return
	}
	defer rows.Close()

	row := 1
	for rows.Next() {
		var date time.Time
		var accountName, payee, memo, filename string
		var checkNum int
		var amount float64

		err := rows.Scan(&date, &accountName, &checkNum, &payee, &amount, &memo, &filename)
		if err != nil {
			continue
		}

		table.SetCell(row, 0, tview.NewTableCell(date.Format("2006-01-02")))
		table.SetCell(row, 1, tview.NewTableCell(accountName))
		table.SetCell(row, 2, tview.NewTableCell(fmt.Sprintf("%d", checkNum)))
		table.SetCell(row, 3, tview.NewTableCell(payee))
		table.SetCell(row, 4, tview.NewTableCell(fmt.Sprintf("$%.2f", amount)))
		table.SetCell(row, 5, tview.NewTableCell(memo))
		table.SetCell(row, 6, tview.NewTableCell(filename))
		row++
	}

	// Add navigation info
	info := tview.NewTextView()
	info.SetText("Press 'b' to go back to main menu")
	info.SetTextAlign(tview.AlignCenter)

	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	flex.AddItem(table, 0, 1, true)
	flex.AddItem(info, 1, 0, false)

	// Handle key events on the flex container instead of just the table
	flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'b' || event.Rune() == 'B' || event.Key() == tcell.KeyEsc {
			a.showMainMenu()
			return nil
		}
		return event
	})

	a.main.Clear()
	a.main.AddItem(flex, 0, 1, true)
	a.app.SetFocus(flex) // Focus on the flex container, not just the table
}

func (a *App) showMainMenu() {
	menu := tview.NewList()
	menu.SetBorder(true).SetTitle("Check Printer - Main Menu")

	menu.AddItem("Write New Check", "Create and print a new check", 'w', func() {
		a.showWriteCheckForm()
	})

	menu.AddItem("Add Account", "Add a new bank account", 'a', func() {
		a.showAddAccountForm()
	})

	menu.AddItem("List Accounts", "View all bank accounts", 'l', func() {
		a.showListAccounts()
	})

	menu.AddItem("View Check History", "View all printed checks", 'h', func() {
		a.showCheckHistory()
	})

	menu.AddItem("Quit", "Exit the application", 'q', func() {
		a.app.Stop()
	})

	// Set up proper key handling
	menu.SetSelectedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		switch index {
		case 0:
			a.showWriteCheckForm()
		case 1:
			a.showAddAccountForm()
		case 2:
			a.showListAccounts()
		case 3:
			a.showCheckHistory()
		case 4:
			a.app.Stop()
		}
	})

	// Enable mouse and keyboard navigation
	menu.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			// Let the list handle Enter key
			return event
		case tcell.KeyEsc:
			a.app.Stop()
			return nil
		}

		// Handle shortcut keys
		switch event.Rune() {
		case 'w', 'W':
			a.showWriteCheckForm()
			return nil
		case 'a', 'A':
			a.showAddAccountForm()
			return nil
		case 'l', 'L':
			a.showListAccounts()
			return nil
		case 'h', 'H':
			a.showCheckHistory()
			return nil
		case 'q', 'Q':
			a.app.Stop()
			return nil
		}

		return event
	})

	a.main.Clear()
	a.main.AddItem(menu, 0, 1, true)
	a.app.SetFocus(menu)
}

func (a *App) showError(message string) {
	modal := tview.NewModal()
	modal.SetText(message)
	modal.AddButtons([]string{"OK"})
	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		a.app.SetRoot(a.main, true)
	})
	a.app.SetRoot(modal, true)
}

func (a *App) showSuccess(message string) {
	modal := tview.NewModal()
	modal.SetText(message)
	modal.AddButtons([]string{"OK"})
	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		a.app.SetRoot(a.main, true)
		a.showMainMenu()
	})
	a.app.SetRoot(modal, true)
}

func main() {
	// Initialize database
	db, err := initDB()
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Create the application
	app := &App{
		db:   db,
		app:  tview.NewApplication(),
		main: tview.NewFlex(),
	}

	// Enable mouse support
	app.app.EnableMouse(true)

	// Set up the main layout
	app.app.SetRoot(app.main, true)

	// Show main menu
	app.showMainMenu()

	// Run the application
	if err := app.app.Run(); err != nil {
		log.Fatal("Error running application:", err)
	}
}
