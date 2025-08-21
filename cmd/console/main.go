package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"go-finance-advisor/internal/application"
	"go-finance-advisor/internal/domain"

	"github.com/glebarez/sqlite"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gorm.io/gorm"
)

type App struct {
	userSvc      *application.UserService
	txSvc        *application.TransactionService
	advisorSvc   *application.AdvisorService
	analyticsSvc *application.AnalyticsService
	budgetSvc    *application.BudgetService
	categorySvc  *application.CategoryService
	reportsSvc   *application.ReportsService
	exportSvc    *application.ExportService
	currentUser  *domain.User
	reader       *bufio.Reader
}

func main() {
	printHeader()
	
	// Initialize database
	db := initializeDatabase()
	if db == nil {
		return
	}
	
	// Initialize application
	app := initializeApp(db)
	
	// Start main application loop
	app.run()
}

func printHeader() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("    PERSONAL FINANCE ADVISOR - CONSOLE APPLICATION")
	fmt.Println("" + strings.Repeat("=", 60))
	fmt.Println("    Professional Financial Management System")
	fmt.Println("" + strings.Repeat("-", 60))
}

func initializeDatabase() *gorm.DB {
	fmt.Println("[INFO] Initializing database connection...")
	db, err := gorm.Open(sqlite.Open("finance.db"), &gorm.Config{})
	if err != nil {
		fmt.Printf("[ERROR] Database connection failed: %v\n", err)
		return nil
	}

	fmt.Println("[INFO] Running database migrations...")
	err = db.AutoMigrate(&domain.User{}, &domain.Transaction{}, &domain.Category{}, &domain.Budget{})
	if err != nil {
		fmt.Printf("[ERROR] Database migration failed: %v\n", err)
		return nil
	}
	
	fmt.Println("[SUCCESS] Database initialized successfully")
	return db
}

func initializeApp(db *gorm.DB) *App {
	fmt.Println("[INFO] Initializing application services...")
	
	// Initialize services
	userSvc := &application.UserService{DB: db}
	txSvc := &application.TransactionService{DB: db}
	advisorSvc := &application.AdvisorService{DB: db}
	analyticsSvc := &application.AnalyticsService{DB: db}
	budgetSvc := &application.BudgetService{DB: db}
	categorySvc := &application.CategoryService{DB: db}
	reportsSvc := &application.ReportsService{DB: db}
	exportSvc := &application.ExportService{DB: db}

	// Initialize default categories
	fmt.Println("[INFO] Setting up default categories...")
	err := categorySvc.InitializeDefaultCategories()
	if err != nil {
		fmt.Printf("[WARNING] Could not initialize default categories: %v\n", err)
	} else {
		fmt.Println("[SUCCESS] Default categories initialized")
	}

	return &App{
		userSvc:      userSvc,
		txSvc:        txSvc,
		advisorSvc:   advisorSvc,
		analyticsSvc: analyticsSvc,
		budgetSvc:    budgetSvc,
		categorySvc:  categorySvc,
		reportsSvc:   reportsSvc,
		exportSvc:    exportSvc,
		reader:       bufio.NewReader(os.Stdin),
	}
}

func (app *App) run() {
	fmt.Println("\n[INFO] Application started successfully!")
	fmt.Println("[INFO] Type 'help' at any time for assistance")
	
	// Main application loop
	for {
		if app.currentUser == nil {
			app.showLoginMenu()
		} else {
			app.showMainMenu()
		}
	}
}

func (app *App) showLoginMenu() {
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("                 LOGIN MENU")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println("  1. Login to Account")
	fmt.Println("  2. Create New Account")
	fmt.Println("  3. Exit Application")
	fmt.Println(strings.Repeat("-", 50))
	fmt.Print("Please select an option (1-3): ")

	choice, _ := app.reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		app.login()
	case "2":
		app.register()
	case "3":
		fmt.Println("\n[INFO] Thank you for using Personal Finance Advisor!")
		fmt.Println("[INFO] Goodbye!")
		os.Exit(0)
	default:
		fmt.Println("\n[ERROR] Invalid selection! Please choose 1, 2, or 3.")
	}
}

