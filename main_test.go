package main

import (
	"github.com/tealeg/xlsx"
	"os"
	"path"
	"stuInfoCapturer/score"
	"testing"
)

func Test(t *testing.T) {
	files, _ := os.ReadDir("myScore")
	for _, file := range files {
		s, err := filenameToScore(path.Join("myScore", file.Name()))
		if err != nil {
			t.Fatal(err)
		}
		t.Log(file.Name()+":", s)
	}
}

func filenameToScore(filename string) (score.ZCScore, error) {
	s := score.Score{
		Items: make([]score.Item, 0),
	}

	file, err := xlsx.OpenFile(filename)
	if err != nil {
		return score.ZCScore{}, err
	}
	sheet := file.Sheets[0]
	for i := 1; i < len(sheet.Rows); i++ {
		row := sheet.Rows[i]
		item := score.Item{
			Xnmmc:  row.Cells[0].Value,
			Xqmmc:  row.Cells[1].Value,
			Kch:    row.Cells[2].Value,
			Kcmc:   row.Cells[3].Value,
			Kcxzmc: row.Cells[4].Value,
			Xf:     row.Cells[5].Value,
			Cj:     row.Cells[6].Value,
			Jd:     row.Cells[8].Value,
			Ksxz:   row.Cells[9].Value,
			Sfxwkc: row.Cells[10].Value,
			Kkbmmc: row.Cells[11].Value,
			Kclbmc: row.Cells[13].Value,
			Kcgsmc: row.Cells[14].Value,
			Jxbmc:  row.Cells[15].Value,
			Jsxm:   row.Cells[16].Value,
			Khfsmc: row.Cells[17].Value,
			Xh:     row.Cells[18].Value,
			Xm:     row.Cells[19].Value,
			Cjsfzf: row.Cells[21].Value,
			Xfjd:   row.Cells[22].Value,
		}
		s.Items = append(s.Items, item)
	}

	zc := score.Calculate(s)

	return zc, nil
}
