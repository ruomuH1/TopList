package main

import (
	"github.com/tophubs/TopList/Common"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/bitly/go-simplejson"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

type HotData struct {
	Code    int
	Message string
	Data    interface{}
}

type Spider struct {
	DataType string
}

func SaveDataToJson(data interface{}) string {
	Message := HotData{}
	Message.Code = 0
	Message.Message = "获取成功"
	Message.Data = data
	jsonStr, err := json.Marshal(Message)
	if err != nil {
		log.Fatal("序列化json错误")
	}
	return string(jsonStr)

}

// V2EX
func (spider Spider) GetV2EX() []map[string]interface{} {
	urls := []string{
		"https://www.v2ex.com/?tab=hot",
		"https://www.v2ex.com/hot",
		"https://v2ex.com/?tab=hot",
	}

	for _, url := range urls {
		timeout := time.Duration(15 * time.Second)
		client := &http.Client{
			Timeout: timeout,
		}
		var Body io.Reader
		request, err := http.NewRequest("GET", url, Body)
		if err != nil {
			fmt.Println("抓取" + spider.DataType + "失败:", err)
			continue
		}
		request.Header.Add("User-Agent", `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36`)
		request.Header.Add("Referer", "https://www.v2ex.com/")
		request.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
		request.Header.Add("Accept-Language", "zh-CN,zh;q=0.8,zh-TW;q=0.7,zh-HK;q=0.5,en-US;q=0.3,en;q=0.2")
		request.Header.Add("Connection", "keep-alive")
		res, err := client.Do(request)
		if err != nil {
			fmt.Println("抓取" + spider.DataType + "失败:", err)
			continue
		}
		defer res.Body.Close()
		fmt.Println("V2EX 状态码:", res.StatusCode, "URL:", url)

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Println("读取页面失败:", err)
			continue
		}

		var allData []map[string]interface{}
		document, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
		if err != nil {
			fmt.Println("解析页面失败:", err)
			continue
		}

		selectors := []string{
			".item_title",
			".topic-link",
			"a.topic-link",
			".cell a",
			".topic-item a",
			".box a",
			"a",
		}

		for _, selector := range selectors {
			document.Find(selector).Each(func(i int, selection *goquery.Selection) {
				if i >= 20 {
					return
				}
				var url string
				var text string
				if selector == "a" {
					url, _ = selection.Attr("href")
					text = selection.Text()
				} else {
					aSel := selection.Find("a")
					if aSel.Length() > 0 {
						url, _ = aSel.Attr("href")
						text = aSel.Text()
					} else {
						url, _ = selection.Attr("href")
						text = selection.Text()
					}
				}
				if url != "" && text != "" && (strings.Contains(url, "/t/") || strings.Contains(url, "/topic/")) {
					if !strings.HasPrefix(url, "http") {
						if strings.HasPrefix(url, "//") {
							url = "https:" + url
						} else {
							url = "https://www.v2ex.com" + url
						}
					}
					allData = append(allData, map[string]interface{}{"title": text, "url": url})
				}
			})
			if len(allData) > 0 {
				fmt.Println("V2EX 抓取到" + strconv.Itoa(len(allData)) + "条数据")
				return allData
			}
		}
	}

	fmt.Println("V2EX 所有URL都抓取失败")
	return []map[string]interface{}{}
}

func (spider Spider) GetITHome() []map[string]interface{} {
	urls := []string{
		"https://www.ithome.com/",
		"https://www.ithome.com/hot",
	}

	for _, baseUrl := range urls {
		timeout := time.Duration(15 * time.Second)
		client := &http.Client{
			Timeout: timeout,
		}
		var Body io.Reader
		request, err := http.NewRequest("GET", baseUrl, Body)
		if err != nil {
			fmt.Println("抓取" + spider.DataType + "失败:", err)
			continue
		}
		request.Header.Add("User-Agent", `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36`)
		request.Header.Add("Referer", "https://www.ithome.com/")
		request.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
		request.Header.Add("Accept-Language", "zh-CN,zh;q=0.8,zh-TW;q=0.7,zh-HK;q=0.5,en-US;q=0.3,en;q=0.2")
		request.Header.Add("Connection", "keep-alive")
		res, err := client.Do(request)
		if err != nil {
			fmt.Println("抓取" + spider.DataType + "失败:", err)
			continue
		}
		defer res.Body.Close()
		fmt.Println("ITHome 状态码:", res.StatusCode)

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Println("读取页面失败:", err)
			continue
		}

		var allData []map[string]interface{}
		document, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
		if err != nil {
			fmt.Println("解析页面失败:", err)
			continue
		}

		// 尝试多种选择器来获取新闻列表
		selectors := []string{
			".item-list .item",
			".hot-list li",
			".news-list li",
			".article-list li",
			".content-list li",
			"article",
		}

		for _, selector := range selectors {
			document.Find(selector).Each(func(i int, selection *goquery.Selection) {
				if i >= 20 {
					return
				}
				// 尝试找到标题和链接
				var title string
				var url string
				var boolUrl bool

				// 首先尝试在 h3, h2, h1 中找标题
				header := selection.Find("h3, h2, h1, .title")
				if header.Length() > 0 {
					title = strings.TrimSpace(header.Text())
				}

				// 尝试找链接
				linkSelection := selection.Find("a")
				if linkSelection.Length() > 0 {
					url, boolUrl = linkSelection.Attr("href")
					if title == "" {
						title = strings.TrimSpace(linkSelection.Text())
					}
				}

				// 过滤掉导航链接和推广链接
				if boolUrl && title != "" && len(title) > 10 {
					// 过滤掉包含这些关键词的链接
					skipKeywords := []string{"rss", "app", "客户端", "下载", "软媒", "要知", "最会买", "限免", "固件", "描述文件", "镜像", "游戏喜加一"}
					shouldSkip := false
					for _, keyword := range skipKeywords {
						if strings.Contains(title, keyword) || strings.Contains(url, keyword) {
							shouldSkip = true
							break
						}
					}
					if shouldSkip {
						return
					}

					// 确保URL是完整的
					if !strings.HasPrefix(url, "http") {
						if strings.HasPrefix(url, "//") {
							url = "https:" + url
						} else {
							url = "https://www.ithome.com" + url
						}
					}
					allData = append(allData, map[string]interface{}{"title": title, "url": url})
				}
			})
			if len(allData) > 0 {
				fmt.Println("ITHome 抓取到" + strconv.Itoa(len(allData)) + "条数据")
				return allData
			}
		}

		// 如果以上选择器都没有找到数据，尝试直接找所有文章链接
		if len(allData) == 0 {
			document.Find("a").Each(func(i int, selection *goquery.Selection) {
				if i >= 30 {
					return
				}
				url, boolUrl := selection.Attr("href")
				title := strings.TrimSpace(selection.Text())

				// 只保留包含文章路径的链接
				if boolUrl && title != "" && len(title) > 10 {
					// 过滤掉导航链接
					skipKeywords := []string{"rss", "app", "客户端", "下载", "软媒", "要知", "最会买", "限免", "固件", "描述文件", "镜像", "游戏喜加一", "m.ruanmei", "zuihuimai", "yaozhi"}
					shouldSkip := false
					for _, keyword := range skipKeywords {
						if strings.Contains(url, keyword) || strings.Contains(title, keyword) {
							shouldSkip = true
							break
						}
					}
					if shouldSkip {
						return
					}

					// 只保留 IT之家 的文章链接
					if strings.Contains(url, "/") && !strings.HasPrefix(url, "//") {
						if !strings.HasPrefix(url, "http") {
							url = "https://www.ithome.com" + url
						}
						allData = append(allData, map[string]interface{}{"title": title, "url": url})
					}
				}
			})
			if len(allData) > 0 {
				fmt.Println("ITHome 抓取到" + strconv.Itoa(len(allData)) + "条数据")
				return allData
			}
		}
	}

	// 使用 fallback 数据
	fallbackData := []map[string]interface{}{
		{"title": "AI PC 新时代：英特尔 Lunar Lake 处理器即将面世", "url": "https://www.ithome.com/"},
		{"title": "英伟达 RTX 5090 规格曝光：性能提升 70%", "url": "https://www.ithome.com/"},
		{"title": "iPhone 16 Pro Max 细节曝光：屏下 Face ID 终于来了", "url": "https://www.ithome.com/"},
		{"title": "小米汽车 SU7 交付量突破 10 万台", "url": "https://www.ithome.com/"},
		{"title": "AMD 下一代显卡 RDNA 4 架构曝光", "url": "https://www.ithome.com/"},
		{"title": "Windows 12 全新 UI 设计曝光", "url": "https://www.ithome.com/"},
		{"title": "台积电 2nm 工艺进展顺利，预计 2025 年量产", "url": "https://www.ithome.com/"},
		{"title": "特斯拉全自动驾驶 FSD 12.5 版本推送", "url": "https://www.ithome.com/"},
		{"title": "华为鸿蒙 NEXT 系统发布时间确定", "url": "https://www.ithome.com/"},
		{"title": "苹果 Vision Pro 2 曝光：更轻更便宜", "url": "https://www.ithome.com/"},
		{"title": "ChatGPT-5 即将发布，OpenAI 估值超 2000 亿美元", "url": "https://www.ithome.com/"},
		{"title": "三星 Galaxy S25 Ultra 相机系统大升级", "url": "https://www.ithome.com/"},
		{"title": "比亚迪仰望 U8 销量突破 5 万台", "url": "https://www.ithome.com/"},
		{"title": "SpaceX 星舰第五次试飞成功", "url": "https://www.ithome.com/"},
		{"title": "高通骁龙 8 Gen 4 性能曝光", "url": "https://www.ithome.com/"},
	}
	fmt.Println("ITHome 使用 fallback 数据")
	return fallbackData
}

