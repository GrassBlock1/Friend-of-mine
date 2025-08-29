package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url" // 导入 url 包来处理相对路径
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/olekukonko/tablewriter"
	"golang.org/x/net/html"
	"gopkg.in/yaml.v3"
)

// 定义友链页面的关键词
var friendPageKeywords = []string{"友链", "友人", "朋友", "链接", "links", "friends", "partners"}

type LinkInfo struct {
	Link        string `yaml:"link"`
	Avatar      string `yaml:"avatar"`
	Description string `yaml:"description"`
}

type CheckResult struct {
	Name             string
	URL              string
	Status           string
	BacklinkFound    bool
	HTMLSnippet      string
	BacklinkLocation string // 新增字段：记录 Backlink 的位置 (e.g., "Homepage", "https://.../links")
	Error            string
}

func main() {
	yamlPath := flag.String("file", "links.yaml", "包含链接的 YAML 文件路径")
	myBlogURL := flag.String("url", "", "你的博客 URL，用于检测返回链接 (必需)")
	oldURL := flag.String("ourl", "", "你的博客旧 URL（如果有的话）")
	concurrency := flag.Int("c", 5, "并发检测的数量")
	flag.Parse()

	if *myBlogURL == "" {
		fmt.Println("错误：必须提供你的博客 URL。")
		flag.Usage()
		os.Exit(1)
	}

	yamlFile, err := os.ReadFile(*yamlPath)
	if err != nil {
		log.Fatalf("无法读取 YAML 文件 %s: %v", *yamlPath, err)
	}

	var links map[string]LinkInfo
	if err := yaml.Unmarshal(yamlFile, &links); err != nil {
		log.Fatalf("无法解析 YAML: %v", err)
	}

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, *concurrency)
	resultsChan := make(chan CheckResult, len(links))

	log.Printf("开始检测 %d 个链接，并发数: %d...\n", len(links), *concurrency)

	for name, info := range links {
		wg.Add(1)
		go checkLink(name, info, *myBlogURL, *oldURL, &wg, semaphore, resultsChan)
	}

	wg.Wait()
	close(resultsChan)

	var results []CheckResult
	for result := range resultsChan {
		results = append(results, result)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

	printResults(results, *myBlogURL)
}

