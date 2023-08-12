package score

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"stuInfoCapturer/constant"
)

type ZCScore struct {
	Error       bool   `json:"error"`
	StuNum      string `json:"stuNum"`
	BaseScore   string `json:"baseScore"`
	ExtraScore  string `json:"extraScore"`
	LowestScore int32  `json:"lowestScore"`
	AvgScore    string `json:"avgScore"`
}

func Calculate(json Score) ZCScore {
	zcScore := ZCScore{
		Error:  false,
		StuNum: json.Items[0].Xh,
	}
	base, lowest, avg, err := calculateScore(json)
	if err != nil {
		log.Println("error when calculate score: ", err)
		zcScore.Error = true
		return zcScore
	}
	zcScore.BaseScore = base
	zcScore.ExtraScore = calculateExtraScore(json)
	zcScore.LowestScore = lowest
	zcScore.AvgScore = avg

	return zcScore
}

func calculateScore(json Score) (baseScore string, lowestScore int32, avgScore string, err error) {
	items := json.Items
	var (
		// 各科成绩x学分(分子)，不包括第1学期补考的科目，包括第1学期挂科的科目(在第2学期补考合格的计60分，否则计0分)
		weightedGrade float64
		// 总学分(分母)，不包括第1学期补考的科目(无论是否及格)，包括第1学期挂科的科目(在第2学期补考合格的计60分，否则计0分)
		sumCredit float64
		// 额外扣分(在基本分中扣除)，必修和选修不及格的，扣该科目学分数
		extraDeduction float64
		// 0分科目，不确定是否为缓考的
		zeroItems = make([]Item, 0)

		// 最低分
		lowest int32 = 100
		// 计算算数平均分用的
		sum   int64
		count int32
		// 不及格科目，如果补考通过，计60分，否则计挂科中的最低分
		failedItems = make([]Item, 0)
	)

	for _, item := range items {
		if !strings.Contains(item.Xnmmc, constant.Xn) {
			// 跳过非指定学年的科目
			continue
		}

		if strings.Contains(item.Cj, "缓考") {
			// 跳过缓考的科目
			continue
		}

		score, err := strconv.ParseInt(strings.TrimSpace(item.Cj), 10, 32)
		if err != nil {
			log.Println("error when parse score: ", strings.TrimSpace(item.Cj))
			return "", -1, "", err
		}
		credit, err := strconv.ParseFloat(strings.TrimSpace(item.Xf), 64)
		if err != nil {
			log.Println("error when parse credit: ", strings.TrimSpace(item.Xf))
			return "", -1, "", err
		}
		semester, err := strconv.ParseInt(strings.TrimSpace(item.Xqmmc), 10, 32)
		if err != nil {
			log.Println("error when parse semester: ", strings.TrimSpace(item.Xqmmc))
			return "", -1, "", err
		}

		if strings.Contains(item.Ksxz, "重修") {
			// 重修科目
			lowest = min(lowest, int32(score))
			sum += score
			count++
			if score >= 60 {
				continue
			} else {
				extraDeduction += credit
				continue
			}
		}

		if score == 0 && strings.Contains(item.Ksxz, "正常考试") {
			// 0分科目，不确定是否为缓考，暂时跳过
			zeroItems = append(zeroItems, item)
			continue
		}

		if strings.Contains(item.Ksxz, "缓考") {
			// 缓考科目，检查是否在0分科目列表中
			for i := 0; i < len(zeroItems); i++ {
				if zeroItems[i].Kch == item.Kch {
					// 在0分科目列表中，从列表中删除
					zeroItems = append(zeroItems[:i], zeroItems[i+1:]...)
					break
				}
			}
		}

		if strings.Contains(item.Ksxz, "正常考试") || strings.Contains(item.Ksxz, "缓考") {
			if score >= 60 {
				// 正常考试或缓考及格
				lowest = min(lowest, int32(score))
				sum += score
				count++
				weightedGrade += float64(score) * credit
				sumCredit += credit
				continue
			} else {
				// 正常考试或缓考不及格
				failedItems = append(failedItems, item)
				count++
				sumCredit += credit
				extraDeduction += credit
				continue
			}
		}

		if strings.Contains(item.Ksxz, "补考") {
			if semester == 1 {
				// 第1学期补考
				if score >= 60 {
					continue
				} else {
					extraDeduction += credit
					continue
				}
			} else if semester == 2 {
				// 第2学期补考
				for i := 0; i < len(failedItems); i++ {
					if failedItems[i].Kch == item.Kch {
						// 在不及格科目列表中，从列表中删除
						if score >= 60 {
							// 补考及格，单科最低分计60，算平均分时也计60
							lowest = min(lowest, 60)
							sum += 60
						} else {
							// 补考不及格，单科最低分计挂科中的最低分，算平均分时也计挂科中的最低分
							iS, _ := strconv.ParseInt(strings.TrimSpace(failedItems[i].Cj), 10, 32)
							l := min(int32(score), int32(iS))
							lowest = min(lowest, l)
							sum += int64(l)
						}

						failedItems = append(failedItems[:i], failedItems[i+1:]...)
						break
					}
				}

				if score >= 60 {
					weightedGrade += 60 * credit
					continue
				} else {
					continue
				}
			}
		}

		return "", -1, "", fmt.Errorf("unknown error")
	}

	for _, item := range zeroItems {
		// 真实的0分(非缓考)
		lowest = 0
		count++
		credit, _ := strconv.ParseFloat(item.Xf, 64)
		sumCredit += credit
		extraDeduction += credit
	}

	for _, item := range failedItems {
		// 没有补考的挂科
		score, _ := strconv.ParseInt(strings.TrimSpace(item.Cj), 10, 32)
		lowest = min(lowest, int32(score))
		sum += score
	}

	base := ((weightedGrade / sumCredit) * 0.8) - extraDeduction
	avg := float64(sum) / float64(count)
	return strconv.FormatFloat(base, 'f', 2, 64), lowest, strconv.FormatFloat(avg, 'f', 2, 64), nil
}

func calculateExtraScore(json Score) string {
	items := json.Items
	var (
		extraScore float64
		over90     int
		over85     int
	)
	for _, item := range items {
		if !strings.Contains(item.Xnmmc, constant.Xn) {
			// 跳过非指定学年的科目
			continue
		}

		if strings.Contains(item.Ksxz, "补考") {
			continue
		}

		score, err := strconv.ParseInt(item.Cj, 10, 32)
		if err != nil {
			continue
		}

		if score >= 90 {
			over90++
			continue
		}
		if score >= 85 {
			over85++
			continue
		}
	}

	extraScore = float64(over90) + float64(over85)*0.5
	scoreStr := strconv.FormatFloat(extraScore, 'f', 2, 64)
	return fmt.Sprintf("%s (90分以上科目数:%d，85分以上科目数:%d)", scoreStr, over90, over85)
}

func min(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}