func (app *App) login() {
	fmt.Println("\n" + strings.Repeat("-", 30))
	fmt.Println("         USER LOGIN")
	fmt.Println(strings.Repeat("-", 30))
	
	fmt.Print("Email Address: ")
	email, _ := app.reader.ReadString('\n')
	email = strings.TrimSpace(email)

	fmt.Print("Password: ")
	password, _ := app.reader.ReadString('\n')
	password = strings.TrimSpace(password)

	// Authenticate user
	user, err := app.userSvc.Login(email, password)
	if err != nil {
		fmt.Printf("\n[ERROR] Login failed: %v\n", err)
		fmt.Println("[INFO] Please check your credentials and try again.")
		return
	}

	app.currentUser = user
	fmt.Printf("\n[SUCCESS] Welcome back, %s %s!\n", user.FirstName, user.LastName)
	fmt.Println("[INFO] Login successful. Redirecting to main menu...")
}

func (app *App) register() {
	fmt.Println("\n" + strings.Repeat("-", 35))
	fmt.Println("       CREATE NEW ACCOUNT")
	fmt.Println(strings.Repeat("-", 35))
	
	fmt.Print("First Name: ")
	firstName, _ := app.reader.ReadString('\n')
	firstName = strings.TrimSpace(firstName)

	fmt.Print("Last Name: ")
	lastName, _ := app.reader.ReadString('\n')
	lastName = strings.TrimSpace(lastName)

	fmt.Print("Email Address: ")
	email, _ := app.reader.ReadString('\n')
	email = strings.TrimSpace(email)

	fmt.Print("Password: ")
	password, _ := app.reader.ReadString('\n')
	password = strings.TrimSpace(password)

	_, err := app.userSvc.Register(email, password, firstName, lastName)
	if err != nil {
		fmt.Printf("\n[ERROR] Registration failed: %v\n", err)
		fmt.Println("[INFO] Please try again with different credentials.")
		return
	}

	fmt.Println("\n[SUCCESS] Account created successfully!")
	fmt.Println("[INFO] You can now login with your credentials.")
}

func (app *App) showMainMenu() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Printf("    MAIN MENU - Welcome %s %s\n", app.currentUser.FirstName, app.currentUser.LastName)
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  📊 TRANSACTIONS")
	fmt.Println("    1. Add New Transaction")
	fmt.Println("    2. View Transaction History")
	fmt.Println("")
	fmt.Println("  💰 BUDGET & PLANNING")
	fmt.Println("    3. Budget Management")
	fmt.Println("    4. Financial Reports")
	fmt.Println("")
	fmt.Println("  📈 ANALYSIS & INSIGHTS")
	fmt.Println("    5. Financial Analytics")
	fmt.Println("    6. Investment Advice")
	fmt.Println("")
	fmt.Println("  ⚙️  SETTINGS")
	fmt.Println("    7. Manage Categories")
	fmt.Println("    8. Logout")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Print("Please select an option (1-8): ")

	choice, _ := app.reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		app.addTransaction()
	case "2":
		app.listTransactions()
	case "3":
		app.budgetMenu()
	case "4":
		app.reportsMenu()
	case "5":
		app.analyticsMenu()
	case "6":
		app.investmentAdvice()
	case "7":
		app.categoryMenu()
	case "8":
		app.currentUser = nil
		fmt.Println("\n[INFO] Successfully logged out. Returning to login menu...")
	default:
		fmt.Println("\n[ERROR] Invalid selection! Please choose a number between 1-8.")
	}
}

