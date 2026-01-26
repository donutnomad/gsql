package tutorial

import (
	"strings"
	"testing"
	"time"

	gsql "github.com/donutnomad/gsql"
)

// ==================== DateTime Function Tests ====================

// TestDateTimeFunctions_Extended tests extended datetime functions
func TestDateTimeFunctions_Extended(t *testing.T) {
	e := EmployeeSchema
	setupTable(t, e.ModelType())
	db := getDB()

	// Insert employees with specific dates
	employees := []Employee{
		{Name: "Alice", Email: "alice@test.com", Department: "IT", Salary: 80000,
			HireDate: time.Date(2020, 1, 15, 9, 0, 0, 0, time.UTC), BirthDate: time.Date(1990, 6, 20, 0, 0, 0, 0, time.UTC), IsActive: true},
		{Name: "Bob", Email: "bob@test.com", Department: "HR", Salary: 75000,
			HireDate: time.Date(2021, 6, 1, 9, 0, 0, 0, time.UTC), BirthDate: time.Date(1985, 12, 10, 0, 0, 0, 0, time.UTC), IsActive: true},
		{Name: "Charlie", Email: "charlie@test.com", Department: "IT", Salary: 90000,
			HireDate: time.Date(2019, 3, 10, 9, 0, 0, 0, time.UTC), BirthDate: time.Date(1988, 2, 28, 0, 0, 0, 0, time.UTC), IsActive: true},
	}
	if err := db.Create(&employees).Error; err != nil {
		t.Fatalf("Failed to create employees: %v", err)
	}

	t.Run("DayOfWeek and DayOfYear", func(t *testing.T) {
		// MySQL: SELECT DAYOFWEEK(employees.hire_date) AS dow, DAYOFYEAR(employees.hire_date) AS doy
		type Result struct {
			Name string `gorm:"column:name"`
			DOW  int    `gorm:"column:dow"`
			DOY  int    `gorm:"column:doy"`
		}

		var results []Result
		err := gsql.Select(
			e.Name,
			e.HireDate.DayOfWeek().As("dow"),
			e.HireDate.DayOfYear().As("doy"),
		).From(&e).
			Where(e.Name.Eq("Alice")).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
		// 2020-01-15 is Wednesday (4), day 15 of year
		if results[0].DOW != 4 {
			t.Errorf("Expected DOW 4 (Wednesday), got %d", results[0].DOW)
		}
		if results[0].DOY != 15 {
			t.Errorf("Expected DOY 15, got %d", results[0].DOY)
		}
	})

	t.Run("Quarter and Week functions", func(t *testing.T) {
		// MySQL: SELECT QUARTER(employees.hire_date) AS quarter, WEEK(employees.hire_date) AS week
		type Result struct {
			Name    string `gorm:"column:name"`
			Quarter int    `gorm:"column:quarter"`
			Week    int    `gorm:"column:week"`
		}

		var results []Result
		err := gsql.Select(
			e.Name,
			e.HireDate.Quarter().As("quarter"),
			e.HireDate.Week().As("week"),
		).From(&e).
			Where(e.Name.Eq("Bob")).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
		// 2021-06-01 is Q2
		if results[0].Quarter != 2 {
			t.Errorf("Expected quarter 2, got %d", results[0].Quarter)
		}
	})

	t.Run("LastDay function", func(t *testing.T) {
		// MySQL: SELECT LAST_DAY(employees.birth_date) AS last_day
		type Result struct {
			Name    string `gorm:"column:name"`
			LastDay string `gorm:"column:last_day"`
		}

		var results []Result
		err := gsql.Select(
			e.Name,
			e.BirthDate.LastDay().As("last_day"),
		).From(&e).
			Where(e.Name.Eq("Charlie")).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
		// 1988-02-28 -> last day of Feb 1988 is 29 (leap year)
		if !strings.Contains(results[0].LastDay, "1988-02-29") {
			t.Errorf("Expected last day 1988-02-29, got '%s'", results[0].LastDay)
		}
	})

	t.Run("DayName and MonthName", func(t *testing.T) {
		// MySQL: SELECT DAYNAME(employees.hire_date) AS day_name, MONTHNAME(employees.hire_date) AS month_name
		type Result struct {
			Name      string `gorm:"column:name"`
			DayName   string `gorm:"column:day_name"`
			MonthName string `gorm:"column:month_name"`
		}

		var results []Result
		err := gsql.Select(
			e.Name,
			e.HireDate.DayName().As("day_name"),
			e.HireDate.MonthName().As("month_name"),
		).From(&e).
			Where(e.Name.Eq("Alice")).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
		// 2020-01-15 is Wednesday, January
		if results[0].DayName != "Wednesday" {
			t.Errorf("Expected day name 'Wednesday', got '%s'", results[0].DayName)
		}
		if results[0].MonthName != "January" {
			t.Errorf("Expected month name 'January', got '%s'", results[0].MonthName)
		}
	})

	t.Run("AddInterval and DateDiff", func(t *testing.T) {
		// MySQL: SELECT DATE_ADD(employees.hire_date, INTERVAL 1 YEAR) AS next_year,
		//               DATEDIFF(NOW(), employees.hire_date) AS days_employed
		type Result struct {
			Name         string `gorm:"column:name"`
			NextYear     string `gorm:"column:next_year"`
			DaysEmployed int    `gorm:"column:days_employed"`
		}

		now := gsql.Sys.Now()
		var results []Result
		err := gsql.Select(
			e.Name,
			e.HireDate.AddInterval("1 YEAR").As("next_year"),
			now.DateDiff(e.HireDate).As("days_employed"),
		).From(&e).
			Where(e.Name.Eq("Alice")).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
		// 2020-01-15 + 1 YEAR = 2021-01-15
		if !strings.Contains(results[0].NextYear, "2021-01-15") {
			t.Errorf("Expected next year containing 2021-01-15, got '%s'", results[0].NextYear)
		}
		// Should have worked for many days
		if results[0].DaysEmployed < 1000 {
			t.Errorf("Expected > 1000 days employed, got %d", results[0].DaysEmployed)
		}
	})

	t.Run("TimestampDiff function", func(t *testing.T) {
		// MySQL: SELECT TIMESTAMPDIFF(MONTH, employees.hire_date, NOW()) AS months_employed
		type Result struct {
			Name           string `gorm:"column:name"`
			MonthsEmployed int64  `gorm:"column:months_employed"`
		}

		now := gsql.Sys.Now()
		var results []Result
		err := gsql.Select(
			e.Name,
			now.TimestampDiff("MONTH", e.HireDate).As("months_employed"),
		).From(&e).
			OrderBy(e.HireDate.Asc()).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 3 {
			t.Errorf("Expected 3 results, got %d", len(results))
		}
		// Charlie (2019) should have more months than Alice (2020) and Bob (2021)
		for _, r := range results {
			if r.MonthsEmployed < 12 {
				t.Errorf("Expected > 12 months employed for %s, got %d", r.Name, r.MonthsEmployed)
			}
		}
	})
}
