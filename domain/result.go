package domain

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"strconv"
	"strings"
)

/**
- table 0 => last update title table
- table 1 => row 0 => rollNo, name, fathersName
- tables from 2 to end, not including last one have semester data
	- two table for each semester
	- first table subjects
	- second table summary
- last table useless
*/

type Student struct {
	RollNumber      string           `json:"roll_number"`
	Name            string           `json:"name"`
	FathersName     string           `json:"fathers_name"`
	SemesterResults []SemesterResult `json:"semester_results"`
	CGPI            string           `json:"cgpi"`
}

type SemesterResult struct {
	SemesterNumber int             `json:"semester_number"`
	SubjectResults []SubjectResult `json:"subject_results"`
	SGPI           string          `json:"sgpi"`
	CGPI           string          `json:"cgpi"`
}

type SubjectResult struct {
	SubjectName string `json:"subject_name"`
	SubjectCode string `json:"subject_code"`
	SubPoint    int    `json:"sub_point"`
	Grade       string `json:"grade"`
	SubGP       int    `json:"sub_gp"`
}

func ParseResultHtml(body io.ReadCloser) (user *Student, lastUpdateResultName string, err error) {
	resultDoc, _ := goquery.NewDocumentFromReader(body)
	user = &Student{}
	tableFind := resultDoc.Find("table")
	semesters := (tableFind.Length() - 3) / 2
	if semesters < 0 || semesters >= tableFind.Length() {
		return nil, "", fmt.Errorf("something went wrong")
	}
	user.SemesterResults = make([]SemesterResult, semesters)
	tableFind.Each(func(tableIndex int, selection *goquery.Selection) {
		if tableIndex == 0 {
			//last update title table
			lastUpdateResultName = strings.TrimSpace(selection.Find("tr td").Text())
		} else if tableIndex == 1 {
			//student roll number, name, father's name
			selection.Find("td").Each(func(cellIndex int, selection *goquery.Selection) {
				txt := strings.Replace(selection.Text(), "ROLL NUMBER", "", -1)
				txt = strings.Replace(txt, "STUDENT NAME", "", -1)
				txt = strings.Replace(txt, "FATHER NAME", "", -1)
				txt = strings.TrimSpace(txt)
				switch cellIndex {
				case 0:
					user.RollNumber = txt
				case 1:
					user.Name = txt
				case 2:
					user.FathersName = txt
				}
			})
		} else if tableIndex == tableFind.Length()-1 {
			//useless table
			return
		} else if tableIndex%2 == 0 {
			//semester result table: subjects data
			rowFind := selection.Find("tr")
			subjectsResult := make([]SubjectResult, rowFind.Length()-2)
			rowFind.Each(func(rowIndex int, selection *goquery.Selection) {
				if rowIndex < 2 {
					return
				}
				//each row is a subject after row index 1
				selection.Find("td").Each(func(cellIndex int, selection *goquery.Selection) {
					text := strings.TrimSpace(selection.Text())
					switch cellIndex {
					case 1:
						subjectsResult[rowIndex-2].SubjectName = text
					case 2:
						subjectsResult[rowIndex-2].SubjectCode = text
					case 3:
						subjectsResult[rowIndex-2].SubPoint, _ = strconv.Atoi(text)
					case 4:
						subjectsResult[rowIndex-2].Grade = text
					case 5:
						subjectsResult[rowIndex-2].SubGP, _ = strconv.Atoi(text)
					}
				})
			})
			user.SemesterResults[(tableIndex-2)/2].SubjectResults = subjectsResult
			user.SemesterResults[(tableIndex-2)/2].SemesterNumber = (tableIndex-2)/2 + 1
		} else {
			//semester result table: semester overall data
			selection.Find("tr td").Each(func(cellIndex int, selection *goquery.Selection) {
				equalCharPosition := strings.Index(selection.Text(), "=")
				text := strings.TrimSpace(selection.Text()[equalCharPosition+1:])
				if cellIndex == 1 {
					user.SemesterResults[(tableIndex-2)/2].SGPI = text
				} else if cellIndex == 3 {
					user.SemesterResults[(tableIndex-2)/2].CGPI = text
				}
			})
		}
	})
	user.CGPI = user.SemesterResults[len(user.SemesterResults)-1].CGPI
	return
}
