package score

import (
	"encoding/json"
	"fmt"
	"github.com/tealeg/xlsx"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"stuInfoCapturer/constant"
)

type Score struct {
	CurrentPage   int           `json:"currentPage"`
	CurrentResult int           `json:"currentResult"`
	EntityOrField bool          `json:"entityOrField"`
	Items         []Item        `json:"items"`
	Limit         int           `json:"limit"`
	Offset        int           `json:"offset"`
	PageNo        int           `json:"pageNo"`
	PageSize      int           `json:"pageSize"`
	ShowCount     int           `json:"showCount"`
	SortName      string        `json:"sortName"`
	SortOrder     string        `json:"sortOrder"`
	Sorts         []interface{} `json:"sorts"`
	TotalCount    int           `json:"totalCount"`
	TotalPage     int           `json:"totalPage"`
	TotalResult   int           `json:"totalResult"`
}

type Item struct {
	Bfzcj              string `json:"bfzcj"`
	Bh                 string `json:"bh"`
	BhID               string `json:"bh_id"`
	Bj                 string `json:"bj"`
	Cj                 string `json:"cj"`
	Cjsfzf             string `json:"cjsfzf"`
	Date               string `json:"date"`
	DateDigit          string `json:"dateDigit"`
	DateDigitSeparator string `json:"dateDigitSeparator"`
	Day                string `json:"day"`
	Jd                 string `json:"jd"`
	JgID               string `json:"jg_id"`
	Jgmc               string `json:"jgmc"`
	Jgpxzd             string `json:"jgpxzd"`
	Jsxm               string `json:"jsxm"`
	JxbID              string `json:"jxb_id"`
	Jxbmc              string `json:"jxbmc"`
	Kch                string `json:"kch"`
	KchID              string `json:"kch_id"`
	Kclbmc             string `json:"kclbmc"`
	Kcmc               string `json:"kcmc"`
	Kcxzdm             string `json:"kcxzdm"`
	Kcxzmc             string `json:"kcxzmc"`
	Key                string `json:"key"`
	Khfsmc             string `json:"khfsmc,omitempty"`
	Kkbmmc             string `json:"kkbmmc"`
	Kklxdm             string `json:"kklxdm"`
	Ksxz               string `json:"ksxz"`
	Ksxzdm             string `json:"ksxzdm"`
	Listnav            string `json:"listnav"`
	LocaleKey          string `json:"localeKey"`
	Month              string `json:"month"`
	NjdmID             string `json:"njdm_id"`
	Njmc               string `json:"njmc"`
	PageTotal          int    `json:"pageTotal"`
	Pageable           bool   `json:"pageable"`
	QueryModel         struct {
		CurrentPage   int           `json:"currentPage"`
		CurrentResult int           `json:"currentResult"`
		EntityOrField bool          `json:"entityOrField"`
		Limit         int           `json:"limit"`
		Offset        int           `json:"offset"`
		PageNo        int           `json:"pageNo"`
		PageSize      int           `json:"pageSize"`
		ShowCount     int           `json:"showCount"`
		Sorts         []interface{} `json:"sorts"`
		TotalCount    int           `json:"totalCount"`
		TotalPage     int           `json:"totalPage"`
		TotalResult   int           `json:"totalResult"`
	} `json:"queryModel"`
	Rangeable   bool   `json:"rangeable"`
	RowID       string `json:"row_id"`
	Rwzxs       string `json:"rwzxs,omitempty"`
	Sfdkbcx     string `json:"sfdkbcx"`
	Sfxwkc      string `json:"sfxwkc"`
	Sfzh        string `json:"sfzh"`
	Tjrxm       string `json:"tjrxm,omitempty"`
	Tjsj        string `json:"tjsj"`
	TotalResult string `json:"totalResult"`
	UserModel   struct {
		Monitor    bool   `json:"monitor"`
		RoleCount  int    `json:"roleCount"`
		RoleKeys   string `json:"roleKeys"`
		RoleValues string `json:"roleValues"`
		Status     int    `json:"status"`
		Usable     bool   `json:"usable"`
	} `json:"userModel"`
	Xb     string `json:"xb"`
	Xbm    string `json:"xbm"`
	Xf     string `json:"xf"`
	Xfjd   string `json:"xfjd"`
	Xh     string `json:"xh"`
	XhID   string `json:"xh_id"`
	Xm     string `json:"xm"`
	Xnm    string `json:"xnm"`
	Xnmmc  string `json:"xnmmc"`
	Xqm    string `json:"xqm"`
	Xqmmc  string `json:"xqmmc"`
	Xslb   string `json:"xslb"`
	Year   string `json:"year"`
	Zsxymc string `json:"zsxymc"`
	ZyhID  string `json:"zyh_id"`
	Zymc   string `json:"zymc"`
	Kcgsmc string `json:"kcgsmc,omitempty"`
}

