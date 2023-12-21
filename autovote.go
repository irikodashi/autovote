package main

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

const (
	MAIN_PAGE               string = "https://XXX.XX/YYYYYYY/m/ZZZZZZ"
	INPUT_FIELD_SERIAL_CODE string = `//*[@id="e_7183"]`
	DROPDOWN_MEMBERS        string = `//*[@id="e_7184"]`
	DROPDOWN_TARGET_VALUE   string = `13`
	DROPDOWN_PREFECTURES    string = `//*[@id="e_7196"]`
	NEXT_PAGE               string = `//*[@id="__send"]`
	CONFIRM_TEXT            string = `/html/body/form/font[5]/font` // 確認画面に現れるテキスト
	BUTTON_SUBMIT           string = `//*[@id="__commit"]`
	SCREENSHOT_QUALITY      int    = 60 // スクショのクオリティ
	PAGE_BODY               string = `/html/body`
)

// カレントパス
var currentPath string = path.Dir(os.Args[0])

var buf []byte

func main() {

	/*********************************
	いろいろ設定
	*********************************/
	// ヘッドレスモードで実行しない
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-gpu", false),
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("disable-extensions", false),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()
	ctx, cancel = context.WithTimeout(ctx, 24*time.Hour)
	defer cancel()

	/*********************************
	都道府県取得
	*********************************/
	prefecture := getPrefecture()

	/*********************************
	シリアルコードの個数分実行
	*********************************/
	serials := getSerials()
	for _, serial := range serials {
		if strings.TrimSpace(serial) != "" {

			// 初期画面
			init_page_tasks := chromedp.Tasks{
				// 投票ページ遷移
				chromedp.Navigate(MAIN_PAGE),
				chromedp.Sleep(500 * time.Millisecond),
			}
			err := chromedp.Run(ctx, init_page_tasks)

			// 投票画面
			var res string
			vote_tasks := chromedp.Tasks{
				chromedp.WaitVisible(INPUT_FIELD_SERIAL_CODE),
				chromedp.SendKeys(INPUT_FIELD_SERIAL_CODE, serial, chromedp.BySearch),
				chromedp.SetValue(DROPDOWN_MEMBERS, DROPDOWN_TARGET_VALUE, chromedp.BySearch),
				chromedp.SetValue(DROPDOWN_PREFECTURES, prefecture, chromedp.BySearch),
				chromedp.Click(NEXT_PAGE, chromedp.BySearch),
				chromedp.Text(PAGE_BODY, &res, chromedp.BySearch),
			}
			err = chromedp.Run(ctx, vote_tasks)
			if err != nil {
				log.Fatal(err)
			}

			switch errorCheck(res) {
			case 1:
				writeFile(serial+" 存在しないシリアルコード", "エラーシリアルコード.txt")
			case 2:
				writeFile(serial+" 使用済みシリアルコード", "エラーシリアルコード.txt")
			case 3:
				writeFile(serial+" なんらかのエラー", "エラーシリアルコード.txt")
			case 0:
				// 確認画面
				confirm_tasks := chromedp.Tasks{
					chromedp.WaitVisible(CONFIRM_TEXT, chromedp.BySearch),
					chromedp.FullScreenshot(&buf, SCREENSHOT_QUALITY),
					chromedp.Sleep(1 * time.Second),
					chromedp.Click(BUTTON_SUBMIT, chromedp.BySearch),
					chromedp.Sleep(1 * time.Second),
				}
				err = chromedp.Run(ctx, confirm_tasks)
				if err != nil {
					log.Fatal(err)
				}

				if err := ioutil.WriteFile(
					fmt.Sprintf(
						"%s%s%s%s", currentPath, "/screenshots/シリアルコード_", serial, ".jpg"), buf, 0o644,
				); err != nil {
					log.Fatal(err)
				}

				writeFile(serial, "使用済みシリアルコード.txt")
			}

		}
	}

	log.Printf("Finished!")

}

func errorCheck(s string) int {

	var res int
	if strings.Contains(s, "エラー：該当のシリアルコード") {
		res = 1
	} else if strings.Contains(s, "エラー: 同じシリアルコード") {
		res = 2
	} else if strings.Contains(s, "エラー: ") {
		res = 3
	} else {
		res = 0
	}

	return res
}

/*********************************
シリアルコード取得
*********************************/
func getSerials() []string {

	var serials []string
	fp := openFile("input/シリアルコード.txt")
	fs := readFileLineByLine(fp)
	for fs.Scan() {
		serials = append(serials, fs.Text())
	}

	return serials
}

/*********************************
都道府県の取得
*********************************/
func getPrefecture() string {

	prefecture := ""
	fp := openFile("input/都道府県.txt")
	fs := readFileLineByLine(fp)

	counter := 0
	for fs.Scan() {
		// 1行目が都道府県
		if counter == 0 {
			prefecture = fs.Text()
		}
		// 2行目以降は読み込まない
		counter += 1
		if counter > 0 {
			break
		}
	}
	if err := fs.Err(); err != nil {
		fmt.Println(err)
	} else if prefecture == "" {
		fmt.Println("都道府県.txtに都道府県を入力してください")
	}
	defer fp.Close()

	prefectures_cds := map[string]string{
		"北海道":  "1",
		"青森県":  "2",
		"岩手県":  "3",
		"宮城県":  "4",
		"秋田県":  "5",
		"山形県":  "6",
		"福島県":  "7",
		"茨城県":  "8",
		"栃木県":  "9",
		"群馬県":  "10",
		"埼玉県":  "11",
		"千葉県":  "12",
		"東京都":  "13",
		"神奈川県": "14",
		"新潟県":  "15",
		"富山県":  "16",
		"石川県":  "17",
		"福井県":  "18",
		"山梨県":  "19",
		"長野県":  "20",
		"岐阜県":  "21",
		"静岡県":  "22",
		"愛知県":  "23",
		"三重県":  "24",
		"滋賀県":  "25",
		"京都府":  "26",
		"大阪府":  "27",
		"兵庫県":  "28",
		"奈良県":  "29",
		"和歌山県": "30",
		"鳥取県":  "31",
		"島根県":  "32",
		"岡山県":  "33",
		"広島県":  "34",
		"山口県":  "35",
		"徳島県":  "36",
		"香川県":  "37",
		"愛媛県":  "38",
		"高知県":  "39",
		"福岡県":  "40",
		"佐賀県":  "41",
		"長崎県":  "42",
		"熊本県":  "43",
		"大分県":  "44",
		"宮崎県":  "45",
		"鹿児島県": "46",
		"沖縄県":  "47",
	}

	prefecture = prefectures_cds[prefecture]

	return prefecture
}

/*********************************
ファイルを開く
*********************************/
func openFile(fileName string) *os.File {

	// ファイルを開く
	f, err := os.Open(
		fmt.Sprintf("%s%s%s", currentPath, "/", fileName),
	)
	if err != nil {
		fmt.Println(err)
	}

	return f
}

/*********************************
書き込み(追記)
*********************************/
func writeFile(serial string, fileName string) {

	file, err := os.OpenFile(
		fmt.Sprintf("%s%s%s", currentPath, "/", fileName),
		os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666,
	)

	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	fmt.Fprintln(file, serial) //追記
}

/*********************************
1行ずつ取得する
*********************************/
func readFileLineByLine(f *os.File) *bufio.Scanner {

	scanner := bufio.NewScanner(f)

	return scanner
}
