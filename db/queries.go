package db

import (
	"context"
	"fmt"
)

type SortType int8

type StudentBranch int8

type StudentBatch int8

const (
	SortByName SortType = iota
	SortByroll_number
	SortByCGPI
)

const (
	NoneBranchFilter StudentBranch = iota
)

const (
	NoneBatchFilter StudentBatch = iota
)

func (branch StudentBranch) GetBranchName() string {
	return ""
}

func (branch StudentBatch) GetBatchName() string {
	return ""
}

func (q *Queries) GetStudents(
	ctx context.Context,
	searchString string, //roll Number or name
	isSortOrderAscending bool,
	sortingType SortType,
	pageSize int,
	pageIndex int, //starting index is 0
	branch StudentBranch,
	batch StudentBatch,
	minCG float32,
	maxCG float32,
) ([]Student, error) {
	var sortField string
	var orderType string
	offset := pageIndex * pageSize
	switch sortingType {
	case SortByName:
		sortField = "name"
	case SortByCGPI:
		sortField = "cgpi"
	default:
		sortField = "roll_number"
	}
	switch isSortOrderAscending {
	case false:
		orderType = "DESC"
	default:
		orderType = "ASC"

	}
	query := "SELECT * FROM student WHERE 1 = 1 "
	if len(searchString) != 0 {
		query += "AND name like " + searchString + " or roll_number like " + searchString + " "
	}
	if branch != NoneBranchFilter {
		query += "AND branch = " + branch.GetBranchName() + " "
	}
	if batch != NoneBatchFilter {
		query += "AND branch = " + branch.GetBranchName() + " "
	}
	query += "AND cgpi >= " + fmt.Sprintf("%.2f", minCG) + " "
	query += "AND cgpi <= " + fmt.Sprintf("%.2f", maxCG) + " "
	query += "ORDER BY " + sortField + " " + orderType + " "
	query += " LIMIT " + fmt.Sprintf("%d", pageSize) + " OFFSET " + fmt.Sprintf("%d", offset)
	println(query)

	//boilerplate(from sqlc auto-generate)
	rows, err := q.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Student
	for rows.Next() {
		var i Student
		if err := rows.Scan(
			&i.RollNumber,
			&i.Name,
			&i.FathersName,
			&i.Batch,
			&i.Branch,
			&i.LatestSemester,
			&i.Cgpi,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