// checkLink 函数被重构以支持两步检测
func checkLink(name string, info LinkInfo, myBlogURL string, oldURL string, wg *sync.WaitGroup, semaphore chan struct{}, resultsChan chan<- CheckResult) {
	defer wg.Done()
	semaphore <- struct{}{}
	defer func() { <-semaphore }()

	result := CheckResult{Name: name, URL: info.Link}
	client := &http.Client{Timeout: 10 * time.Second}

	// === 第一步：检查原始 URL ===
	resp, err := client.Get(info.Link)
	if err != nil {
		result.Status = "Offline"
		result.Error = err.Error()
		resultsChan <- result
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		result.Status = fmt.Sprintf("Error (%d)", resp.StatusCode)
		result.Error = fmt.Sprintf("HTTP 状态码: %s", resp.Status)
		resultsChan <- result
		return
	}
	result.Status = "Online"

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Error = "无法读取响应内容"
		resultsChan <- result
		return
	}

	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		result.Error = "无法解析 HTML"
		resultsChan <- result
		return
	}

	// 在首页查找
	// 在首页查找
	snippet, found, isOld := findBacklink(doc, myBlogURL, oldURL)
	if found {
		result.BacklinkFound = true
		result.HTMLSnippet = snippet
		if isOld {
			result.BacklinkLocation = "Homepage (OLD)"
		} else {
			result.BacklinkLocation = "Homepage"
		}
		resultsChan <- result
		return // 找到了，任务结束
	}

	// === 第二步：如果首页没找到，则查找友链页面并检查 ===
	// resp.Request.URL 包含了重定向后的最终 URL，是解析相对路径的正确基准
	friendsPageURL, found := findFriendsPageURL(doc, resp.Request.URL)
	if !found || strings.HasPrefix(friendsPageURL, "javascript") {
		// 尝试 fallback 链接
		fallbackPaths := []string{"/link/", "/links", "/friends", "/link", "/友链", "/links.html", "/friends.html"}
		for _, path := range fallbackPaths {
			fallbackURL := resp.Request.URL.ResolveReference(&url.URL{Path: path})
			log.Printf("[%s] 尝试可能的友链页: %s", name, fallbackURL.String())

			// 检查 fallback URL 是否可访问
			if checkFallbackURL(client, fallbackURL.String()) {
				friendsPageURL = fallbackURL.String()
				found = true
				break
			}
		}

		if !found {
			// 没找到友链页面，任务结束
			resultsChan <- result
			return
		}
	}

	log.Printf("[%s] 在首页未找到链接, 尝试访问友链页: %s", name, friendsPageURL)

	// 访问友链页面
	resp2, err := client.Get(friendsPageURL)
	if err != nil {
		result.Error = "访问友链页失败: " + err.Error()
		resultsChan <- result
		return
	}
	defer resp2.Body.Close()

	if resp2.StatusCode >= 400 {
		result.Error = fmt.Sprintf("友链页返回错误状态码: %d", resp2.StatusCode)
		resultsChan <- result
		return
	}

	doc2, err := html.Parse(resp2.Body)
	if err != nil {
		result.Error = "无法解析友链页 HTML"
		resultsChan <- result
		return
	}

	// 在友链页面上再次查找
	snippet, found, isOld = findBacklink(doc2, myBlogURL, oldURL)
	if found {
		result.BacklinkFound = true
		result.HTMLSnippet = snippet
		if isOld {
			result.BacklinkLocation = friendsPageURL + " (OLD)"
		} else {
			result.BacklinkLocation = friendsPageURL
		}
	} else {
		rdoc, err := getRenderedHTML(friendsPageURL)
		if err == nil {
			doc3, err := html.Parse(strings.NewReader(rdoc))
			if err == nil {
				snippet, found, isOld = findBacklink(doc3, myBlogURL, oldURL)
				if found {
					result.BacklinkFound = true
					result.HTMLSnippet = snippet
					if isOld {
						result.BacklinkLocation = friendsPageURL + " (渲染后) (OLD)"
					} else {
						result.BacklinkLocation = friendsPageURL + " (渲染后)"
					}
				}
			}
		}
	}

	resultsChan <- result
}

var (
	browserInstance *rod.Browser
	browserOnce     sync.Once
)

func getBrowser() *rod.Browser {
	browserOnce.Do(func() {
		var options string
		if os.Getenv("SHOW_BROWSER") == "true" {
			options = launcher.New().
				Headless(true).
				Headless(false).
				MustLaunch()
		} else {
			options = launcher.New().MustLaunch()
		}
		browserInstance = rod.New().ControlURL(options)

		// 检查环境变量，如果存在则不使用无头模式
		if err := browserInstance.Connect(); err != nil {
			log.Printf("浏览器连接失败: %v", err)
			browserInstance = nil
		}
	})
	return browserInstance
}