func (app *App) addTransaction() {
	fmt.Println("\n" + strings.Repeat("-", 40))
	fmt.Println("         ADD NEW TRANSACTION")
	fmt.Println(strings.Repeat("-", 40))
	
	// List available categories
	categories, err := app.categorySvc.GetAllCategories()
	if err != nil {
		fmt.Printf("[ERROR] Could not retrieve categories: %v\n", err)
		return
	}
	
	fmt.Println("\n📂 Available Categories:")
	fmt.Println(strings.Repeat("-", 30))
	for _, cat := range categories {
		fmt.Printf("  %d. %s (%s)\n", cat.ID, cat.Name, cat.Type)
	}
	fmt.Println(strings.Repeat("-", 30))
	
	fmt.Print("Category ID: ")
	categoryIDStr, _ := app.reader.ReadString('\n')
	categoryID, err := strconv.Atoi(strings.TrimSpace(categoryIDStr))
	if err != nil {
		fmt.Println("[ERROR] Invalid category ID! Please enter a valid number.")
		return
	}
	
	fmt.Print("Amount: $")
	amountStr, _ := app.reader.ReadString('\n')
	amount, err := strconv.ParseFloat(strings.TrimSpace(amountStr), 64)
	if err != nil {
		fmt.Println("[ERROR] Invalid amount! Please enter a valid number.")
		return
	}
	
	fmt.Print("Description: ")
	description, _ := app.reader.ReadString('\n')
	description = strings.TrimSpace(description)
	
	fmt.Print("Type (income/expense): ")
	transactionType, _ := app.reader.ReadString('\n')
	transactionType = strings.TrimSpace(transactionType)
	
	if transactionType != "income" && transactionType != "expense" {
		fmt.Println("[ERROR] Invalid transaction type! Please enter 'income' or 'expense'.")
		return
	}
	
	transaction := &domain.Transaction{
		UserID:      app.currentUser.ID,
		CategoryID:  uint(categoryID),
		Amount:      amount,
		Description: description,
		Type:        transactionType,
		Date:        time.Now(),
	}
	
	err = app.txSvc.Create(transaction)
	if err != nil {
		fmt.Printf("[ERROR] Could not add transaction: %v\n", err)
		return
	}
	
	fmt.Println("\n[SUCCESS] ✅ Transaction added successfully!")
	fmt.Printf("[INFO] Added %s of $%.2f for %s\n", transactionType, amount, description)
}

func (app *App) listTransactions() {
	fmt.Println("\n" + strings.Repeat("-", 50))
	fmt.Println("           TRANSACTION HISTORY")
	fmt.Println(strings.Repeat("-", 50))
	
	transactions, err := app.txSvc.List(app.currentUser.ID)
	if err != nil {
		fmt.Printf("[ERROR] Could not retrieve transactions: %v\n", err)
		return
	}
	
	if len(transactions) == 0 {
		fmt.Println("\n📝 No transactions found.")
		fmt.Println("[INFO] Start by adding your first transaction!")
		return
	}
	
	fmt.Printf("\n📊 Found %d transaction(s):\n\n", len(transactions))
	fmt.Printf("%-4s %-12s %-10s %-20s %-12s\n", "ID", "Date", "Type", "Description", "Amount")
	fmt.Println(strings.Repeat("-", 60))
	
	for _, tx := range transactions {
		typeIcon := "💰"
		if tx.Type == "expense" {
			typeIcon = "💸"
		}
		fmt.Printf("%-4d %-12s %-10s %-20s %s$%.2f\n",
			tx.ID, 
			tx.Date.Format("2006-01-02"), 
			tx.Type, 
			tx.Description, 
			typeIcon,
			tx.Amount)
	}
	fmt.Println(strings.Repeat("-", 60))
}

