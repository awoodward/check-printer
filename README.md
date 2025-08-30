# Check Printer Application

A Go-based check printing application with SQLite database storage and a terminal-based user interface. This application allows you to manage multiple bank accounts and print professional checks as PDF files.

## Features

- **Multi-Account Management**: Store and manage multiple bank accounts with full details
- **Check Printing**: Generate professional PDF checks with proper formatting
- **Automatic Numbering**: Automatic check number increment and tracking per account
- **Check History**: Complete audit trail of all printed checks
- **Terminal UI**: Clean, keyboard-friendly interface using TView
- **MICR Encoding**: Proper MICR line formatting for bank processing
- **Input Validation**: Form validation for required fields and numeric inputs

## Prerequisites

Before running this application, you need:

1. **Go 1.18+** installed on your system
2. **Font files** for proper check formatting:
   - `security.ttf` - For the dollar amount display
   - `micr.ttf` - For the MICR line at the bottom of checks

### Paper Requirements

The application requires 8.5 x 11 inch preprinted check forms where the top third of the form is the check. 

## Installation

1. Clone the repository:
```bash
git clone https://github.com/awoodward/check-printer.git
cd check-printer
```

2. Initialize Go modules:
```bash
go mod init check-printer
```

3. Install dependencies:
```bash
go get github.com/rivo/tview
go get github.com/gdamore/tcell/v2
go get github.com/mattn/go-sqlite3
go get github.com/phpdave11/gofpdf
go get golang.org/x/text
```

4. Place the required font files (`security.ttf` and `micr.ttf`) in the project directory.

5. Build the application:
```bash
go build -o check-printer
```

## Usage

### Running the Application

```bash
./check-printer
```

### Navigation

**Main Menu:**
- Use arrow keys (↑/↓) + Enter to select options
- Or use letter shortcuts: 'w' (Write Check), 'a' (Add Account), 'h' (History), 'q' (Quit)
- Mouse clicks also work
- Press Escape to quit

**Forms:**
- **Tab**: Move to next field
- **Shift+Tab**: Move to previous field
- **Enter**: Activate buttons when focused
- **Escape**: Cancel and return to main menu

### Setting Up Accounts

1. Select "Add Account" from the main menu
2. Fill in the required information:
   - **Account Name**: A descriptive name for the account
   - **Name on Account**: The name printed on the check
   - **Bank Name**: The name of your bank
   - **Address 1**: Primary address line
   - **Address 2**: Secondary address line (city, state, zip)
   - **Account Number**: Your bank account number
   - **Routing Number**: Your bank's routing transit number
   - **Starting Check Number**: The first check number to use (default: 1001)

3. Click "Save" to store the account

### Writing Checks

1. Select "Write New Check" from the main menu
2. Choose the account from the dropdown
3. Enter:
   - **Payee**: Who the check is made out to
   - **Amount**: Dollar amount (numbers only, e.g., 123.45)
   - **Memo**: Optional memo line
4. Click "Print Check"

The application will:
- Generate a unique filename based on account and check number
- Create a PDF file with the check
- Increment the check number automatically
- Log the transaction in the database

### Viewing History

Select "View Check History" to see all printed checks with:
- Date printed
- Account used
- Check number
- Payee
- Amount
- Memo
- Generated filename

Press 'b' to return to the main menu from the history view.

## File Structure

```
check-printer/
├── main.go           # Main application code
├── checks.db         # SQLite database (created automatically)
├── security.ttf      # Security font for check amounts
├── micr.ttf         # MICR font for bank encoding
├── check_*.pdf      # Generated check files
└── README.md        # This file
```

## Database Schema

The application creates two tables:

### accounts
- `id`: Primary key
- `account_name`: Descriptive account name
- `name_on_account`: Name printed on checks
- `address1`, `address2`: Account holder address
- `account_number`: Bank account number
- `rtn`: Routing transit number
- `last_check_num`: Last used check number
- `bank_name`: Bank name

### check_logs
- `id`: Primary key
- `account_id`: Foreign key to accounts table
- `check_number`: Check number used
- `payee`: Check recipient
- `amount`: Check amount
- `memo`: Memo line
- `date`: When check was printed
- `filename`: Generated PDF filename

## Generated Files

Checks are saved as PDF files with the naming convention:
```
check_{AccountName}_{CheckNumber}.pdf
```

For example: `check_Business_Checking_1001.pdf`

## Security Considerations

- The application stores sensitive banking information in a local SQLite database
- Ensure the database file (`checks.db`) is properly secured
- Consider encrypting the database for production use
- Keep generated check files secure and dispose of them properly

## Troubleshooting

### Common Issues

1. **Missing fonts**: Ensure `security.ttf` and `micr.ttf` are in the executable directory
2. **Database permissions**: Make sure the application has write permissions in the directory
3. **SQLite driver**: The CGO-enabled SQLite driver requires a C compiler

### Build Issues

If you encounter build issues with the SQLite driver, you may need to:
- Install a C compiler (gcc, clang, or Visual Studio Build Tools on Windows)
- Use the `CGO_ENABLED=1` environment variable

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Dependencies

- [tview](https://github.com/rivo/tview) - Terminal UI framework
- [tcell](https://github.com/gdamore/tcell) - Terminal cell manipulation
- [go-sqlite3](https://github.com/mattn/go-sqlite3) - SQLite driver
- [gofpdf](https://github.com/phpdave11/gofpdf) - PDF generation
- [golang.org/x/text](https://golang.org/x/text) - Text processing utilities