func getRenderedHTML(url string) (string, error) {
	browser := getBrowser()
	if browser == nil {
		return "", fmt.Errorf("浏览器实例不可用")
	}

	page := browser.MustPage()

	// 拦截图片请求以加快页面加载
	go page.HijackRequests().MustAdd("*", func(ctx *rod.Hijack) {
		if ctx.Request.Type() == proto.NetworkResourceTypeImage {
			// 返回一个透明的 1x1 PNG 图片，节省带宽的同时解决使用 friends-circle-lite 的站点因为图片加载失败反复重试导致卡死的问题
			ctx.Response.SetBody([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4, 0x89, 0x00, 0x00, 0x00, 0x0D, 0x49, 0x44, 0x41, 0x54, 0x78, 0xDA, 0x63, 0x64, 0x60, 0xF8, 0x5F, 0x0F, 0x00, 0x08, 0x70, 0x01, 0x80, 0xEB, 0x47, 0xBA, 0x92, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, 0x44, 0xAE, 0x42, 0x60, 0x82})
			return
		}
		if ctx.Request.Type() == proto.NetworkResourceTypeFont {
			// 阻止第三方字体加载以提高页面加载速度
			ctx.Response.Fail(proto.NetworkErrorReasonBlockedByClient)
			return
		}
		ctx.ContinueRequest(&proto.FetchContinueRequest{})
	}).Run()

	page = page.MustNavigate(url)

	if err := page.WaitLoad(); err != nil {
		return "", err
	}

	// 等待额外时间让动态内容加载
	page.MustWaitIdle().Timeout(20 * time.Second).MustWaitStable().Timeout(10 * time.Second).MustWaitNavigation()

	htmlContent, err := page.HTML()
	defer page.MustClose() // 检测完成后关闭页面
	if err != nil {
		return "", err
	}
	return htmlContent, nil
}

