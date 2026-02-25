package database

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
)

type BagRail struct {
	UUID         string `json:"uuid"`
	BagNumber    int    `json:"bag_number"`
	Fixed        bool   `json:"fixed"`
	InProduction bool   `json:"in_production"`
}

// AddOrUpdateBag adds or updates a bag in the specified rail table
func AddOrUpdateBag(railType, bagNumber string) error {
	// Convert bag number to integer, default to 0 if empty
	bagNum := 0
	if bagNumber != "" {
		_, err := fmt.Sscanf(bagNumber, "%d", &bagNum)
		if err != nil {
			bagNum = 0
		}
	}

	tableName, err := getRailTableName(railType)
	if err != nil {
		return err
	}

	// Check if bag already exists
	query := fmt.Sprintf("SELECT uuid FROM %s WHERE bag_number = ?", tableName)
	var existingUUID string
	err = DB.QueryRow(query, bagNum).Scan(&existingUUID)

	if err == nil {
		// Bag exists, update fixed and in_production to true
		updateQuery := fmt.Sprintf("UPDATE %s SET fixed = TRUE, in_production = TRUE WHERE bag_number = ?", tableName)
		_, err = DB.Exec(updateQuery, bagNum)
		if err != nil {
			return fmt.Errorf("failed to update bag: %w", err)
		}
		log.Printf("Updated bag %d in %s", bagNum, tableName)
		return nil
	} else if err == sql.ErrNoRows {
		// Bag doesn't exist, insert new row
		// SQLite will generate UUID automatically with ROWID, but we need explicit UUID
		// Use SQLite's built-in randomblob and hex functions to generate a UUID-like string
		insertQuery := fmt.Sprintf(`
			INSERT INTO %s (uuid, bag_number, fixed, in_production)
			VALUES (lower(hex(randomblob(16))), ?, TRUE, TRUE)
		`, tableName)
		_, err = DB.Exec(insertQuery, bagNum)
		if err != nil {
			return fmt.Errorf("failed to insert bag: %w", err)
		}
		log.Printf("Inserted new bag %d in %s", bagNum, tableName)

		// If this is a new non-zero bag, check if we should delete an anonymous bag
		if bagNum != 0 {
			if err := deleteOneAnonymousBag(railType); err != nil {
				log.Printf("Warning: failed to delete anonymous bag: %v", err)
				// Continue anyway, the main operation succeeded
			}
		}
		return nil
	} else {
		// Some other error
		return fmt.Errorf("failed to check existing bag: %w", err)
	}
}

// GetBags retrieves all bags from the specified rail table
func GetBags(railType string) ([]BagRail, error) {
	tableName, err := getRailTableName(railType)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("SELECT uuid, bag_number, fixed, in_production FROM %s ORDER BY bag_number", tableName)
	rows, err := DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query bags: %w", err)
	}
	defer rows.Close()

	var bags []BagRail
	for rows.Next() {
		var bag BagRail
		err := rows.Scan(&bag.UUID, &bag.BagNumber, &bag.Fixed, &bag.InProduction)
		if err != nil {
			return nil, fmt.Errorf("failed to scan bag row: %w", err)
		}
		bags = append(bags, bag)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return bags, nil
}

// RemoveBag deactivates a bag in the specified rail table by setting fixed and in_production to false
func RemoveBag(railType, bagNumber string) error {
	// Convert bag number to integer, return error if empty or invalid
	if bagNumber == "" {
		return fmt.Errorf("bag number is required")
	}
	bagNum := 0
	_, err := fmt.Sscanf(bagNumber, "%d", &bagNum)
	if err != nil {
		return fmt.Errorf("invalid bag number")
	}

	tableName, err := getRailTableName(railType)
	if err != nil {
		return err
	}

	// Check if bag exists
	query := fmt.Sprintf("SELECT uuid FROM %s WHERE bag_number = ?", tableName)
	var existingUUID string
	err = DB.QueryRow(query, bagNum).Scan(&existingUUID)

	if err == nil {
		// Bag exists, update fixed and in_production to false
		updateQuery := fmt.Sprintf("UPDATE %s SET fixed = FALSE, in_production = FALSE WHERE bag_number = ?", tableName)
		_, err = DB.Exec(updateQuery, bagNum)
		if err != nil {
			return fmt.Errorf("failed to deactivate bag: %w", err)
		}
		log.Printf("Deactivated bag %d in %s (set fixed and in_production to false)", bagNum, tableName)
		return nil
	} else if err == sql.ErrNoRows {
		// Bag doesn't exist
		return fmt.Errorf("bag not found")
	} else {
		// Some other error
		return fmt.Errorf("failed to check existing bag: %w", err)
	}
}