func (app *App) budgetMenu() {
	fmt.Println("\n" + strings.Repeat("-", 40))
	fmt.Println("         BUDGET MANAGEMENT")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Println("  1. Create New Budget")
	fmt.Println("  2. View Budget Overview")
	fmt.Println("  3. Budget Summary")
	fmt.Println("  4. Return to Main Menu")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Print("Please select an option (1-4): ")

	choice, _ := app.reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		app.createBudget()
	case "2":
		app.listBudgets()
	case "3":
		app.budgetSummary()
	case "4":
		return
	default:
		fmt.Println("[ERROR] Invalid selection! Please choose 1, 2, 3, or 4.")
	}
}

func (app *App) createBudget() {
	fmt.Println("\n" + strings.Repeat("-", 35))
	fmt.Println("        CREATE NEW BUDGET")
	fmt.Println(strings.Repeat("-", 35))
	
	// List available categories
	categories, err := app.categorySvc.GetAllCategories()
	if err != nil {
		fmt.Printf("[ERROR] Could not retrieve categories: %v\n", err)
		return
	}
	
	fmt.Println("\n📂 Available Categories:")
	fmt.Println(strings.Repeat("-", 25))
	for _, cat := range categories {
		fmt.Printf("  %d. %s (%s)\n", cat.ID, cat.Name, cat.Type)
	}
	fmt.Println(strings.Repeat("-", 25))
	
	fmt.Print("Category ID: ")
	categoryIDStr, _ := app.reader.ReadString('\n')
	categoryID, err := strconv.Atoi(strings.TrimSpace(categoryIDStr))
	if err != nil {
		fmt.Println("[ERROR] Invalid category ID! Please enter a valid number.")
		return
	}

	fmt.Print("Budget Amount: $")
	amountStr, _ := app.reader.ReadString('\n')
	amount, err := strconv.ParseFloat(strings.TrimSpace(amountStr), 64)
	if err != nil {
		fmt.Println("[ERROR] Invalid amount! Please enter a valid number.")
		return
	}

	fmt.Print("Period (monthly/weekly/yearly): ")
	periodStr, _ := app.reader.ReadString('\n')
	period := strings.TrimSpace(periodStr)
	
	if period != "monthly" && period != "weekly" && period != "yearly" {
		fmt.Println("[ERROR] Invalid period! Please enter 'monthly', 'weekly', or 'yearly'.")
		return
	}

	budget := &domain.Budget{
		UserID:     app.currentUser.ID,
		CategoryID: uint(categoryID),
		Amount:     amount,
		Period:     period,
		StartDate:  time.Now(),
		EndDate:    time.Now().AddDate(0, 1, 0), // Default to 1 month
	}

	err = app.budgetSvc.CreateBudget(budget)
	if err != nil {
		fmt.Printf("[ERROR] Could not create budget: %v\n", err)
		return
	}

	fmt.Println("\n[SUCCESS] ✅ Budget created successfully!")
	fmt.Printf("[INFO] Created %s budget of $%.2f for category ID %d\n", period, amount, categoryID)
}

func (app *App) listBudgets() {
	fmt.Println("\n" + strings.Repeat("-", 50))
	fmt.Println("              BUDGET OVERVIEW")
	fmt.Println(strings.Repeat("-", 50))
	
	budgets, err := app.budgetSvc.GetBudgetsByUser(app.currentUser.ID)
	if err != nil {
		fmt.Printf("[ERROR] Could not retrieve budgets: %v\n", err)
		return
	}

	if len(budgets) == 0 {
		fmt.Println("\n💰 No budgets found.")
		fmt.Println("[INFO] Create your first budget to start tracking expenses!")
		return
	}

	fmt.Printf("\n📊 Found %d budget(s):\n\n", len(budgets))
	fmt.Printf("%-4s %-15s %-10s %-10s %-10s %-12s\n", "ID", "Category", "Period", "Budget", "Spent", "Remaining")
	fmt.Println(strings.Repeat("-", 70))

	for _, budget := range budgets {
		remaining := budget.Amount - budget.Spent
		status := "✅"
		if budget.Spent > budget.Amount {
			status = "❌"
		} else if budget.Spent > budget.Amount*0.8 {
			status = "⚠️"
		}
		
		fmt.Printf("%-4d %-15s %-10s $%-9.2f $%-8.2f $%-10.2f %s\n",
			budget.ID,
			budget.Category.Name,
			budget.Period,
			budget.Amount,
			budget.Spent,
			remaining,
			status)
	}
	fmt.Println(strings.Repeat("-", 70))
}