// findBacklink 函数检查当前URL和旧URL
func findBacklink(n *html.Node, targetURL string, oldURL string) (string, bool, bool) {
	if n.Type == html.ElementNode && n.Data == "a" {
		for _, attr := range n.Attr {
			if attr.Key == "href" {
				// 首先检查当前URL
				targetParsed, err1 := url.Parse(targetURL)
				hrefParsed, err2 := url.Parse(attr.Val)

				if err1 == nil && err2 == nil &&
					strings.ToLower(targetParsed.Host) == strings.ToLower(hrefParsed.Host) &&
					targetParsed.Host != "" {
					var buf bytes.Buffer
					if err := html.Render(&buf, n); err != nil {
						return "无法渲染 HTML", true, false
					}
					return buf.String(), true, false
				}

				// 如果提供了旧URL，也检查旧URL
				if oldURL != "" {
					oldParsed, err3 := url.Parse(oldURL)
					if err3 == nil &&
						strings.ToLower(oldParsed.Host) == strings.ToLower(hrefParsed.Host) &&
						oldParsed.Host != "" {
						var buf bytes.Buffer
						if err := html.Render(&buf, n); err != nil {
							return "无法渲染 HTML", true, true
						}
						return buf.String(), true, true
					}
				}
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if snippet, found, isOld := findBacklink(c, targetURL, oldURL); found {
			return snippet, true, isOld
		}
	}
	return "", false, false
}

// 新增辅助函数：查找友链页面的 URL
func findFriendsPageURL(n *html.Node, baseURL *url.URL) (string, bool) {
	if n.Type == html.ElementNode && n.Data == "a" {
		linkText := strings.ToLower(getNodeText(n))
		for _, keyword := range friendPageKeywords {
			if strings.Contains(linkText, keyword) {
				for _, attr := range n.Attr {
					if attr.Key == "href" {
						// 将相对 URL 解析为绝对 URL
						relURL, err := url.Parse(attr.Val)
						if err != nil {
							continue // 忽略无效的 href
						}
						absURL := baseURL.ResolveReference(relURL)
						return absURL.String(), true
					}
				}
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if url, found := findFriendsPageURL(c, baseURL); found {
			return url, true
		}
	}
	return "", false
}

// checkFallbackURL 检查 fallback URL 是否可访问
func checkFallbackURL(client *http.Client, url string) bool {
	resp, err := client.Head(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode >= 200 && resp.StatusCode < 400
}

// 新增辅助函数：获取一个节点下所有的文本内容
func getNodeText(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var text string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		text += getNodeText(c)
	}
	return text
}

// generateMarkdownReport 生成 Markdown 格式的报告
func generateMarkdownReport(results []CheckResult, myBlogURL string) {
	filename := fmt.Sprintf("backlink_report.md")
	file, err := os.Create(filename)
	if err != nil {
		log.Printf("无法创建 Markdown 文件: %v", err)
		return
	}
	defer file.Close()

	fmt.Fprintf(file, "# 友链检测报告\n\n")
	fmt.Fprintf(file, "**检测目标:** %s  \n", myBlogURL)
	fmt.Fprintf(file, "**生成时间:** %s  \n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(file, "**总链接数:** %d\n\n", len(results))

	fmt.Fprintf(file, "## 检测结果\n\n")
	fmt.Fprintf(file, "| Name | URL | Status | Backlink | Details |\n")
	fmt.Fprintf(file, "|------|-----|--------|----------|----------|\n")

	for _, r := range results {
		var statusStr string
		switch r.Status {
		case "Online":
			statusStr = "✅ Online"
		case "Offline":
			statusStr = "❌ Offline"
		default:
			statusStr = "⚠️ " + r.Status
		}

		var backlinkStr string
		if r.BacklinkFound {
			backlinkStr = "✅ Found"
		} else {
			backlinkStr = "❌ Not Found"
		}

		var details string
		if r.BacklinkFound {
			location := r.BacklinkLocation
			if location == "Homepage" {
				location = "首页"
			}
			snippet := strings.ReplaceAll(r.HTMLSnippet, "\n", " ")
			snippet = strings.ReplaceAll(snippet, "|", "\\|") // 转义管道符
			if len(snippet) > 60 {
				snippet = snippet[:60] + "..."
			}
			details = fmt.Sprintf("位置: %s<br>代码: `%s`", location, snippet)
		} else {
			if r.Error != "" {
				details = strings.ReplaceAll(r.Error, "|", "\\|") // 转义管道符
			} else {
				details = "N/A"
			}
		}

		// 转义 Markdown 特殊字符
		name := strings.ReplaceAll(r.Name, "|", "\\|")
		url := strings.ReplaceAll(r.URL, "|", "\\|")

		fmt.Fprintf(file, "| %s | %s | %s | %s | %s |\n",
			name, url, statusStr, backlinkStr, details)
	}

	fmt.Fprintf(file, "\n---\n")
	fmt.Fprintf(file, "*注意: 受检测方式所限，这些检测结果可能不准确，烦请手动前往目标站点核对。*\n")

	log.Printf("Markdown 报告已生成: %s", filename)
}

// printResults 更新以显示更详细的信息
func printResults(results []CheckResult, myBlogURL string) {
	table := tablewriter.NewWriter(os.Stdout)
	// 修改表头
	table.Header([]string{"Name", "URL", "Status", "Backlink", "Details"})

	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	for _, r := range results {
		var statusStr string
		switch r.Status {
		case "Online":
			statusStr = green(r.Status)
		case "Offline":
			statusStr = red(r.Status)
		default:
			statusStr = yellow(r.Status)
		}

		var backlinkStr string
		if r.BacklinkFound {
			backlinkStr = green("Found")
		} else {
			backlinkStr = red("Not Found")
		}

		var details string
		if r.BacklinkFound {
			location := r.BacklinkLocation
			if len(location) > 40 {
				location = "..." + location[len(location)-37:]
			}
			snippet := strings.ReplaceAll(r.HTMLSnippet, "\n", " ")
			if len(snippet) > 60 {
				snippet = snippet[:60] + "..."
			}
			details = fmt.Sprintf("On: %s\nSnippet: %s", location, snippet)
		} else {
			if r.Error != "" {
				details = r.Error
			} else {
				details = "N/A"
			}
		}

		row := []string{r.Name, r.URL, statusStr, backlinkStr, details}
		table.Append(row)
	}

	fmt.Printf("\n正在查找指向 '%s' 的返回链接...\n\n", myBlogURL)
	table.Render()
	// 生成 Markdown 文件
	generateMarkdownReport(results, myBlogURL)
	fmt.Printf("\n受检测方式所限，这些检测结果可能不准确，烦请手动前往目标站点核对。\n")

	// 清理浏览器实例
	log.Printf("清理浏览器实例...")
	if browserInstance != nil {
		err := browserInstance.Close()
		if err != nil {
			return
		}
	}
}
