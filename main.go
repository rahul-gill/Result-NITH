package main

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"result_scrape/db"
	"result_scrape/domain"
	"strconv"
	"strings"
)

func isRollNumberValid(rollNumber string) bool {
	return true
}

func getUrlForRollNumber(rollNumber string) string {
	scheme := rollNumber[:2]
	return fmt.Sprintf("http://14.139.56.19/scheme%s/studentresult/result.asp", scheme)
}

func getResultHtml(rollNumber string) (io.ReadCloser, error) {
	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	httpClient := &http.Client{
		Jar: cookieJar,
	}
	path := getUrlForRollNumber(rollNumber)

	//get tokens
	formPageResponse, err := httpClient.Get(path)
	if err != nil {
		return nil, err
	}
	formPageDoc, err := goquery.NewDocumentFromReader(formPageResponse.Body)
	if err != nil {
		return nil, err
	}
	csrfToken, exists := formPageDoc.Find("[name=CSRFToken]").Attr("value")
	if !exists {
		return nil, fmt.Errorf("CSRFToken not found")
	}
	verToken, exists := formPageDoc.Find("[name=RequestVerificationToken]").Attr("value")
	if !exists {
		return nil, fmt.Errorf("RequestVerificationToken not found")
	}
	//get result html
	data := url.Values{
		"RollNumber":               {rollNumber},
		"CSRFToken":                {csrfToken},
		"RequestVerificationToken": {verToken},
		"B1":                       {"Submit"},
	}
	postReq, err := http.NewRequest(http.MethodPost, path, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	postReq.Header.Set("DNT", "1")
	postReq.Header.Set("Content-Type", " application/x-www-form-urlencoded")
	postReq.AddCookie(formPageResponse.Cookies()[0])
	resp, err := httpClient.Do(postReq)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

//go:embed db/sql/schema.sql
var ddl string

func domainStudentToCreateStudentParams(student domain.Student) db.CreateStudentParams {
	cgpi, _ := strconv.ParseFloat(student.CGPI, 64)
	return db.CreateStudentParams{
		RollNumber:     student.RollNumber,
		Name:           sql.NullString{String: student.Name, Valid: true},
		FathersName:    sql.NullString{String: student.FathersName, Valid: true},
		Batch:          sql.NullString{String: "", Valid: true},
		Branch:         sql.NullString{String: "", Valid: true},
		LatestSemester: sql.NullInt64{Int64: int64(len(student.SemesterResults)), Valid: true},
		Cgpi:           sql.NullFloat64{Float64: cgpi, Valid: true},
	}
}

func getDbQueries() *db.Queries {
	os.RemoveAll("result_db_tmp.db")
	os.Create("result_db_tmp.db")

	database, _ := sql.Open("sqlite3", "./result_db_tmp.db")
	ctx := context.Background()

	// create tables
	database.ExecContext(ctx, ddl)
	queries := db.New(database)
	return queries
}

/**
 * For testing
 */
func parseResultHtml(body io.Reader) {
	resultDoc, _ := goquery.NewDocumentFromReader(body)

	tableFind := resultDoc.Find("table")
	tableFind.Each(func(i int, selection *goquery.Selection) {
		if i == tableFind.Length()-1 {
			return
		}
		println("TABLE no", i, "------------------------------------")
		selection.Find("tr").Each(func(j int, selection *goquery.Selection) {
			print("Row number ", j, ": ")
			selection.Find("td").Each(func(k int, selection *goquery.Selection) {
				print(strings.TrimSpace(selection.Text()), "|")

			})
			print("\n")
		})
		println("------------------------------------\n")
	})
}

func getResultsFromWeb() []domain.Student {
	//build an array of roll numbers
	var rollNumbers []string
	for i := 0; i < 150; i++ {
		rollNumbers = append(
			rollNumbers,
			fmt.Sprintf("191%.3d", i),
			fmt.Sprintf("192%.3d", i),
			fmt.Sprintf("193%.3d", i),
			fmt.Sprintf("194%.3d", i),
			fmt.Sprintf("195%.3d", i),
			fmt.Sprintf("196%.3d", i),
			fmt.Sprintf("197%.3d", i),
			fmt.Sprintf("198%.3d", i),
		)
		if i < 100 {
			rollNumbers = append(
				rollNumbers,
				fmt.Sprintf("1955%.2d", i),
				fmt.Sprintf("1945%.2d", i),
			)
		}
	}
	//build an array of student objects that contain result
	var students []domain.Student
	ch := make(chan *domain.Student)
	for _, rollNumber := range rollNumbers {
		go func(ch chan *domain.Student, rollNumber string) {
			resultHtml, err := getResultHtml(rollNumber)
			if err != nil {
				err = fmt.Errorf("error for rollNumber %s: %w", rollNumber, err)
				log.Print(err)
				return
			}
			student, _, err := domain.ParseResultHtml(resultHtml)
			if err == nil {
				ch <- student
			} else {
				ch <- nil
			}
		}(ch, rollNumber)
	}
	for range rollNumbers {
		newStudent := <-ch
		if newStudent != nil {
			students = append(students, *newStudent)
		}
	}
	return students
}

func main() {
	students := getResultsFromWeb()
	xx, _ := json.Marshal(students)
	os.WriteFile("outputData.json", xx, 0644)
}