func (app *App) budgetSummary() {
	fmt.Println("\n" + strings.Repeat("-", 40))
	fmt.Println("           BUDGET SUMMARY")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Println("\n[INFO] This feature is under development...")
	fmt.Println("[INFO] Coming soon: Detailed budget analysis and insights!")
}

func (app *App) reportsMenu() {
	fmt.Println("\n" + strings.Repeat("-", 40))
	fmt.Println("           FINANCIAL REPORTS")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Println("  1. Monthly Report")
	fmt.Println("  2. Yearly Report")
	fmt.Println("  3. Category Analysis")
	fmt.Println("  4. Export Reports")
	fmt.Println("  5. Return to Main Menu")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Print("Please select an option (1-5): ")

	choice, _ := app.reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		app.monthlyReport()
	case "2":
		app.yearlyReport()
	case "3":
		fmt.Println("\n[INFO] Category Analysis feature is under development...")
	case "4":
		fmt.Println("\n[INFO] Export Reports feature is under development...")
	case "5":
		return
	default:
		fmt.Println("[ERROR] Invalid selection! Please choose 1, 2, 3, 4, or 5.")
	}
}

func (app *App) monthlyReport() {
	fmt.Println("\n" + strings.Repeat("-", 35))
	fmt.Println("        MONTHLY REPORT")
	fmt.Println(strings.Repeat("-", 35))
	fmt.Println("\n[INFO] Monthly Report feature is under development...")
	fmt.Println("[INFO] Coming soon: Detailed monthly financial analysis!")
}

func (app *App) yearlyReport() {
	fmt.Println("\n" + strings.Repeat("-", 35))
	fmt.Println("         YEARLY REPORT")
	fmt.Println(strings.Repeat("-", 35))
	fmt.Println("\n[INFO] Yearly Report feature is under development...")
	fmt.Println("[INFO] Coming soon: Comprehensive yearly financial overview!")
}

func (app *App) analyticsMenu() {
	fmt.Println("\n" + strings.Repeat("-", 45))
	fmt.Println("           FINANCIAL ANALYTICS")
	fmt.Println(strings.Repeat("-", 45))
	fmt.Println("  1. Income-Expense Analysis")
	fmt.Println("  2. Category Analysis")
	fmt.Println("  3. Dashboard Summary")
	fmt.Println("  4. Return to Main Menu")
	fmt.Println(strings.Repeat("-", 45))
	fmt.Print("Please select an option (1-4): ")

	choice, _ := app.reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		app.incomeExpenseAnalysis()
	case "2":
		app.categoryAnalysis()
	case "3":
		app.dashboardSummary()
	case "4":
		return
	default:
		fmt.Println("[ERROR] Invalid selection! Please choose 1, 2, 3, or 4.")
	}
}

