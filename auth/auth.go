package auth

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"stuInfoCapturer/constant"
	"time"
)

type QRStatus int

const (
	Available QRStatus = 0
	Succeed   QRStatus = 1
	Scanned   QRStatus = 2
	Timeout   QRStatus = 3
)

type Session struct {
	QrLoginForm map[string]string `json:"qr_login_form"`
	Cookies     map[string]string `json:"cookies"`
	UUID        string            `json:"uuid"`
}

func NewSession() (s *Session, err error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", constant.MainURL, nil)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, errors.New("bad login status code: " + strconv.Itoa(resp.StatusCode))
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(bodyText)))
	if err != nil {
		return nil, err
	}
	inputs := doc.Find("#qrLoginForm").Find("input")

	s = new(Session)
	s.QrLoginForm = make(map[string]string)
	inputs.Each(func(i int, input *goquery.Selection) {
		name, _ := input.Attr("name")
		value, _ := input.Attr("value")
		s.QrLoginForm[name] = value
	})
	s.Cookies = make(map[string]string)
	for _, cookie := range resp.Cookies() {
		s.Cookies[cookie.Name] = cookie.Value
	}

	token, err := s.getToken()
	if err != nil {
		return nil, err
	}
	s.UUID = token

	return s, nil
}

func (session *Session) getToken() (uuid string, err error) {
	client := &http.Client{}
	ts := strconv.FormatInt(time.Now().UnixMilli(), 10)
	url := constant.GetTokenURL + ts

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	cookieStr := fmt.Sprintf("JSESSIONID=%s; route=%s", session.Cookies["JSESSIONID"], session.Cookies["route"])
	req.Header.Add("Cookie", cookieStr)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("bad getToken status code: %d", resp.StatusCode)
	}

	uuid = string(bodyText)
	session.QrLoginForm["uuid"] = uuid
	return uuid, nil
}

func (session *Session) GetQRCode() (qrCode []byte, err error) {
	client := &http.Client{}
	url := constant.GetQRCodeURL + session.UUID

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	cookieStr := fmt.Sprintf("JSESSIONID=%s; route=%s", session.Cookies["JSESSIONID"], session.Cookies["route"])
	req.Header.Add("Cookie", cookieStr)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("bad getQRCode status code: %d", resp.StatusCode)
	}
	return body, nil
}

func (session *Session) CheckQRStatus() (status QRStatus, err error) {
	client := &http.Client{}
	url := fmt.Sprintf(constant.GetQRStatusURL, time.Now().UnixMilli(), session.UUID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("error creating request: %w", err)
	}

	cookieStr := fmt.Sprintf("JSESSIONID=%s; route=%s", session.Cookies["JSESSIONID"], session.Cookies["route"])
	req.Header.Add("Cookie", cookieStr)

	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	if resp.StatusCode != 200 {
		return 0, errors.New("bad getQRStatus status code: " + strconv.Itoa(resp.StatusCode))
	}
	statusInt, err := strconv.Atoi(string(bodyText))
	if err != nil {
		return 0, err
	}
	return QRStatus(statusInt), nil
}

func (session *Session) Login() (xjwglCookie map[string]string, err error) {
	xjwglCookie = make(map[string]string)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	data := url.Values{}
	for k, v := range session.QrLoginForm {
		data.Set(k, v)
	}

	req, err := http.NewRequest("POST", constant.LoginPOSTURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("login request creating error: %w", err)
	}

	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/114.0")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	cookieStr := fmt.Sprintf("route=%s; JSESSIONID=%s; Secure; org.springframework.web.servlet.i18n.CookieLocaleResolver.LOCALE=zh_CN", session.Cookies["route"], session.Cookies["JSESSIONID"])
	req.Header.Add("Cookie", cookieStr)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("login request sending error: %w", err)
	}
	defer resp.Body.Close()

	headers := resp.Header
	redirectLoc := headers.Get("Location")
	redirectLoc = strings.Replace(redirectLoc, "http://", "https://", 1)

	req, err = http.NewRequest("GET", redirectLoc, nil)
	if err != nil {
		return nil, fmt.Errorf("login redirect 1 request creating error: %w", err)
	}

	resp, err = client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("login redirect 1 request sending error: %w", err)
	}
	defer resp.Body.Close()

	headers = resp.Header
	redirectLoc = headers.Get("Location")
	redirectLoc = strings.Replace(redirectLoc, "http://", "https://", 1)

	cookies := make(map[string]string)
	for _, cookie := range resp.Cookies() {
		if cookie.Name != "JSESSIONID" && cookie.Name != "route" {
			continue
		}
		cookies[cookie.Name] = cookie.Value
	}

	xjwglCookie["route"] = cookies["route"]

	req, err = http.NewRequest("GET", redirectLoc, nil)
	if err != nil {
		return nil, fmt.Errorf("login redirect 2 request creating error: %w", err)
	}

	req.Header.Add("Cookie", fmt.Sprintf("JSESSIONID=%s; route=%s", cookies["JSESSIONID"], cookies["route"]))

	// TODO: 等待0.5秒
	time.Sleep(500 * time.Millisecond)

	resp, err = client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("login redirect 2 request sending error: %w", err)
	}
	defer resp.Body.Close()

	headers = resp.Header
	redirectLoc = headers.Get("Location")
	redirectLoc = strings.Replace(redirectLoc, "http://", "https://", 1)

	req, err = http.NewRequest("GET", redirectLoc, nil)
	if err != nil {
		return nil, fmt.Errorf("login redirect 3 request creating error: %w", err)
	}

	req.Header.Add("Cookie", fmt.Sprintf("route=%s", cookies["route"]))

	// TODO: 等待0.5秒
	time.Sleep(500 * time.Millisecond)

	resp, err = client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("login redirect 3 request sending error: %w", err)
	}
	defer resp.Body.Close()

	for _, cookie := range resp.Cookies() {
		if cookie.Name != "JSESSIONID" {
			continue
		}
		xjwglCookie["JSESSIONID"] = cookie.Value
	}

	return xjwglCookie, nil
}