func GenerateScoreFile(cookie map[string]string) (ZCScore, error) {
	score, err := getScoreJson(cookie)
	if err != nil {
		return ZCScore{}, err
	}
	go func() {
		// 保存为xlsx文件
		toExcel(score)
	}()

	zcScore := Calculate(score)
	go func() {
		// 保存综测分为json文件
		bytes, _ := json.Marshal(&zcScore)
		os.WriteFile(filepath.Join(constant.ZCScoreDir, zcScore.StuNum+".txt"), bytes, 0644)
	}()

	return zcScore, nil
}

func toExcel(score Score) (filename string) {
	xlsx.SetDefaultFont(10, "Arial")
	file := xlsx.NewFile()
	sheet, err := file.AddSheet("Sheet1")
	if err != nil {
		log.Fatal(err)
		return ""
	}
	firstRow := sheet.AddRow()
	firstRow.AddCell().Value = "学年"
	firstRow.AddCell().Value = "学期"
	firstRow.AddCell().Value = "课程代码"
	firstRow.AddCell().Value = "课程名称"
	firstRow.AddCell().Value = "课程性质"
	firstRow.AddCell().Value = "学分"
	firstRow.AddCell().Value = "成绩"
	firstRow.AddCell().Value = "成绩备注"
	firstRow.AddCell().Value = "绩点"
	firstRow.AddCell().Value = "成绩性质"
	firstRow.AddCell().Value = "是否学位课程"
	firstRow.AddCell().Value = "开课学院"
	firstRow.AddCell().Value = "课程标记"
	firstRow.AddCell().Value = "课程类别"
	firstRow.AddCell().Value = "课程归属"
	firstRow.AddCell().Value = "教学班"
	firstRow.AddCell().Value = "任课教师"
	firstRow.AddCell().Value = "考核方式"
	firstRow.AddCell().Value = "学号"
	firstRow.AddCell().Value = "姓名"
	firstRow.AddCell().Value = "学生标记"
	firstRow.AddCell().Value = "是否成绩作废"
	firstRow.AddCell().Value = "学分绩点"

	for _, item := range score.Items {
		row := sheet.AddRow()
		row.AddCell().Value = item.Xnmmc
		row.AddCell().Value = item.Xqmmc
		row.AddCell().Value = item.Kch
		row.AddCell().Value = item.Kcmc
		row.AddCell().Value = item.Kcxzmc
		row.AddCell().Value = item.Xf
		row.AddCell().Value = item.Cj
		row.AddCell().Value = ""
		row.AddCell().Value = item.Jd
		row.AddCell().Value = item.Ksxz
		row.AddCell().Value = item.Sfxwkc
		row.AddCell().Value = item.Kkbmmc
		row.AddCell().Value = ""
		row.AddCell().Value = item.Kclbmc
		row.AddCell().Value = item.Kcgsmc
		row.AddCell().Value = item.Jxbmc
		row.AddCell().Value = item.Jsxm
		row.AddCell().Value = item.Khfsmc
		row.AddCell().Value = item.Xh
		row.AddCell().Value = item.Xm
		row.AddCell().Value = ""
		row.AddCell().Value = item.Cjsfzf
		row.AddCell().Value = item.Xfjd
	}

	// 设置第一行单元格背景色为#008000，字体为白色加粗，居中
	style := xlsx.NewStyle()
	style.Fill = *xlsx.NewFill("solid", "008000", "008000")
	style.Font = *xlsx.NewFont(10, "Arial")
	style.Font.Color = "FFFFFF"
	style.Font.Bold = true
	style.Alignment = xlsx.Alignment{Horizontal: "center", Vertical: "center"}
	for _, cell := range firstRow.Cells {
		cell.SetStyle(style)
	}

	stuID := score.Items[0].Xh
	stuName := score.Items[0].Xm
	filename = stuID + " " + stuName + ".xlsx"
	err = file.Save(path.Join(constant.ScoreDir, filename))
	if err != nil {
		log.Fatal(err)
		return ""
	}
	return filename
}

func getScoreJson(cookie map[string]string) (Score, error) {
	client := &http.Client{}
	var data = strings.NewReader(constant.ScoreParam)
	req, err := http.NewRequest("POST", constant.ScoreURL, data)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/114.0")
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.8,zh-TW;q=0.7,zh-HK;q=0.5,en-US;q=0.3,en;q=0.2")
	// req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=utf-8")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	//req.Header.Set("Origin", "https://****.****.edu.cn")
	req.Header.Set("Connection", "keep-alive")
	//req.Header.Set("Referer", "https://****.****.edu.cn/cjcx/cjcx_cxDgXscj.html?gnmkdm=N305005&layout=default&su=")
	cookieStr := fmt.Sprintf("route=%s; JSESSIONID=%s", cookie["route"], cookie["JSESSIONID"])
	req.Header.Set("Cookie", cookieStr)
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")

	resp, err := client.Do(req)
	if err != nil {
		return Score{}, err
	}
	defer resp.Body.Close()
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		return Score{}, err
	}

	var score Score
	err = json.Unmarshal(bodyText, &score)
	if err != nil {
		return Score{}, err
	}
	return score, nil
}