func (app *App) incomeExpenseAnalysis() {
	fmt.Println("\n" + strings.Repeat("-", 45))
	fmt.Println("        INCOME-EXPENSE ANALYSIS")
	fmt.Println(strings.Repeat("-", 45))
	
	// Get basic analytics
	transactions, err := app.txSvc.List(app.currentUser.ID)
	if err != nil {
		fmt.Printf("[ERROR] Could not retrieve transactions: %v\n", err)
		return
	}
	
	if len(transactions) == 0 {
		fmt.Println("\n📊 No transaction data available for analysis.")
		fmt.Println("[INFO] Add some transactions to see your financial insights!")
		return
	}
	
	// Calculate basic metrics
	var totalIncome, totalExpenses float64
	incomeCount, expenseCount := 0, 0
	
	for _, tx := range transactions {
		if tx.Type == "income" {
			totalIncome += tx.Amount
			incomeCount++
		} else if tx.Type == "expense" {
			totalExpenses += tx.Amount
			expenseCount++
		}
	}
	
	netWorth := totalIncome - totalExpenses
	savingsRate := 0.0
	if totalIncome > 0 {
		savingsRate = ((totalIncome - totalExpenses) / totalIncome) * 100
	}
	
	fmt.Println("\n💰 FINANCIAL OVERVIEW")
	fmt.Println(strings.Repeat("-", 30))
	fmt.Printf("Total Income:     $%.2f (%d transactions)\n", totalIncome, incomeCount)
	fmt.Printf("Total Expenses:   $%.2f (%d transactions)\n", totalExpenses, expenseCount)
	fmt.Printf("Net Worth:        $%.2f\n", netWorth)
	fmt.Printf("Savings Rate:     %.1f%%\n", savingsRate)
	
	if netWorth > 0 {
		fmt.Println("\n✅ Status: Positive cash flow - Great job!")
	} else {
		fmt.Println("\n⚠️  Status: Negative cash flow - Consider reducing expenses")
	}
	
	fmt.Println("\n[INFO] Advanced analytics features coming soon!")
}

func (app *App) categoryAnalysis() {
	fmt.Println("\n" + strings.Repeat("-", 40))
	fmt.Println("         CATEGORY ANALYSIS")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Println("\n[INFO] Category Analysis feature is under development...")
	fmt.Println("[INFO] Coming soon: Detailed spending breakdown by category!")
}

func (app *App) dashboardSummary() {
	fmt.Println("\n" + strings.Repeat("-", 40))
	fmt.Println("         DASHBOARD SUMMARY")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Println("\n[INFO] Dashboard Summary feature is under development...")
	fmt.Println("[INFO] Coming soon: Complete financial dashboard overview!")
}

func (app *App) investmentAdvice() {
	fmt.Println("\n" + strings.Repeat("-", 45))
	fmt.Println("          INVESTMENT ADVISOR")
	fmt.Println(strings.Repeat("-", 45))
	fmt.Println("  1. Portfolio Recommendations")
	fmt.Println("  2. Risk Assessment")
	fmt.Println("  3. Market Analysis")
	fmt.Println("  4. Investment Calculator")
	fmt.Println(strings.Repeat("-", 45))
	fmt.Print("Please select an option (1-4): ")
	
	choice, _ := app.reader.ReadString('\n')
	choice = strings.TrimSpace(choice)
	
	switch choice {
	case "1":
		app.portfolioRecommendations()
	case "2":
		app.riskAssessment()
	case "3":
		fmt.Println("\n[INFO] Market Analysis feature is under development...")
	case "4":
		fmt.Println("\n[INFO] Investment Calculator feature is under development...")
	default:
		fmt.Println("[ERROR] Invalid selection! Please choose 1, 2, 3, or 4.")
	}
}

func (app *App) categoryMenu() {
	fmt.Println("\n" + strings.Repeat("-", 40))
	fmt.Println("         CATEGORY MANAGEMENT")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Println("  1. List Categories")
	fmt.Println("  2. Add Category")
	fmt.Println("  3. Return to Main Menu")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Print("Please select an option (1-3): ")

	choice, _ := app.reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		app.listCategories()
	case "2":
		app.addCategory()
	case "3":
		return
	default:
		fmt.Println("[ERROR] Invalid selection! Please choose 1, 2, or 3.")
	}
}

