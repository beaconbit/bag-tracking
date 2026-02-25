package database

import (
	"fmt"
	"log"
	"strings"
	"time"
)

type WorkOrder struct {
	UUID                string    `json:"uuid"`
	BagNumber           int       `json:"bag_number"`
	Date                time.Time `json:"date"`
	WorkRequestOrder    bool      `json:"work_request_order"`
	WorkCompletionOrder bool      `json:"work_completion_order"`
	MainTrolley         bool      `json:"main_trolley"`
	SecondaryTrolley    bool      `json:"secondary_trolley"`
	Clutch              bool      `json:"clutch"`
	LargeRope           bool      `json:"large_rope"`
	SmallRope           bool      `json:"small_rope"`
	BagWeight           bool      `json:"bag_weight"`
	RopeWeight          bool      `json:"rope_weight"`
	Frame               bool      `json:"frame"`
	Carabina            bool      `json:"carabina"`
	Fabric              bool      `json:"fabric"`
}

// CreateWorkOrder inserts a new work order for the specified rail.
// The work_request_order is set to true, work_completion_order to false,
// and date to current UTC time.
func CreateWorkOrder(railType string, bagNumber int, flags map[string]bool) error {
	tableName, err := getWorkOrderTableName(railType)
	if err != nil {
		return err
	}

	// Ensure required flags are set (default to request fix)
	if _, exists := flags["work_request_order"]; !exists {
		flags["work_request_order"] = true
	}
	if _, exists := flags["work_completion_order"]; !exists {
		flags["work_completion_order"] = false
	}

	// Build column list and placeholders
	columns := []string{"uuid", "bag_number", "date"}
	placeholders := []string{"lower(hex(randomblob(16)))", "?", "?"}
	values := []interface{}{bagNumber, time.Now().UTC()}

	// Add boolean columns
	boolColumns := []string{
		"work_request_order",
		"work_completion_order",
		"main_trolley",
		"secondary_trolley",
		"clutch",
		"large_rope",
		"small_rope",
		"bag_weight",
		"rope_weight",
		"frame",
		"carabina",
		"fabric",
	}

	for _, col := range boolColumns {
		columns = append(columns, col)
		placeholders = append(placeholders, "?")
		val, ok := flags[col]
		if !ok {
			val = false // default
		}
		values = append(values, val)
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	_, err = DB.Exec(query, values...)
	if err != nil {
		return fmt.Errorf("failed to insert work order: %w", err)
	}

	log.Printf("Created work order for bag %d in %s", bagNumber, tableName)
	return nil
}

// GetWorkOrders retrieves all work orders for the specified rail, ordered by date descending.
func GetWorkOrders(railType string) ([]WorkOrder, error) {
	tableName, err := getWorkOrderTableName(railType)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf(`
		SELECT uuid, bag_number, date,
		       work_request_order, work_completion_order,
		       main_trolley, secondary_trolley, clutch,
		       large_rope, small_rope, bag_weight, rope_weight,
		       frame, carabina, fabric
		FROM %s
		ORDER BY date DESC
	`, tableName)

	rows, err := DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query work orders: %w", err)
	}
	defer rows.Close()

	var orders []WorkOrder
	for rows.Next() {
		var wo WorkOrder
		err := rows.Scan(
			&wo.UUID,
			&wo.BagNumber,
			&wo.Date,
			&wo.WorkRequestOrder,
			&wo.WorkCompletionOrder,
			&wo.MainTrolley,
			&wo.SecondaryTrolley,
			&wo.Clutch,
			&wo.LargeRope,
			&wo.SmallRope,
			&wo.BagWeight,
			&wo.RopeWeight,
			&wo.Frame,
			&wo.Carabina,
			&wo.Fabric,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan work order row: %w", err)
		}
		orders = append(orders, wo)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return orders, nil
}

// getWorkOrderTableName returns the database table name for a rail type.
func getWorkOrderTableName(railType string) (string, error) {
	switch strings.ToLower(railType) {
	case "clean":
		return "clean_bag_work_order", nil
	case "ironer":
		return "ironer_bag_work_order", nil
	case "sorting":
		return "sorting_bag_work_order", nil
	default:
		return "", fmt.Errorf("invalid rail type: %s", railType)
	}
}
