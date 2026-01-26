package tutorial

import (
	"testing"
	"time"

	gsql "github.com/donutnomad/gsql"
)

// ==================== Window Function Extended Tests ====================

// TestWindowFunctions_Extended tests more window function scenarios
func TestWindowFunctions_Extended(t *testing.T) {
	s := SalesRecordSchema
	setupTable(t, s.ModelType())
	db := getDB()

	// Insert test data with specific ordering
	records := []SalesRecord{
		{Region: "North", Salesperson: "Alice", Amount: 1000, SaleDate: time.Now()},
		{Region: "North", Salesperson: "Bob", Amount: 1500, SaleDate: time.Now()},
		{Region: "North", Salesperson: "Charlie", Amount: 1500, SaleDate: time.Now()}, // Same as Bob
		{Region: "South", Salesperson: "David", Amount: 2000, SaleDate: time.Now()},
		{Region: "South", Salesperson: "Eve", Amount: 1800, SaleDate: time.Now()},
		{Region: "South", Salesperson: "Frank", Amount: 1200, SaleDate: time.Now()},
	}
	if err := db.Create(&records).Error; err != nil {
		t.Fatalf("Failed to create records: %v", err)
	}

	t.Run("ROW_NUMBER with partition", func(t *testing.T) {
		// MySQL: SELECT *, ROW_NUMBER() OVER (PARTITION BY region ORDER BY amount DESC) AS rn
		rn := gsql.RowNumber().
			PartitionBy(s.Region).
			OrderBy(s.Amount.Desc()).
			As("rn")

		type Result struct {
			Region      string  `gorm:"column:region"`
			Salesperson string  `gorm:"column:salesperson"`
			Amount      float64 `gorm:"column:amount"`
			RN          int     `gorm:"column:rn"`
		}

		var results []Result
		err := gsql.Select(s.Region, s.Salesperson, s.Amount, rn).
			From(&s).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 6 {
			t.Errorf("Expected 6 results, got %d", len(results))
		}
		// Check that each region has 1, 2, 3 rankings
		northCounts := make(map[int]int)
		southCounts := make(map[int]int)
		for _, r := range results {
			if r.Region == "North" {
				northCounts[r.RN]++
			} else {
				southCounts[r.RN]++
			}
		}
		if northCounts[1] != 1 || northCounts[2] != 1 || northCounts[3] != 1 {
			t.Errorf("North should have exactly 1 of each rank 1,2,3: %v", northCounts)
		}
	})

	t.Run("RANK vs DENSE_RANK with ties", func(t *testing.T) {
		// MySQL: SELECT *,
		//               RANK() OVER (ORDER BY amount DESC) AS rank_num,
		//               DENSE_RANK() OVER (ORDER BY amount DESC) AS dense_rank_num
		rank := gsql.Rank().OrderBy(s.Amount.Desc()).As("rank_num")
		denseRank := gsql.DenseRank().OrderBy(s.Amount.Desc()).As("dense_rank_num")

		type Result struct {
			Salesperson  string  `gorm:"column:salesperson"`
			Amount       float64 `gorm:"column:amount"`
			RankNum      int     `gorm:"column:rank_num"`
			DenseRankNum int     `gorm:"column:dense_rank_num"`
		}

		var results []Result
		err := gsql.Select(s.Salesperson, s.Amount, rank, denseRank).
			From(&s).
			OrderBy(s.Amount.Desc()).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}

		// Check rank behavior with ties
		// Amount order: 2000, 1800, 1500, 1500, 1200, 1000
		// RANK:        1,    2,    3,    3,    5,    6
		// DENSE_RANK:  1,    2,    3,    3,    4,    5
		for _, r := range results {
			if r.Amount == 1500 {
				// Both Bob and Charlie have 1500, should be rank 3
				if r.RankNum != 3 {
					t.Errorf("Expected rank 3 for amount 1500, got %d", r.RankNum)
				}
				if r.DenseRankNum != 3 {
					t.Errorf("Expected dense_rank 3 for amount 1500, got %d", r.DenseRankNum)
				}
			}
			if r.Amount == 1200 {
				// After the tie at 3, RANK jumps to 5, DENSE_RANK goes to 4
				if r.RankNum != 5 {
					t.Errorf("Expected rank 5 for amount 1200, got %d", r.RankNum)
				}
				if r.DenseRankNum != 4 {
					t.Errorf("Expected dense_rank 4 for amount 1200, got %d", r.DenseRankNum)
				}
			}
		}
	})

	t.Run("Multiple window functions", func(t *testing.T) {
		// Combine multiple window functions
		rn := gsql.RowNumber().OrderBy(s.Amount.Desc()).As("rn")
		rank := gsql.Rank().OrderBy(s.Amount.Desc()).As("rank_num")

		type Result struct {
			Salesperson string  `gorm:"column:salesperson"`
			Amount      float64 `gorm:"column:amount"`
			RN          int     `gorm:"column:rn"`
			RankNum     int     `gorm:"column:rank_num"`
		}

		var results []Result
		err := gsql.Select(s.Salesperson, s.Amount, rn, rank).
			From(&s).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 6 {
			t.Errorf("Expected 6 results, got %d", len(results))
		}
	})
}