func (app *App) listCategories() {
	fmt.Println("\n" + strings.Repeat("-", 40))
	fmt.Println("         CATEGORY LIST")
	fmt.Println(strings.Repeat("-", 40))
	
	categories, err := app.categorySvc.GetAllCategories()
	if err != nil {
		fmt.Printf("[ERROR] Could not retrieve categories: %v\n", err)
		return
	}

	if len(categories) == 0 {
		fmt.Println("\n📂 No categories found.")
		fmt.Println("[INFO] Add your first category to get started!")
		return
	}

	fmt.Printf("\n📂 Available Categories (%d total):\n\n", len(categories))
	fmt.Printf("%-4s %-20s %-15s\n", "ID", "Name", "Type")
	fmt.Println(strings.Repeat("-", 40))

	for _, cat := range categories {
		typeIcon := "💰"
		if cat.Type == "expense" {
			typeIcon = "💸"
		}
		fmt.Printf("%-4d %-20s %s %-13s\n", cat.ID, cat.Name, typeIcon, cat.Type)
	}
	fmt.Println(strings.Repeat("-", 40))
}

func (app *App) addCategory() {
	fmt.Println("\n" + strings.Repeat("-", 35))
	fmt.Println("        ADD NEW CATEGORY")
	fmt.Println(strings.Repeat("-", 35))
	
	fmt.Print("Category Name: ")
	name, _ := app.reader.ReadString('\n')
	name = strings.TrimSpace(name)

	fmt.Print("Type (income/expense): ")
	catType, _ := app.reader.ReadString('\n')
	catType = strings.TrimSpace(catType)
	if catType != "income" && catType != "expense" {
		fmt.Println("[ERROR] Invalid type! Please enter 'income' or 'expense'.")
		return
	}

	category := &domain.Category{
		Name: name,
		Type: catType,
	}

	err := app.categorySvc.CreateCategory(category)
	if err != nil {
		fmt.Printf("[ERROR] Could not add category: %v\n", err)
		return
	}

	fmt.Println("\n[SUCCESS] ✅ Category added successfully!")
	fmt.Printf("[INFO] Added %s category: %s\n", catType, name)
}

func (app *App) portfolioRecommendations() {
	fmt.Println("\n" + strings.Repeat("-", 45))
	fmt.Println("       PORTFOLIO RECOMMENDATIONS")
	fmt.Println(strings.Repeat("-", 45))
	
	// Get user's risk tolerance
	riskLevel := app.currentUser.RiskTolerance
	if riskLevel == "" {
		riskLevel = "moderate" // Default
	}
	
	caser := cases.Title(language.English)
	fmt.Printf("\n👤 Your Risk Profile: %s\n\n", caser.String(riskLevel))
	
	switch riskLevel {
	case "conservative":
		fmt.Println("💼 CONSERVATIVE PORTFOLIO RECOMMENDATIONS:")
		fmt.Println(strings.Repeat("-", 40))
		fmt.Println("• 60% Bonds and Fixed Income")
		fmt.Println("• 30% Large-cap Stocks")
		fmt.Println("• 10% Cash and Money Market")
		fmt.Println("\n📊 Expected Annual Return: 4-6%")
		fmt.Println("⚠️  Risk Level: Low")
	case "aggressive":
		fmt.Println("🚀 AGGRESSIVE PORTFOLIO RECOMMENDATIONS:")
		fmt.Println(strings.Repeat("-", 40))
		fmt.Println("• 70% Growth Stocks")
		fmt.Println("• 20% International Stocks")
		fmt.Println("• 10% Alternative Investments")
		fmt.Println("\n📊 Expected Annual Return: 8-12%")
		fmt.Println("⚠️  Risk Level: High")
	default: // moderate
		fmt.Println("⚖️  MODERATE PORTFOLIO RECOMMENDATIONS:")
		fmt.Println(strings.Repeat("-", 40))
		fmt.Println("• 50% Diversified Stocks")
		fmt.Println("• 30% Bonds")
		fmt.Println("• 15% International Funds")
		fmt.Println("• 5% Cash")
		fmt.Println("\n📊 Expected Annual Return: 6-8%")
		fmt.Println("⚠️  Risk Level: Medium")
	}
	
	fmt.Println("\n💡 INVESTMENT TIPS:")
	fmt.Println("• Diversify across asset classes")
	fmt.Println("• Consider low-cost index funds")
	fmt.Println("• Rebalance portfolio quarterly")
	fmt.Println("• Invest consistently over time")
	
	fmt.Println("\n[INFO] This is general advice. Consult a financial advisor for personalized recommendations.")
}