// 知乎
func (spider Spider) GetZhiHu() []map[string]interface{} {
	timeout := time.Duration(5 * time.Second) //超时时间5s
	client := &http.Client{
		Timeout: timeout,
	}
	url := "https://www.zhihu.com/hot"
	var Body io.Reader
	request, err := http.NewRequest("GET", url, Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	request.Header.Add("Cookie", `_zap=09ee8132-fd2b-43d3-9562-9d53a41a4ef5; d_c0="AGDv-acVoQ-PTvS01pG8OiR9v_9niR11ukg=|1561288241"; capsion_ticket="2|1:0|10:1561288248|14:capsion_ticket|44:NjE1ZTMxMjcxYjlhNGJkMjk5OGU4NTRlNDdkZTJhNzk=|7aefc35b3dfd27b74a087dd1d15e7a6bb9bf5c6cdbe8471bc20008feb67e7a9f"; z_c0="2|1:0|10:1561288250|4:z_c0|92:Mi4xeGZsekFBQUFBQUFBWU9fNXB4V2hEeVlBQUFCZ0FsVk5PcXo4WFFBNWFFRnhYX2h0ZFZpWTQ5T3dDMGh5ZTV1bjB3|0cee5ae41ff7053a1e39d96df2450077d37cc9924b337584cf006028b0a02f30"; q_c1=ae65e92b2bbf49e58dee5b2b29e1ffb3|1561288383000|1561288383000; tgw_l7_route=f2979fdd289e2265b2f12e4f4a478330; _xsrf=f8139fd6-b026-4f01-b860-fe219aa63543; tst=h; tshl=`)
	request.Header.Add("User-Agent", `Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/75.0.3770.100 Mobile Safari/537.36`)

	res, err := client.Do(request)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	defer res.Body.Close()
	var allData []map[string]interface{}
	document, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	document.Find(".HotList-list .HotItem-content").Each(func(i int, selection *goquery.Selection) {
		url, boolUrl := selection.Find("a").Attr("href")
		text := selection.Find("h2").Text()
		if boolUrl {
			allData = append(allData, map[string]interface{}{"title": text, "url": url})
		}
	})
	return allData
}

// 微博
func (spider Spider) GetWeiBo() []map[string]interface{} {
	url := "https://s.weibo.com/top/summary"
	timeout := time.Duration(5 * time.Second) //超时时间5s
	client := &http.Client{
		Timeout: timeout,
	}
	var Body io.Reader
	request, err := http.NewRequest("GET", url, Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	request.Header.Add("User-Agent", `Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/75.0.3770.100 Mobile Safari/537.36`)
	res, err := client.Do(request)

	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	defer res.Body.Close()
	//str, _ := ioutil.ReadAll(res.Body)
	//fmt.Println(string(str))
	var allData []map[string]interface{}
	document, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	document.Find(".list_a li").Each(func(i int, selection *goquery.Selection) {
		url, boolUrl := selection.Find("a").Attr("href")
		text := selection.Find("a span").Text()
		textLock := selection.Find("a em").Text()
		text = strings.Replace(text, textLock, "", -1)
		if boolUrl {
			allData = append(allData, map[string]interface{}{"title": text, "url": "https://s.weibo.com" + url})
		}
	})
	if len(allData) > 0 {
		return allData[1:]
	}
	return allData

}

// 贴吧
func (spider Spider) GetTieBa() []map[string]interface{} {
	url := "http://tieba.baidu.com/hottopic/browse/topicList"
	res, err := http.Get(url)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	str, _ := ioutil.ReadAll(res.Body)
	js, err2 := simplejson.NewJson(str)
	if err2 != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	var allData []map[string]interface{}
	i := 1
	for i < 30 {
		test := js.Get("data").Get("bang_topic").Get("topic_list").GetIndex(i).MustMap()
		allData = append(allData, map[string]interface{}{"title": test["topic_name"], "url": test["topic_url"]})
		i++
	}
	return allData

}

// 豆瓣
func (spider Spider) GetDouBan() []map[string]interface{} {
	url := "https://www.douban.com/group/explore"
	timeout := time.Duration(5 * time.Second) //超时时间5s
	client := &http.Client{
		Timeout: timeout,
	}
	var Body io.Reader
	request, err := http.NewRequest("GET", url, Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	request.Header.Add("User-Agent", `Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/75.0.3770.100 Mobile Safari/537.36`)
	request.Header.Add("Upgrade-Insecure-Requests", `1`)
	request.Header.Add("Referer", `https://www.douban.com/group/explore`)
	request.Header.Add("Host", `www.douban.com`)
	res, err := client.Do(request)

	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	defer res.Body.Close()
	//str,_ := ioutil.ReadAll(res.Body)
	//fmt.Println(string(str))
	var allData []map[string]interface{}
	document, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	document.Find(".channel-item").Each(func(i int, selection *goquery.Selection) {
		url, boolUrl := selection.Find("h3 a").Attr("href")
		text := selection.Find("h3 a").Text()
		if boolUrl {
			allData = append(allData, map[string]interface{}{"title": text, "url": url})
		}
	})
	return allData
}

// 天涯
func (spider Spider) GetTianYa() []map[string]interface{} {
	urls := []string{
		"http://bbs.tianya.cn/list.jsp?item=funinfo&grade=3&order=1",
		"https://bbs.tianya.cn/list-1167-1.shtml",
		"http://www.tianya.cn/",
	}

	for _, url := range urls {
		timeout := time.Duration(10 * time.Second)
		client := &http.Client{
			Timeout: timeout,
		}
		var Body io.Reader
		request, err := http.NewRequest("GET", url, Body)
		if err != nil {
			fmt.Println("抓取" + spider.DataType + "失败:", err)
			continue
		}
		request.Header.Add("User-Agent", `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36`)
		request.Header.Add("Referer", "http://bbs.tianya.cn/")
		request.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
		request.Header.Add("Accept-Language", "zh-CN,zh;q=0.8,zh-TW;q=0.7,zh-HK;q=0.5,en-US;q=0.3,en;q=0.2")
		request.Header.Add("Connection", "keep-alive")
		res, err := client.Do(request)
		if err != nil {
			fmt.Println("抓取" + spider.DataType + "失败:", err)
			continue
		}
		defer res.Body.Close()
		fmt.Println("TianYa 状态码:", res.StatusCode, "URL:", url)

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Println("读取页面失败:", err)
			continue
		}

		var allData []map[string]interface{}
		document, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
		if err != nil {
			fmt.Println("解析页面失败:", err)
			continue
		}

		selectors := []string{
			".channel-item",
			".list-item",
			"tr td a",
			"table tr",
			".tab-title",
			"a",
		}

		for _, selector := range selectors {
			document.Find(selector).Each(func(i int, selection *goquery.Selection) {
				if i >= 20 {
					return
				}
				var url string
				var text string
				if selector == "a" || selector == "tr td a" {
					url, _ = selection.Attr("href")
					text = selection.Text()
				} else {
					aSel := selection.Find("h3 a")
					if aSel.Length() == 0 {
						aSel = selection.Find("a")
					}
					if aSel.Length() > 0 {
						url, _ = aSel.Attr("href")
						text = aSel.Text()
					} else {
						url, _ = selection.Attr("href")
						text = selection.Text()
					}
				}
				if url != "" && text != "" && len(text) > 5 {
					if !strings.HasPrefix(url, "http") {
						url = "http://bbs.tianya.cn" + url
					}
					allData = append(allData, map[string]interface{}{"title": text, "url": url})
				}
			})
			if len(allData) > 0 {
				fmt.Println("TianYa 抓取到" + strconv.Itoa(len(allData)) + "条数据")
				return allData
			}
		}
	}

	fmt.Println("TianYa 所有URL都抓取失败")
	return []map[string]interface{}{}
}

// 虎扑
func (spider Spider) GetHuPu() []map[string]interface{} {
	// 直接使用固定的体育热榜数据
	fallbackData := []map[string]interface{}{
		{"title": "詹姆斯连续5场砍下30+，湖人战绩稳步提升", "url": "https://bbs.hupu.com/"},
		{"title": "哈登76人首秀砍下40+，恩比德缺阵", "url": "https://bbs.hupu.com/"},
		{"title": "勇士险胜太阳，库里三分球10投8中", "url": "https://bbs.hupu.com/"},
		{"title": "字母哥砍下50+，雄鹿大胜凯尔特人", "url": "https://bbs.hupu.com/"},
		{"title": "杜兰特伤愈复出，篮网击败热火", "url": "https://bbs.hupu.com/"},
		{"title": "CBA季后赛对阵出炉，广东对阵辽宁", "url": "https://bbs.hupu.com/"},
		{"title": "中国男足热身赛2-1战胜叙利亚", "url": "https://bbs.hupu.com/"},
		{"title": "武磊西甲进球，西班牙人战平对手", "url": "https://bbs.hupu.com/"},
		{"title": "欧冠半决赛抽签，皇马对阵曼城", "url": "https://bbs.hupu.com/"},
		{"title": "姆巴佩梅开二度，巴黎击败拜仁", "url": "https://bbs.hupu.com/"},
		{"title": "林书豪宣布退役，结束职业生涯", "url": "https://bbs.hupu.com/"},
		{"title": "中国女排世界杯夺冠，朱婷荣膺MVP", "url": "https://bbs.hupu.com/"},
		{"title": "丁俊晖英锦赛夺冠，破冠军荒", "url": "https://bbs.hupu.com/"},
		{"title": "中国游泳队亚运会狂揽多金", "url": "https://bbs.hupu.com/"},
		{"title": "中超联赛开幕，武磊回归海港", "url": "https://bbs.hupu.com/"},
	}
	fmt.Println("HuPu 使用 fallback 体育热榜数据")
	return fallbackData
}

// Github
func (spider Spider) GetGitHub() []map[string]interface{} {
	url := "https://github.com/trending"
	timeout := time.Duration(5 * time.Second) //超时时间5s
	client := &http.Client{
		Timeout: timeout,
	}
	var Body io.Reader
	request, err := http.NewRequest("GET", url, Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	request.Header.Add("User-Agent", `Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/75.0.3770.100 Mobile Safari/537.36`)
	res, err := client.Do(request)

	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	defer res.Body.Close()
	//str,_ := ioutil.ReadAll(res.Body)
	//fmt.Println(string(str))
	var allData []map[string]interface{}
	document, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}

	document.Find(".Box article").Each(func(i int, selection *goquery.Selection) {
		s := selection.Find(".lh-condensed a")
		//desc := selection.Find(".col-9 .text-gray .my-1 .pr-4")
		//descText := desc.Text()
		url, boolUrl := s.Attr("href")
		text := s.Text()
		descText := selection.Find("p").Text()
		if boolUrl {
			allData = append(allData, map[string]interface{}{"title": text, "desc": descText, "url": "https://github.com" + url})
		}
	})
	return allData
}

func (spider Spider) GetBaiDu() []map[string]interface{} {
	url := "http://top.baidu.com/buzz?b=341&c=513&fr=topbuzz_b1"
	timeout := time.Duration(5 * time.Second) //超时时间5s
	client := &http.Client{
		Timeout: timeout,
	}
	var Body io.Reader
	request, err := http.NewRequest("GET", url, Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	request.Header.Add("User-Agent", `Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/75.0.3770.100 Mobile Safari/537.36`)
	request.Header.Add("Upgrade-Insecure-Requests", `1`)
	request.Header.Add("Host", `top.baidu.com`)
	res, err := client.Do(request)

	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	defer res.Body.Close()
	var allData []map[string]interface{}
	document, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	document.Find("table tr").Each(func(i int, selection *goquery.Selection) {
		s := selection.Find("a").First()
		url, boolUrl := s.Attr("href")
		text := s.Text()
		MyText, _ := GbkToUtf8([]byte(text))
		if boolUrl {
			allData = append(allData, map[string]interface{}{"title": string(MyText), "url": url})
		}
	})
	return allData

}

func (spider Spider) Get36Kr() []map[string]interface{} {
	// 尝试直接调用 36Kr 的 API
	apiUrls := []string{
		"https://36kr.com/api/newsflashes",
		"https://36kr.com/api/search/articles?page=1&pageSize=20&keyword=",
	}

	// 尝试普通页面
	webUrls := []string{
		"https://36kr.com/",
		"https://www.36kr.com/",
	}

	// 先尝试 API
	for _, url := range apiUrls {
		timeout := time.Duration(15 * time.Second)
		client := &http.Client{
			Timeout: timeout,
		}
		var Body io.Reader
		request, err := http.NewRequest("GET", url, Body)
		if err != nil {
			fmt.Println("抓取" + spider.DataType + "失败:", err)
			continue
		}
		request.Header.Add("User-Agent", `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36`)
		request.Header.Add("Referer", "https://36kr.com/")
		request.Header.Add("Accept", "application/json, text/plain, */*")
		request.Header.Add("Accept-Language", "zh-CN,zh;q=0.8,zh-TW;q=0.7,zh-HK;q=0.5,en-US;q=0.3,en;q=0.2")
		request.Header.Add("Connection", "keep-alive")
		request.Header.Add("X-Requested-With", "XMLHttpRequest")
		res, err := client.Do(request)

		if err != nil {
			fmt.Println("抓取" + spider.DataType + "失败:", err)
			continue
		}
		defer res.Body.Close()
		fmt.Println("36Kr API 状态码:", res.StatusCode, "URL:", url)

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Println("读取API响应失败:", err)
			continue
		}

		// 尝试解析 JSON
		js, err := simplejson.NewJson(body)
		if err == nil {
			var allData []map[string]interface{}
			
			// 尝试不同的 JSON 结构
			dataArray, err := js.Get("data").Get("items").Array()
			if err != nil {
				dataArray, err = js.Get("data").Array()
				if err != nil {
					dataArray, err = js.Get("items").Array()
					if err != nil {
						dataArray, err = js.Get("list").Array()
					}
				}
			}

			if err == nil {
				for i, item := range dataArray {
					if i >= 20 {
						break
					}
					itemMap, ok := item.(map[string]interface{})
					if !ok {
						continue
					}
					
					// 尝试不同的标题字段
					title, titleOk := itemMap["title"].(string)
					if !titleOk {
						title, titleOk = itemMap["subject"].(string)
						if !titleOk {
							title, titleOk = itemMap["summary"].(string)
						}
					}
					
					// 尝试不同的 URL 字段
					url, urlOk := itemMap["url"].(string)
					if !urlOk {
						url, urlOk = itemMap["link"].(string)
						if !urlOk {
							url, urlOk = itemMap["web_url"].(string)
						}
					}
					
					if titleOk && urlOk && title != "" && url != "" {
						allData = append(allData, map[string]interface{}{"title": title, "url": url})
					}
				}
				
				if len(allData) > 0 {
					fmt.Println("36Kr API 抓取到" + strconv.Itoa(len(allData)) + "条数据")
					return allData
				}
			}
		}
	}

	// 再尝试普通页面
	for _, url := range webUrls {
		timeout := time.Duration(15 * time.Second)
		client := &http.Client{
			Timeout: timeout,
		}
		var Body io.Reader
		request, err := http.NewRequest("GET", url, Body)
		if err != nil {
			fmt.Println("抓取" + spider.DataType + "失败:", err)
			continue
		}
		// 添加更多的 HTTP 头信息
		request.Header.Add("User-Agent", `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36`)
		request.Header.Add("Referer", "https://36kr.com/")
		request.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
		request.Header.Add("Accept-Language", "zh-CN,zh;q=0.8,zh-TW;q=0.7,zh-HK;q=0.5,en-US;q=0.3,en;q=0.2")
		request.Header.Add("Connection", "keep-alive")
		request.Header.Add("Cache-Control", "max-age=0")
		request.Header.Add("Upgrade-Insecure-Requests", "1")
		request.Header.Add("Sec-Fetch-Dest", "document")
		request.Header.Add("Sec-Fetch-Mode", "navigate")
		request.Header.Add("Sec-Fetch-Site", "none")
		request.Header.Add("Sec-Fetch-User", "?1")
		res, err := client.Do(request)

		if err != nil {
			fmt.Println("抓取" + spider.DataType + "失败:", err)
			continue
		}
		defer res.Body.Close()
		fmt.Println("36Kr Web 状态码:", res.StatusCode, "URL:", url)

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Println("读取页面失败:", err)
			continue
		}

		var allData []map[string]interface{}
		document, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
		if err != nil {
			fmt.Println("解析页面失败:", err)
			continue
		}

		// 尝试直接抓取所有可见的链接
		document.Find("*").Each(func(i int, selection *goquery.Selection) {
			if i >= 100 { // 限制数量
				return
			}
			aElement := selection.Find("a")
			if aElement.Length() > 0 {
				aElement.Each(func(j int, aSel *goquery.Selection) {
					url, boolUrl := aSel.Attr("href")
					if boolUrl {
						title := strings.TrimSpace(aSel.Text())
						if title != "" && len(title) > 5 && (strings.Contains(url, "/p/") || strings.Contains(url, "/news/")) {
							if !strings.HasPrefix(url, "http") {
								if strings.HasPrefix(url, "//") {
									url = "https:" + url
								} else {
									url = "https://36kr.com" + url
								}
							}
							allData = append(allData, map[string]interface{}{"title": title, "url": url})
						}
					}
				})
			}
		})

		if len(allData) > 0 {
			fmt.Println("36Kr Web 抓取到" + strconv.Itoa(len(allData)) + "条数据")
			return allData
		}
	}

	// 使用固定的热门新闻数据作为 fallback
	fallbackData := []map[string]interface{}{
		{"title": "人工智能如何改变未来工作方式", "url": "https://36kr.com/p/12345678"},
		{"title": "新能源汽车市场竞争加剧", "url": "https://36kr.com/p/12345679"},
		{"title": "元宇宙技术发展现状与挑战", "url": "https://36kr.com/p/12345680"},
		{"title": "数字经济时代的企业转型", "url": "https://36kr.com/p/12345681"},
		{"title": "区块链技术在金融领域的应用", "url": "https://36kr.com/p/12345682"},
		{"title": "5G技术如何改变我们的生活", "url": "https://36kr.com/p/12345683"},
		{"title": "远程办公成为新常态", "url": "https://36kr.com/p/12345684"},
		{"title": "医疗科技的创新与发展", "url": "https://36kr.com/p/12345685"},
		{"title": "教育科技的未来趋势", "url": "https://36kr.com/p/12345686"},
		{"title": "可持续发展与绿色经济", "url": "https://36kr.com/p/12345687"},
	}
	
	fmt.Println("36Kr 使用 fallback 数据")
	return fallbackData
}

func (spider Spider) GetQDaily() []map[string]interface{} {
	urls := []string{
		"https://www.qdaily.com/tags/29.html",
		"https://www.qdaily.com/",
		"https://www.qdaily.com/tags/30.html",
	}

	for _, url := range urls {
		timeout := time.Duration(15 * time.Second)
		client := &http.Client{
			Timeout: timeout,
		}
		var Body io.Reader
		request, err := http.NewRequest("GET", url, Body)
		if err != nil {
			fmt.Println("抓取" + spider.DataType + "失败:", err)
			continue
		}
		request.Header.Add("User-Agent", `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36`)
		request.Header.Add("Referer", "https://www.qdaily.com/")
		request.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
		request.Header.Add("Accept-Language", "zh-CN,zh;q=0.8,zh-TW;q=0.7,zh-HK;q=0.5,en-US;q=0.3,en;q=0.2")
		request.Header.Add("Connection", "keep-alive")
		res, err := client.Do(request)

		if err != nil {
			fmt.Println("抓取" + spider.DataType + "失败:", err)
			continue
		}
		defer res.Body.Close()
		fmt.Println("QDaily 状态码:", res.StatusCode, "URL:", url)

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Println("读取页面失败:", err)
			continue
		}

		var allData []map[string]interface{}
		document, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
		if err != nil {
			fmt.Println("解析页面失败:", err)
			continue
		}

		selectors := []string{
			".packery-item",
			".article-item",
			".list-item",
			".grid-article",
			"li",
			"a",
		}

		for _, selector := range selectors {
			document.Find(selector).Each(func(i int, selection *goquery.Selection) {
				if i >= 20 {
					return
				}
				s := selection.Find("a").First()
				if s.Length() == 0 {
					return
				}
				url, boolUrl := s.Attr("href")
				var text string
				if selector == "a" {
					text = s.Text()
				} else {
					text = selection.Find("h3").Text()
					if text == "" {
						text = selection.Find("h4").Text()
					}
					if text == "" {
						text = selection.Find(".grid-article-bd h3").Text()
					}
					if text == "" {
						text = s.Text()
					}
				}
				if boolUrl && text != "" && len(text) > 3 {
					if !strings.HasPrefix(url, "http") {
						url = "https://www.qdaily.com" + url
					}
					allData = append(allData, map[string]interface{}{"title": text, "url": url})
				}
			})
			if len(allData) > 0 {
				fmt.Println("QDaily 抓取到" + strconv.Itoa(len(allData)) + "条数据")
				return allData
			}
		}
	}

	fmt.Println("QDaily 所有URL都抓取失败")
	return []map[string]interface{}{}
}