// RemoveAnonymousBag creates an anonymous bag with bag number 0 (deactivated state)
func RemoveAnonymousBag(railType string) error {
	tableName, err := getRailTableName(railType)
	if err != nil {
		return err
	}

	// Insert a new anonymous bag with bag number 0, fixed and in_production set to false
	insertQuery := fmt.Sprintf(`
		INSERT INTO %s (uuid, bag_number, fixed, in_production)
		VALUES (lower(hex(randomblob(16))), 0, FALSE, FALSE)
	`, tableName)
	_, err = DB.Exec(insertQuery)
	if err != nil {
		return fmt.Errorf("failed to create anonymous bag: %w", err)
	}
	log.Printf("Created anonymous bag (bag number 0) in %s", tableName)
	return nil
}

// deleteOneAnonymousBag deletes one anonymous bag (bag number 0) if any exist
func deleteOneAnonymousBag(railType string) error {
	tableName, err := getRailTableName(railType)
	if err != nil {
		return err
	}

	// Find one anonymous bag to delete
	query := fmt.Sprintf("SELECT uuid FROM %s WHERE bag_number = 0 LIMIT 1", tableName)
	var uuid string
	err = DB.QueryRow(query).Scan(&uuid)

	if err == sql.ErrNoRows {
		// No anonymous bags to delete
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to find anonymous bag: %w", err)
	}

	// Delete the found anonymous bag
	deleteQuery := fmt.Sprintf("DELETE FROM %s WHERE uuid = ?", tableName)
	_, err = DB.Exec(deleteQuery, uuid)
	if err != nil {
		return fmt.Errorf("failed to delete anonymous bag: %w", err)
	}
	log.Printf("Deleted anonymous bag %s from %s", uuid, tableName)
	return nil
}

// DeactivateBagIfExists sets fixed = FALSE, in_production = FALSE for a bag if it exists in the rail table.
// If the bag does not exist, no error is returned (silent success).
func DeactivateBagIfExists(railType string, bagNumber int) error {
	tableName, err := getRailTableName(railType)
	if err != nil {
		return err
	}

	// Check if bag exists
	query := fmt.Sprintf("SELECT uuid FROM %s WHERE bag_number = ?", tableName)
	var existingUUID string
	err = DB.QueryRow(query, bagNumber).Scan(&existingUUID)

	if err == nil {
		// Bag exists, update fixed and in_production to false
		updateQuery := fmt.Sprintf("UPDATE %s SET fixed = FALSE, in_production = FALSE WHERE bag_number = ?", tableName)
		_, err = DB.Exec(updateQuery, bagNumber)
		if err != nil {
			return fmt.Errorf("failed to deactivate bag: %w", err)
		}
		log.Printf("Deactivated bag %d in %s (set fixed and in_production to false)", bagNumber, tableName)
		return nil
	} else if err == sql.ErrNoRows {
		// Bag doesn't exist - that's fine, just return success
		return nil
	} else {
		// Some other error
		return fmt.Errorf("failed to check existing bag: %w", err)
	}
}

// MarkBagAsFixedIfExists sets fixed = TRUE, in_production = FALSE for a bag if it exists in the rail table.
// If the bag does not exist, no error is returned (silent success).
func MarkBagAsFixedIfExists(railType string, bagNumber int) error {
	tableName, err := getRailTableName(railType)
	if err != nil {
		return err
	}

	// Check if bag exists
	query := fmt.Sprintf("SELECT uuid FROM %s WHERE bag_number = ?", tableName)
	var existingUUID string
	err = DB.QueryRow(query, bagNumber).Scan(&existingUUID)

	if err == nil {
		// Bag exists, update fixed to true and in_production to false
		updateQuery := fmt.Sprintf("UPDATE %s SET fixed = TRUE, in_production = FALSE WHERE bag_number = ?", tableName)
		_, err = DB.Exec(updateQuery, bagNumber)
		if err != nil {
			return fmt.Errorf("failed to mark bag as fixed: %w", err)
		}
		log.Printf("Marked bag %d as fixed in %s (set fixed = TRUE, in_production = FALSE)", bagNumber, tableName)
		return nil
	} else if err == sql.ErrNoRows {
		// Bag doesn't exist - that's fine, just return success
		return nil
	} else {
		// Some other error
		return fmt.Errorf("failed to check existing bag: %w", err)
	}
}

// getRailTableName returns the database table name for a rail type
func getRailTableName(railType string) (string, error) {
	switch strings.ToLower(railType) {
	case "clean":
		return "clean_bag_rail", nil
	case "ironer":
		return "ironer_bag_rail", nil
	case "sorting":
		return "sorting_bag_rail", nil
	default:
		return "", fmt.Errorf("invalid rail type: %s", railType)
	}
}