func (app *App) riskAssessment() {
	fmt.Println("\n" + strings.Repeat("-", 40))
	fmt.Println("         RISK ASSESSMENT")
	fmt.Println(strings.Repeat("-", 40))
	
	fmt.Println("\n📋 Please answer the following questions to assess your risk tolerance:")
	fmt.Println("\n1. What is your investment time horizon?")
	fmt.Println("   a) Less than 3 years")
	fmt.Println("   b) 3-10 years")
	fmt.Println("   c) More than 10 years")
	
	fmt.Print("Your answer (a/b/c): ")
	timeHorizon, _ := app.reader.ReadString('\n')
	timeHorizon = strings.TrimSpace(strings.ToLower(timeHorizon))
	
	fmt.Println("\n2. How would you react to a 20% portfolio loss?")
	fmt.Println("   a) Sell everything immediately")
	fmt.Println("   b) Hold and wait for recovery")
	fmt.Println("   c) Buy more at lower prices")
	
	fmt.Print("Your answer (a/b/c): ")
	lossReaction, _ := app.reader.ReadString('\n')
	lossReaction = strings.TrimSpace(strings.ToLower(lossReaction))
	
	fmt.Println("\n3. What is your primary investment goal?")
	fmt.Println("   a) Capital preservation")
	fmt.Println("   b) Steady income")
	fmt.Println("   c) Long-term growth")
	
	fmt.Print("Your answer (a/b/c): ")
	investmentGoal, _ := app.reader.ReadString('\n')
	investmentGoal = strings.TrimSpace(strings.ToLower(investmentGoal))
	
	// Calculate risk score
	score := 0
	if timeHorizon == "c" { score += 2 }
	if timeHorizon == "b" { score += 1 }
	
	if lossReaction == "c" { score += 2 }
	if lossReaction == "b" { score += 1 }
	
	if investmentGoal == "c" { score += 2 }
	if investmentGoal == "b" { score += 1 }
	
	// Determine risk profile
	var riskProfile string
	switch {
	case score <= 2:
		riskProfile = "conservative"
	case score <= 4:
		riskProfile = "moderate"
	default:
		riskProfile = "aggressive"
	}
	
	fmt.Println("\n" + strings.Repeat("-", 40))
	fmt.Println("         ASSESSMENT RESULTS")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Printf("\n🎯 Your Risk Profile: %s\n", strings.ToUpper(riskProfile))
	fmt.Printf("📊 Risk Score: %d/6\n", score)
	
	switch riskProfile {
	case "conservative":
		fmt.Println("\n💼 CONSERVATIVE INVESTOR")
		fmt.Println("• Focus on capital preservation")
		fmt.Println("• Prefer stable, predictable returns")
		fmt.Println("• Low tolerance for volatility")
	case "aggressive":
		fmt.Println("\n🚀 AGGRESSIVE INVESTOR")
		fmt.Println("• Seek maximum long-term growth")
		fmt.Println("• Comfortable with high volatility")
		fmt.Println("• Long investment time horizon")
	default:
		fmt.Println("\n⚖️  MODERATE INVESTOR")
		fmt.Println("• Balance growth and stability")
		fmt.Println("• Moderate risk tolerance")
		fmt.Println("• Diversified approach")
	}
	
	fmt.Println("\n[INFO] Would you like to update your risk profile? (Feature coming soon)")
}