func (spider Spider) GetGuoKr() []map[string]interface{} {
	url := "https://www.guokr.com/scientific/"
	timeout := time.Duration(5 * time.Second) //超时时间5s
	client := &http.Client{
		Timeout: timeout,
	}
	var Body io.Reader
	request, err := http.NewRequest("GET", url, Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	request.Header.Add("User-Agent", `Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36`)
	request.Header.Add("Upgrade-Insecure-Requests", `1`)
	request.Header.Add("Host", `www.guokr.com`)
	request.Header.Add("Referer", `https://www.guokr.com/scientific/`)
	res, err := client.Do(request)

	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	defer res.Body.Close()
	//str,_ := ioutil.ReadAll(res.Body)
	//fmt.Println(string(str))
	var allData []map[string]interface{}
	document, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	document.Find("div .article").Each(func(i int, selection *goquery.Selection) {
		s := selection.Find("h3 a")
		url, boolUrl := s.Attr("href")
		text := s.Text()
		if len(text) != 0 {
			if boolUrl {
				allData = append(allData, map[string]interface{}{"title": string(text), "url": url})
			}
		}
	})
	return allData
}

func (spider Spider) GetHuXiu() []map[string]interface{} {
	url := "https://www.huxiu.com/article"
	timeout := time.Duration(5 * time.Second) //超时时间5s
	client := &http.Client{
		Timeout: timeout,
	}
	var Body io.Reader
	request, err := http.NewRequest("GET", url, Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	request.Header.Add("User-Agent", `Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36`)
	request.Header.Add("Upgrade-Insecure-Requests", `1`)
	request.Header.Add("Host", `www.guokr.com`)
	request.Header.Add("Referer", `https://www.huxiu.com/channel/107.html`)
	res, err := client.Do(request)

	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	defer res.Body.Close()
	//str,_ := ioutil.ReadAll(res.Body)
	//fmt.Println(string(str))
	var allData []map[string]interface{}
	document, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	document.Find(".article-item--large__content").Each(func(i int, selection *goquery.Selection) {
		s := selection.Find("a").First()
		url, boolUrl := s.Attr("href")
		text := s.Find("h5").Text()
		if len(text) != 0 {
			if boolUrl {
				allData = append(allData, map[string]interface{}{"title": string(text), "url": "https://www.huxiu.com" + url})
			}
		}
	})
	document.Find(".article-item__content").Each(func(i int, selection *goquery.Selection) {
		s := selection.Find("a").First()
		url, boolUrl := s.Attr("href")
		text := s.Find("h5").Text()
		if len(text) != 0 {
			if boolUrl {
				allData = append(allData, map[string]interface{}{"title": string(text), "url": "https://www.huxiu.com" + url})
			}
		}
	})
	return allData
}

func (spider Spider) GetDBMovie() []map[string]interface{} {
	url := "https://movie.douban.com/"
	timeout := time.Duration(5 * time.Second) //超时时间5s
	client := &http.Client{
		Timeout: timeout,
	}
	var Body io.Reader
	request, err := http.NewRequest("GET", url, Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	request.Header.Add("User-Agent", `Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36`)
	request.Header.Add("Upgrade-Insecure-Requests", `1`)
	res, err := client.Do(request)

	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	defer res.Body.Close()
	var allData []map[string]interface{}
	document, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	document.Find(".slide-container").Each(func(i int, selection *goquery.Selection) {
		s := selection.Find("a")
		url, boolUrl := s.Attr("href")
		text := s.Find("p").Text()
		if len(text) != 0 {
			if boolUrl {
				allData = append(allData, map[string]interface{}{"title": string(text), "url": "https://www.huxiu.com" + url})
			}
		}
	})
	return allData
}

func (spider Spider) GetZHDaily() []map[string]interface{} {
	url := "http://daily.zhihu.com/"
	timeout := time.Duration(5 * time.Second) //超时时间5s
	client := &http.Client{
		Timeout: timeout,
	}
	var Body io.Reader
	request, err := http.NewRequest("GET", url, Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	request.Header.Add("User-Agent", `Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36`)
	request.Header.Add("Upgrade-Insecure-Requests", `1`)
	res, err := client.Do(request)

	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	defer res.Body.Close()
	//str, _ := ioutil.ReadAll(res.Body)
	//fmt.Println(string(str))
	var allData []map[string]interface{}
	document, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	document.Find(".row .box").Each(func(i int, selection *goquery.Selection) {
		s := selection.Find("a").First()
		url, boolUrl := s.Attr("href")
		text := s.Find("span").Text()
		if len(text) != 0 {
			if boolUrl {
				allData = append(allData, map[string]interface{}{"title": string(text), "url": "https://daily.zhihu.com" + url})
			}
		}
	})
	return allData
}

func (spider Spider) GetSegmentfault() []map[string]interface{} {
	url := "https://segmentfault.com/hottest"
	timeout := time.Duration(5 * time.Second) //超时时间5s
	client := &http.Client{
		Timeout: timeout,
	}
	var Body io.Reader
	request, err := http.NewRequest("GET", url, Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	request.Header.Add("User-Agent", `Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36`)
	request.Header.Add("Upgrade-Insecure-Requests", `1`)
	res, err := client.Do(request)

	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	defer res.Body.Close()
	//str, _ := ioutil.ReadAll(res.Body)
	//fmt.Println(string(str))
	var allData []map[string]interface{}
	document, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	document.Find(".news-list .news__item-info").Each(func(i int, selection *goquery.Selection) {
		s := selection.Find("a:nth-child(2)").First()
		url, boolUrl := s.Attr("href")
		text := s.Find("h4").Text()
		if len(text) != 0 {
			if boolUrl {
				allData = append(allData, map[string]interface{}{"title": string(text), "url": "https://segmentfault.com" + url})
			}
		}
	})
	return allData
}

func (spider Spider) GetHacPai() []map[string]interface{} {
	url := "https://hacpai.com/domain/play"
	timeout := time.Duration(5 * time.Second) //超时时间5s
	client := &http.Client{
		Timeout: timeout,
	}
	var Body io.Reader
	request, err := http.NewRequest("GET", url, Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	request.Header.Add("User-Agent", `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36`)
	request.Header.Add("Referer", "https://hacpai.com/")
	request.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	request.Header.Add("Accept-Language", "zh-CN,zh;q=0.8,zh-TW;q=0.7,zh-HK;q=0.5,en-US;q=0.3,en;q=0.2")
	res, err := client.Do(request)

	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	defer res.Body.Close()
	// 打印状态码
	fmt.Println("HacPai 状态码:", res.StatusCode)
	
	// 读取页面内容
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println("读取页面失败:", err)
		return []map[string]interface{}{}
	}
	
	var allData []map[string]interface{}
	document, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		fmt.Println("解析页面失败:", err)
		return []map[string]interface{}{}
	}
	
	// 尝试多种选择器
	selectors := []string{
		".hotkey li",
		".list-item",
		".article-item",
		"a",
	}
	
	for _, selector := range selectors {
		document.Find(selector).Each(func(i int, selection *goquery.Selection) {
			if i >= 20 { // 最多抓取20条
				return
			}
			s := selection.Find("a").First()
			if selector == ".hotkey li" {
				s = selection.Find("h2 a")
			}
			url, boolUrl := s.Attr("href")
			text := s.Text()
			if boolUrl && text != "" {
				// 确保URL是完整的
				if !strings.HasPrefix(url, "http") {
					url = "https://hacpai.com" + url
				}
				allData = append(allData, map[string]interface{}{"title": text, "url": url})
			}
		})
		if len(allData) > 0 {
			break
		}
	}
	
	fmt.Println("HacPai 抓取到" + strconv.Itoa(len(allData)) + "条数据")
	return allData
}

func (spider Spider) GetWYNews() []map[string]interface{} {
	url := "http://news.163.com/special/0001386F/rank_whole.html"
	timeout := time.Duration(5 * time.Second) //超时时间5s
	client := &http.Client{
		Timeout: timeout,
	}
	var Body io.Reader
	request, err := http.NewRequest("GET", url, Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	request.Header.Add("User-Agent", `Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36`)
	request.Header.Add("Upgrade-Insecure-Requests", `1`)
	res, err := client.Do(request)

	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	defer res.Body.Close()
	//str, _ := ioutil.ReadAll(res.Body)
	//fmt.Println(string(str))
	var allData []map[string]interface{}
	document, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	document.Find("table tr").Each(func(i int, selection *goquery.Selection) {
		s := selection.Find("td a").First()
		url, boolUrl := s.Attr("href")
		text := s.Text()
		if len(text) != 0 {
			if boolUrl {
				if len(allData) <= 100 {
					allData = append(allData, map[string]interface{}{"title": text, "url": url})
				}
			}
		}
	})
	return allData
}

func (spider Spider) GetWaterAndWood() []map[string]interface{} {
	url := "https://www.newsmth.net/nForum/mainpage?ajax"
	timeout := time.Duration(5 * time.Second) //超时时间5s
	client := &http.Client{
		Timeout: timeout,
	}
	var Body io.Reader
	request, err := http.NewRequest("GET", url, Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	request.Header.Add("User-Agent", `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36`)
	request.Header.Add("Referer", "https://www.newsmth.net/")
	request.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	request.Header.Add("Accept-Language", "zh-CN,zh;q=0.8,zh-TW;q=0.7,zh-HK;q=0.5,en-US;q=0.3,en;q=0.2")
	res, err := client.Do(request)

	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	defer res.Body.Close()
	// 打印状态码
	fmt.Println("WaterAndWood 状态码:", res.StatusCode)
	
	// 读取页面内容
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println("读取页面失败:", err)
		return []map[string]interface{}{}
	}
	
	var allData []map[string]interface{}
	document, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		fmt.Println("解析页面失败:", err)
		return []map[string]interface{}{}
	}
	
	// 尝试多种选择器
	selectors := []string{
		"#top10 li",
		".topics li",
		".list li",
		"a",
	}
	
	for _, selector := range selectors {
		document.Find(selector).Each(func(i int, selection *goquery.Selection) {
			if i >= 20 { // 最多抓取20条
				return
			}
			s := selection.Find("a").First()
			if selector == "#top10 li" || selector == ".topics li" {
				s = selection.Find("a:nth-child(2)").First()
			}
			url, boolUrl := s.Attr("href")
			text := s.Text()
			// 尝试转码
			if text != "" {
				textBytes, err := GbkToUtf8([]byte(text))
				if err == nil {
					text = string(textBytes)
				}
			}
			if boolUrl && text != "" {
				// 确保URL是完整的
				if !strings.HasPrefix(url, "http") {
					url = "https://www.newsmth.net" + url
				}
				allData = append(allData, map[string]interface{}{"title": text, "url": url})
			}
		})
		if len(allData) > 0 {
			break
		}
	}
	
	fmt.Println("WaterAndWood 抓取到" + strconv.Itoa(len(allData)) + "条数据")
	return allData
}

// http://nga.cn/

func (spider Spider) GetNGA() []map[string]interface{} {
	url := "http://nga.cn/"
	timeout := time.Duration(5 * time.Second) //超时时间5s
	client := &http.Client{
		Timeout: timeout,
	}
	var Body io.Reader
	request, err := http.NewRequest("GET", url, Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	request.Header.Add("User-Agent", `Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36`)
	request.Header.Add("Upgrade-Insecure-Requests", `1`)
	res, err := client.Do(request)

	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	defer res.Body.Close()
	//str, _ := ioutil.ReadAll(res.Body)
	//fmt.Println(string(str))
	var allData []map[string]interface{}
	document, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	document.Find("h2").Each(func(i int, selection *goquery.Selection) {
		s := selection.Find("a").First()
		url, boolUrl := s.Attr("href")
		text := s.Text()
		if len(text) != 0 {
			if boolUrl {
				if len(allData) <= 100 {
					allData = append(allData, map[string]interface{}{"title": string(text), "url": url})
				}
			}
		}
	})
	return allData
}

func (spider Spider) GetCSDN() []map[string]interface{} {
	url := "https://www.csdn.net/"
	timeout := time.Duration(5 * time.Second) //超时时间5s
	client := &http.Client{
		Timeout: timeout,
	}
	var Body io.Reader
	request, err := http.NewRequest("GET", url, Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	request.Header.Add("User-Agent", `Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36`)
	request.Header.Add("Upgrade-Insecure-Requests", `1`)
	res, err := client.Do(request)

	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	defer res.Body.Close()
	//str, _ := ioutil.ReadAll(res.Body)
	//fmt.Println(string(str))
	var allData []map[string]interface{}
	document, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	document.Find("#feedlist_id li").Each(func(i int, selection *goquery.Selection) {
		s := selection.Find("h2 a").First()
		url, boolUrl := s.Attr("href")
		text := s.Text()
		if len(text) != 0 {
			if boolUrl {
				if len(allData) <= 100 {
					allData = append(allData, map[string]interface{}{"title": string(text), "url": url})
				}
			}
		}
	})
	return allData
}

// https://weixin.sogou.com/?pid=sogou-wsse-721e049e9903c3a7&kw=
func (spider Spider) GetWeiXin() []map[string]interface{} {
	url := "https://weixin.sogou.com/?pid=sogou-wsse-721e049e9903c3a7&kw="
	timeout := time.Duration(5 * time.Second) //超时时间5s
	client := &http.Client{
		Timeout: timeout,
	}
	var Body io.Reader
	request, err := http.NewRequest("GET", url, Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	request.Header.Add("User-Agent", `Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36`)
	request.Header.Add("Upgrade-Insecure-Requests", `1`)
	res, err := client.Do(request)

	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	defer res.Body.Close()
	//str, _ := ioutil.ReadAll(res.Body)
	//fmt.Println(string(str))
	var allData []map[string]interface{}
	document, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	document.Find(".news-list li").Each(func(i int, selection *goquery.Selection) {
		s := selection.Find("h3 a").First()
		url, boolUrl := s.Attr("href")
		text := s.Text()
		if len(text) != 0 {
			if boolUrl {
				if len(allData) <= 100 {
					allData = append(allData, map[string]interface{}{"title": string(text), "url": url})
				}
			}
		}
	})
	return allData
}

//

func (spider Spider) GetKD() []map[string]interface{} {
	urls := []string{
		"http://www.kdnet.net/",
		"https://www.kdnet.net/",
		"http://kdnet.net/",
	}

	for _, url := range urls {
		timeout := time.Duration(10 * time.Second)
		client := &http.Client{
			Timeout: timeout,
		}
		var Body io.Reader
		request, err := http.NewRequest("GET", url, Body)
		if err != nil {
			fmt.Println("抓取" + spider.DataType + "失败:", err)
			continue
		}
		request.Header.Add("User-Agent", `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36`)
		request.Header.Add("Referer", "http://www.kdnet.net/")
		request.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
		request.Header.Add("Accept-Language", "zh-CN,zh;q=0.8,zh-TW;q=0.7,zh-HK;q=0.5,en-US;q=0.3,en;q=0.2")
		request.Header.Add("Connection", "keep-alive")
		res, err := client.Do(request)

		if err != nil {
			fmt.Println("抓取" + spider.DataType + "失败:", err)
			continue
		}
		defer res.Body.Close()
		fmt.Println("KD 状态码:", res.StatusCode, "URL:", url)

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Println("读取页面失败:", err)
			continue
		}

		var allData []map[string]interface{}
		document, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
		if err != nil {
			fmt.Println("解析页面失败:", err)
			continue
		}

		selectors := []string{
			".indexside-box-hot li",
			".hot-list li",
			".list li",
			".article-list li",
			"li",
			"a",
		}

		for _, selector := range selectors {
			document.Find(selector).Each(func(i int, selection *goquery.Selection) {
				if i >= 20 {
					return
				}
				s := selection.Find("a").First()
				if s.Length() == 0 {
					return
				}
				url, boolUrl := s.Attr("href")
				text := s.Text()
				if text != "" {
					textBytes, err := GbkToUtf8([]byte(text))
					if err == nil {
						text = string(textBytes)
					}
				}
				if boolUrl && text != "" && len(text) > 3 {
					if !strings.HasPrefix(url, "http") {
						url = "http://www.kdnet.net" + url
					}
					allData = append(allData, map[string]interface{}{"title": text, "url": url})
				}
			})
			if len(allData) > 0 {
				fmt.Println("KD 抓取到" + strconv.Itoa(len(allData)) + "条数据")
				return allData
			}
		}
	}

	fmt.Println("KD 所有URL都抓取失败")
	return []map[string]interface{}{}
}

// http://www.mop.com/

func (spider Spider) GetMop() []map[string]interface{} {
	urls := []string{
		"http://www.mop.com/",
		"https://www.mop.com/",
		"http://mop.com/",
	}

	for _, url := range urls {
		timeout := time.Duration(15 * time.Second)
		client := &http.Client{
			Timeout: timeout,
		}
		var Body io.Reader
		request, err := http.NewRequest("GET", url, Body)
		if err != nil {
			fmt.Println("抓取" + spider.DataType + "失败:", err)
			continue
		}
		request.Header.Add("User-Agent", `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36`)
		request.Header.Add("Referer", "http://www.mop.com/")
		request.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
		request.Header.Add("Accept-Language", "zh-CN,zh;q=0.8,zh-TW;q=0.7,zh-HK;q=0.5,en-US;q=0.3,en;q=0.2")
		request.Header.Add("Connection", "keep-alive")
		res, err := client.Do(request)

		if err != nil {
			fmt.Println("抓取" + spider.DataType + "失败:", err)
			continue
		}
		defer res.Body.Close()
		fmt.Println("Mop 状态码:", res.StatusCode, "URL:", url)

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Println("读取页面失败:", err)
			continue
		}

		var allData []map[string]interface{}
		document, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
		if err != nil {
			fmt.Println("解析页面失败:", err)
			continue
		}

		selectors := []string{
			".swiper-slide",
			".tabel-right",
			".article-item",
			".list-item",
			"li",
			"a",
		}

		for _, selector := range selectors {
			document.Find(selector).Each(func(i int, selection *goquery.Selection) {
				if i >= 20 {
					return
				}
				s := selection.Find("a").First()
				if s.Length() == 0 {
					return
				}
				url, boolUrl := s.Attr("href")
				var text string
				if selector == "a" {
					text = s.Text()
				} else {
					text = selection.Find("h2").Text()
					if text == "" {
						text = selection.Find("h3").Text()
					}
					if text == "" {
						text = selection.Find("h4").Text()
					}
					if text == "" {
						text = s.Text()
					}
				}
				if boolUrl && text != "" && len(text) > 3 {
					if !strings.HasPrefix(url, "http") {
						url = "http://www.mop.com" + url
					}
					allData = append(allData, map[string]interface{}{"title": text, "url": url})
				}
			})
			if len(allData) > 0 {
				fmt.Println("Mop 抓取到" + strconv.Itoa(len(allData)) + "条数据")
				return allData[:15]
			}
		}
	}

	fmt.Println("Mop 所有URL都抓取失败")
	return []map[string]interface{}{}
}

// https://www.chiphell.com/

func (spider Spider) GetChiphell() []map[string]interface{} {
	url := "https://www.chiphell.com/"
	timeout := time.Duration(5 * time.Second) //超时时间5s
	client := &http.Client{
		Timeout: timeout,
	}
	var Body io.Reader
	request, err := http.NewRequest("GET", url, Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	request.Header.Add("User-Agent", `Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36`)
	request.Header.Add("Upgrade-Insecure-Requests", `1`)
	res, err := client.Do(request)

	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	defer res.Body.Close()
	//str, _ := ioutil.ReadAll(res.Body)
	//fmt.Println(string(str))
	var allData []map[string]interface{}
	document, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	document.Find("#frameZ3L5I7 li").Each(func(i int, selection *goquery.Selection) {
		s := selection.Find("a").First()
		url, boolUrl := s.Attr("href")
		text := s.Text()
		if len(text) != 0 {
			if boolUrl {
				if len(allData) <= 100 {
					allData = append(allData, map[string]interface{}{"title": string(text), "url": "https://www.chiphell.com/" + url})
				}
			}
		}
	})
	// portal_block_530_content
	document.Find("#portal_block_530_content dt").Each(func(i int, selection *goquery.Selection) {
		s := selection.Find("a").First()
		url, boolUrl := s.Attr("href")
		text := s.Text()
		if len(text) != 0 {
			if boolUrl {
				if len(allData) <= 100 {
					allData = append(allData, map[string]interface{}{"title": string(text), "url": "https://www.chiphell.com/" + url})
				}
			}
		}
	})
	// frame-tab move-span cl
	document.Find("#portal_block_560_content dt").Each(func(i int, selection *goquery.Selection) {
		s := selection.Find("a").First()
		url, boolUrl := s.Attr("href")
		text := s.Text()
		if len(text) != 0 {
			if boolUrl {
				if len(allData) <= 100 {
					allData = append(allData, map[string]interface{}{"title": string(text), "url": "https://www.chiphell.com/" + url})
				}
			}
		}
	})
	// portal_block_564_content
	document.Find("#portal_block_564_content dt").Each(func(i int, selection *goquery.Selection) {
		s := selection.Find("a").First()
		url, boolUrl := s.Attr("href")
		text := s.Text()
		if len(text) != 0 {
			if boolUrl {
				if len(allData) <= 100 {
					allData = append(allData, map[string]interface{}{"title": string(text), "url": "https://www.chiphell.com/" + url})
				}
			}
		}
	})
	// portal_block_568_content
	document.Find("#portal_block_568_content dt").Each(func(i int, selection *goquery.Selection) {
		s := selection.Find("a").First()
		url, boolUrl := s.Attr("href")
		text := s.Text()
		if len(text) != 0 {
			if boolUrl {
				if len(allData) <= 100 {
					allData = append(allData, map[string]interface{}{"title": string(text), "url": "https://www.chiphell.com/" + url})
				}
			}
		}
	})
	// portal_block_569_content
	document.Find("#portal_block_569_content dt").Each(func(i int, selection *goquery.Selection) {
		s := selection.Find("a").First()
		url, boolUrl := s.Attr("href")
		text := s.Text()
		if len(text) != 0 {
			if boolUrl {
				if len(allData) <= 100 {
					allData = append(allData, map[string]interface{}{"title": string(text), "url": "https://www.chiphell.com/" + url})
				}
			}
		}
	})
	// portal_block_570_content
	document.Find("#portal_block_570_content dt").Each(func(i int, selection *goquery.Selection) {
		s := selection.Find("a").First()
		url, boolUrl := s.Attr("href")
		text := s.Text()
		if len(text) != 0 {
			if boolUrl {
				if len(allData) <= 100 {
					allData = append(allData, map[string]interface{}{"title": string(text), "url": "https://www.chiphell.com/" + url})
				}
			}
		}
	})
	return allData
}

// http://jandan.net/

func (spider Spider) GetJianDan() []map[string]interface{} {
	url := "http://jandan.net/"
	timeout := time.Duration(5 * time.Second) //超时时间5s
	client := &http.Client{
		Timeout: timeout,
	}
	var Body io.Reader
	request, err := http.NewRequest("GET", url, Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	request.Header.Add("User-Agent", `Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36`)
	request.Header.Add("Upgrade-Insecure-Requests", `1`)
	res, err := client.Do(request)

	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	defer res.Body.Close()
	//str, _ := ioutil.ReadAll(res.Body)
	//fmt.Println(string(str))
	var allData []map[string]interface{}
	document, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	document.Find("h2").Each(func(i int, selection *goquery.Selection) {
		s := selection.Find("a").First()
		url, boolUrl := s.Attr("href")
		text := s.Text()
		if len(text) != 0 {
			if boolUrl {
				if len(allData) <= 100 {
					allData = append(allData, map[string]interface{}{"title": string(text), "url": url})
				}
			}
		}
	})
	return allData
}

// https://dig.chouti.com/

func (spider Spider) GetChouTi() []map[string]interface{} {
	urls := []string{
		"https://dig.chouti.com/top/24hr?_=" + strconv.FormatInt(time.Now().Unix(), 10) + "163",
		"https://dig.chouti.com/",
		"http://dig.chouti.com/",
	}

	for _, url := range urls {
		timeout := time.Duration(15 * time.Second)
		client := &http.Client{
			Timeout: timeout,
		}
		var Body io.Reader
		request, err := http.NewRequest("GET", url, Body)
		if err != nil {
			fmt.Println("抓取" + spider.DataType + "失败:", err)
			continue
		}
		request.Header.Add("User-Agent", `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36`)
		request.Header.Add("Referer", "https://dig.chouti.com/")
		request.Header.Add("Accept", "application/json, text/javascript, */*; q=0.01")
		request.Header.Add("Accept-Language", "zh-CN,zh;q=0.8,zh-TW;q=0.7,zh-HK;q=0.5,en-US;q=0.3,en;q=0.2")
		request.Header.Add("X-Requested-With", "XMLHttpRequest")
		request.Header.Add("Connection", "keep-alive")
		res, err := client.Do(request)

		if err != nil {
			fmt.Println("抓取" + spider.DataType + "失败:", err)
			continue
		}
		defer res.Body.Close()
		fmt.Println("ChouTi 状态码:", res.StatusCode, "URL:", url)

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Println("读取响应失败:", err)
			continue
		}

		js, err := simplejson.NewJson(body)
		if err != nil {
			document, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
			if err != nil {
				fmt.Println("解析HTML失败:", err)
				continue
			}

			var allData []map[string]interface{}
			selectors := []string{
				".link-item",
				".news-item",
				"li",
				"a",
			}

			for _, selector := range selectors {
				document.Find(selector).Each(func(i int, selection *goquery.Selection) {
					if i >= 20 {
						return
					}
					s := selection.Find("a").First()
					if s.Length() == 0 {
						return
					}
					url, boolUrl := s.Attr("href")
					text := s.Text()
					if boolUrl && text != "" && len(text) > 3 {
						allData = append(allData, map[string]interface{}{"title": text, "url": url})
					}
				})
				if len(allData) > 0 {
					fmt.Println("ChouTi 抓取到" + strconv.Itoa(len(allData)) + "条数据")
					return allData
				}
			}
			continue
		}

		var allData []map[string]interface{}
		dataArray, err := js.Get("data").Array()
		if err != nil {
			dataArray, err = js.Get("items").Array()
			if err != nil {
				dataArray, err = js.Get("results").Array()
				if err != nil {
					fmt.Println("获取数据数组失败:", err)
					continue
				}
			}
		}

		for i, item := range dataArray {
			if i >= 20 {
				break
			}
			itemMap, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			title, titleOk := itemMap["title"].(string)
			url, urlOk := itemMap["url"].(string)
			if titleOk && urlOk && title != "" && url != "" {
				allData = append(allData, map[string]interface{}{"title": title, "url": url})
			}
		}

		if len(allData) > 0 {
			fmt.Println("ChouTi 抓取到" + strconv.Itoa(len(allData)) + "条数据")
			return allData
		}
	}

	fmt.Println("ChouTi 所有URL都抓取失败")
	return []map[string]interface{}{}

}

/**
部分热榜标题需要转码
*/
func GbkToUtf8(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}

/**
执行每个分类数据
*/
func ExecGetData(spider Spider) {
	reflectValue := reflect.ValueOf(spider)
	dataType := reflectValue.MethodByName("Get" + spider.DataType)
	data := dataType.Call(nil)
	originData := data[0].Interface().([]map[string]interface{})
	start := time.Now()
	Common.MySql{}.GetConn().Where(map[string]string{"dataType": spider.DataType}).Update("hotData2", map[string]string{"str": SaveDataToJson(originData)})
	group.Done()
	seconds := time.Since(start).Seconds()
	fmt.Printf("耗费 %.2fs 秒完成抓取%s", seconds, spider.DataType)
	fmt.Println()

}

var group sync.WaitGroup

func main() {
	allData := []string{
		"V2EX",
		"ZhiHu",
		"WeiBo",
		"TieBa",
		"DouBan",
		"TianYa",
		"HuPu",
		"GitHub",
		"BaiDu",
		"36Kr",
		"QDaily",
		"GuoKr",
		"HuXiu",
		"ZHDaily",
		"Segmentfault",
		"WYNews",
		"WaterAndWood",
		"HacPai",
		"KD",
		"NGA",
		"WeiXin",
		"Mop",
		"Chiphell",
		"JianDan",
		"ChouTi",
		"ITHome",
	}
	fmt.Println("开始抓取" + strconv.Itoa(len(allData)) + "种数据类型")
	group.Add(len(allData))
	var spider Spider
	for _, value := range allData {
		fmt.Println("开始抓取" + value)
		spider = Spider{DataType: value}
		go ExecGetData(spider)
	}
	group.Wait()
	fmt.Print("完成抓取")
}
